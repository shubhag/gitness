// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package address

import (
	"net/http"
	"strings"
)

// Handler returns an http.HandlerFunc middleware that sets
// the http.Request scheme and hostname.
func Handler(scheme, host string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// update the scheme and host for the inbound
			// http.Request so they are available to subsequent
			// handlers in the chain.
			r.URL.Scheme = scheme
			r.URL.Host = host

			// if the scheme is not configured, attempt to ascertain
			// the scheme from the inbound http.Request.
			if r.URL.Scheme == "" {
				r.URL.Scheme = resolveScheme(r)
			}

			// if the host is not configured, attempt to ascertain
			// the host from the inbound http.Request.
			if r.URL.Host == "" {
				r.URL.Host = resolveHost(r)
			}

			// invoke the next handler in the chain.
			next.ServeHTTP(w, r)
		})
	}
}

// resolveScheme is a helper function that evaluates the http.Request
// and returns the scheme, HTTP or HTTPS. It is able to detect,
// using the X-Forwarded-Proto, if the original request was HTTPS
// and routed through a reverse proxy with SSL termination.
func resolveScheme(r *http.Request) string {
	const https = "https"
	switch {
	case r.URL.Scheme == https:
		return https
	case r.TLS != nil:
		return https
	case strings.HasPrefix(r.Proto, "HTTPS"):
		return https
	case r.Header.Get("X-Forwarded-Proto") == https:
		return https
	default:
		return "http"
	}
}

// resolveHost is a helper function that evaluates the http.Request
// and returns the hostname. It is able to detect, using the
// X-Forarded-For header, the original hostname when routed
// through a reverse proxy.
func resolveHost(r *http.Request) string {
	switch {
	case len(r.Host) != 0:
		return r.Host
	case len(r.URL.Host) != 0:
		return r.URL.Host
	case len(r.Header.Get("X-Forwarded-For")) != 0:
		return r.Header.Get("X-Forwarded-For")
	case len(r.Header.Get("X-Host")) != 0:
		return r.Header.Get("X-Host")
	case len(r.Header.Get("XFF")) != 0:
		return r.Header.Get("XFF")
	case len(r.Header.Get("X-Real-IP")) != 0:
		return r.Header.Get("X-Real-IP")
	default:
		return "localhost:3000"
	}
}
