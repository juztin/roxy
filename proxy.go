// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package roxy is a simple web proxy that can proxy
// by host or pattern.
package roxy

import (
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"

	"github.com/juztin/marbles/mux"
)

// Type Director receives an http.Request that
// will be used as a proxy. It updates it before
// being passed to the proxied handler.
type Director func(req *http.Request)

// Type Proxy implements an http.Handler
// that proxies http requests
type Proxy struct {
	mu       sync.RWMutex
	hosts    map[string]*httputil.ReverseProxy
	mux      *mux.ServeMux
	ProxyTLS bool
}

// Type Directors chains all of the given directors
// into a single director.
func Directors(directors ...Director) Director {
	return func(req *http.Request) {
		for i := range directors {
			directors[i](req)
		}
	}
}

// hostFrom returns the host portion without the colon + port.
func hostFrom(host string) string {
	if i := strings.Index(host, ":"); i > 0 {
		return host[:i]
	}
	return host
}

// ToHost returns a directory that proxies
// to the given scheme.
func ToScheme(scheme string) Director {
	return func(req *http.Request) {
		req.URL.Scheme = scheme
	}
}

// ToHost returns a directory that proxies
// to the given host.
func ToHost(host string) Director {
	return func(req *http.Request) {
		req.URL.Host = host
	}
}

// ToPort returns a director that proxies
// to the given port.
func ToPort(port int) Director {
	p := ":" + strconv.Itoa(port)

	return func(req *http.Request) {
		req.URL.Host = hostFrom(req.URL.Host) + p
	}
}

// ToPorts returns a director that proxies to the
// given httpPort, when the incoming request is
// non-TLS, and to the httpsPort, when the
// incoming request is TLS.
func ToPorts(httpPort, httpsPort int) Director {
	p0 := ":" + strconv.Itoa(httpPort)
	p1 := ":" + strconv.Itoa(httpsPort)

	return func(req *http.Request) {
		port := p0
		if req.TLS != nil {
			port = p1
		}
		if i := strings.Index(req.URL.Host, ":"); i > 0 {
			req.URL.Host = req.URL.Host[:i] + port
		} else {
			req.URL.Host = req.URL.Host + port
		}
	}
}

// ForHost registers the director for the given host.
func (p *Proxy) ForHost(host string, d Director) *httputil.ReverseProxy {
	p.mu.Lock()
	defer p.mu.Unlock()
	rp, ok := p.hosts[host]
	if !ok {
		rp = &httputil.ReverseProxy{Director: d}
		p.hosts[host] = rp
	}
	return rp
}

// RemoveHost removes the director for the given host.
func (p *Proxy) RemoveHost(host string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.hosts, host)
}

// ClearHosts clears out all of the directors
// currently registered.
func (p *Proxy) ClearHosts() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for host := range p.hosts {
		delete(p.hosts, host)
	}
}

// ForPattern registers the pattern for the given pattern.
func (p *Proxy) ForPattern(pattern string, d Director) *httputil.ReverseProxy {
	rp := &httputil.ReverseProxy{Director: d}
	p.mux.HandleFunc(pattern, rp.ServeHTTP)
	return rp
}

// RemovePattern removes the handler registerd
// under the given pattern.
func (p *Proxy) RemovePattern(pattern string) {
	p.mux.UnHandle(pattern)
}

// ClearPatterns clears out all handlers
// currently registered.
func (p *Proxy) ClearPatterns() {
	p.mux.Clear()
}

// Handle performs the same functionality as http.Handle
func (p *Proxy) Handle(pattern string, fn http.Handler) {
	p.mux.Handle(pattern, fn)
}

// HandleFunc performs the same functionality
// as http.HandlerFunc
func (p *Proxy) HandleFunc(pattern string, fn func(http.ResponseWriter, *http.Request)) {
	p.Handle(pattern, http.HandlerFunc(fn))
}

// http.Handler implementation
func (p Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.Host = req.Host
	req.URL.Scheme = "http"
	if p.ProxyTLS && req.TLS != nil {
		req.URL.Scheme = "https"
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	if rp, ok := p.hosts[hostFrom(req.Host)]; ok {
		rp.ServeHTTP(rw, req)
	} else {
		p.mux.ServeHTTP(rw, req)
	}
}

// New returns a new proxy.
func New() *Proxy {
	return &Proxy{
		hosts:    make(map[string]*httputil.ReverseProxy),
		mux:      mux.NewServeMux(),
		ProxyTLS: false,
	}
}
