// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
package main

import (
	"net/http"

	"code.minty.io/mintyproxy/proxy"
)

func main() {
	p := proxy.New()

	rp := p.ForPattern("/", proxy.Directors(
		// Proxy http, or http and https.
		//proxy.ToPort(8777),
		proxy.ToPorts(8777, 8778),
	))

	// http
	//p.ForPattern("/", proxy.ToPort(8777))
	// https
	rp := p.ForPattern("/", proxy.ToPort(8778))
	// Ignore SSL validation when proxying.
	rp.Transport = &http.Transport{Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}


	p.ForHost("localhost", proxy.ToHost("crackerjack.local"))

	go http.ListenAndServe(":5555", p)
	http.ListenAndServeTLS(":5556", "/Volumes/certs/minty/minty.crt", "/Volumes/certs/minty/minty.key", p)
}
*/

package roxy
