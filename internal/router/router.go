package router

import (
	"context"
	"net/http"
	"path"
	"strings"
)

// Route format.
// Example: /products/:product_id/details
type Route string

// Match path with route.
func (r Route) Match(path string) (Params, bool) {
	rp := strings.FieldsFunc(string(r), isPathSeparator)
	pp := strings.FieldsFunc(path, isPathSeparator)
	if len(rp) != len(pp) {
		return nil, false
	}
	params := Params{}
	for pos, p := range rp {
		if len(p) > 0 && p[0] == ':' {
			if _, ok := params[p[1:]]; ok {
				// NOTE(henvic): This panic is here to help debugging the application.
				// It has a minor limitation: it is only triggered when the format matches.
				panic("route contains duplicated field")
			}
			params[p[1:]] = pp[pos]
		} else if p != pp[pos] {
			return nil, false
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

// NormalizeRedirect redicts if path needs to be normalized. It returns true if redirection is made.
func NormalizeRedirect(w http.ResponseWriter, r *http.Request) (cancel bool) {
	path := cleanPath(r.URL.Path)
	if r.URL.Path != path {
		rc := *r.URL
		rc.User = r.URL.User
		rc.Path = path
		// Use 'moved permanently' for HEAD|GET requests.
		code := http.StatusMovedPermanently
		if r.Method != http.MethodHead && r.Method != http.MethodGet {
			code = http.StatusTemporaryRedirect
		}
		http.Redirect(w, r, rc.String(), code)
		return true
	}
	return false
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
