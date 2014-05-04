// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

type Director func(req *http.Request)

type mux struct {
	*http.ServeMux
	Director Director
}

type Proxy struct {
	hosts map[string]*httputil.ReverseProxy
	mux   *http.ServeMux
}

func Directors(directors ...Director) func(req *http.Request) {
	return func(req *http.Request) {
		for i := range directors {
			directors[i](req)
		}
	}
}

func hostFrom(req *http.Request) string {
	if i := strings.Index(req.Host, ":"); i > 0 {
		return req.Host[:i]
	}
	return req.Host
}

func ToScheme(scheme string) Director {
	return func(req *http.Request) {
		req.URL.Scheme = scheme
	}
}

func ToHost(host string) Director {
	return func(req *http.Request) {
		req.URL.Host = host
	}
}

func ToPort(port int) Director {
	p := ":" + strconv.Itoa(port)

	return func(req *http.Request) {
		if i := strings.Index(req.URL.Host, ":"); i > 0 {
			req.URL.Host = req.URL.Host[:i] + p
		} else {
			req.URL.Host = req.URL.Host + p
		}
	}
}

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

func (p *Proxy) ForHost(host string, d Director) *httputil.ReverseProxy {
	rp, ok := p.hosts[host]
	if !ok {
		rp = &httputil.ReverseProxy{Director: d}
		p.hosts[host] = rp
	}
	return rp
}

func (p *Proxy) ForPattern(pattern string, d Director) *httputil.ReverseProxy {
	rp := &httputil.ReverseProxy{Director: d}
	p.mux.HandleFunc(pattern, rp.ServeHTTP)
	return rp
}

func (p Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.Host = req.Host
	req.URL.Scheme = "http"
	if req.TLS != nil {
		req.URL.Scheme = "https"
	}

	if rp, ok := p.hosts[hostFrom(req)]; ok {
		rp.ServeHTTP(rw, req)
	} else {
		p.mux.ServeHTTP(rw, req)
	}
}

func New() *Proxy {
	return &Proxy{
		hosts: make(map[string]*httputil.ReverseProxy),
		mux:   http.NewServeMux(),
	}
}
