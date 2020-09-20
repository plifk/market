package router

import (
	"context"
	"net/http"
	"path"
	"strings"
)

// Mux is a set of routes.
//
// NOTE(henvic): This could be a tree router too, but this should be good and simple enough.
type Mux struct {
	DefaultHandler http.Handler
	Routes         []Route
}

// Validate router mux entries.
func (m *Mux) Validate() {
	for _, route := range m.Routes {
		for _, method := range route.Methods {
			switch method {
			case http.MethodConnect, http.MethodDelete, http.MethodGet,
				http.MethodHead, http.MethodOptions, http.MethodPatch,
				http.MethodPost, http.MethodPut, http.MethodTrace:
			default:
				panic("unsupported HTTP method: " + method)
			}
		}
	}
}

// ServeHTTP handles request using the mux router.
// URL paths are normalized by stripping trailing slash.
// If a match is not found, it uses the m.DefaultHandler or http.NotFound if a default handler is not set.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if normalizeRedirect(w, r) {
		return
	}
	for _, route := range m.Routes {
		if params, ok := route.Match(r.Method, r.URL.Path); ok {
			if len(params) != 0 {
				r = r.Clone(WithParams(r.Context(), params))
			}
			route.Handler.ServeHTTP(w, r)
			return
		}
	}
	if m.DefaultHandler != nil {
		m.DefaultHandler.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

// Route for the HTTP server. Pattern uses a :field format.
// Pattern example: /products/:product_id/details
// Method cannot be empty.
type Route struct {
	Pattern string
	Methods []string
	Handler http.Handler

	_ struct{}
}

// Match route.
func (r *Route) Match(method, path string) (Params, bool) {
	for _, m := range r.Methods {
		if method == m {
			return r.MatchPath(path)
		}
	}
	return Params{}, false
}

// MatchPath with route.
func (r *Route) MatchPath(path string) (Params, bool) {
	rp := strings.FieldsFunc(string(r.Pattern), isPathSeparator)
	pp := strings.FieldsFunc(path, isPathSeparator)
	params := Params{}
	if len(rp) != len(pp) {
		return params, false
	}
	for pos, p := range rp {
		if len(p) > 0 && p[0] == ':' {
			if _, ok := params[p[1:]]; ok {
				// NOTE(henvic): This panic helps when debugging.
				panic("route contains duplicated field")
			}
			params[p[1:]] = pp[pos]
		} else if p != pp[pos] {
			return params, false
		}
	}
	return params, true
}

func isPathSeparator(r rune) bool {
	return r == '/'
}

// WithParams can be used to annotate the context of a request with route params.
func WithParams(ctx context.Context, params Params) context.Context {
	return context.WithValue(ctx, contextParams{}, params)
}

// ReadParams from context.
func ReadParams(ctx context.Context) Params {
	if p, ok := ctx.Value(contextParams{}).(Params); ok {
		return p
	}
	return Params{}
}

type contextParams struct{}

// Params for routing URL paths.
type Params map[string]string

// Get param field value.
func (p Params) Get(field string) string {
	return p[field]
}

// normalizeRedirect redicts if path needs to be normalized. It returns true if redirection is made.
func normalizeRedirect(w http.ResponseWriter, r *http.Request) (redirected bool) {
	path := cleanPath(r.URL.Path)

	// Remove trailing slash, unless if path = /.
	if path != "/" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if path == r.URL.Path {
		return false
	}
	rc := *r.URL
	rc.User = r.URL.User
	rc.Path = path
	// Use 'moved permanently' for HEAD|GET requests.
	code := http.StatusPermanentRedirect
	if r.Method != http.MethodHead && r.Method != http.MethodGet {
		code = http.StatusTemporaryRedirect
	}
	http.Redirect(w, r, rc.String(), code)
	return true
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
// Copyright 2009 The Go Authors. All rights reserved.
// Source: https://github.com/golang/go/blob/b8fd3cab3944d5dd5f2a50f3cc131b1048897ee1/src/net/http/http.go
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}
