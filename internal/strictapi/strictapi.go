package strictapi

import (
	"net/http"
	"strings"
)

// AllowedEndpoints that can be reached by the user.
// This intercepts requests before they reach out the handlers.
// This is not a router, but a safety net.
type AllowedEndpoints struct {
	endpoints []endpoint
}

// Check if processing the request is safe.
func (e *AllowedEndpoints) Check(r *http.Request) bool {
	for _, endpoint := range e.endpoints {
		if e.matchRoute(r, &endpoint) {
			return true
		}
	}
	return false
}

// Add endpoints.
// Uses %s for parameters (that must always be on their own).
//
// Must be called during initialization (not concurrency safe).
func (e *AllowedEndpoints) Add(path string, methods ...string) {
	for _, method := range methods {
		switch method {
		case http.MethodConnect, http.MethodDelete, http.MethodGet,
			http.MethodHead, http.MethodOptions, http.MethodPatch,
			http.MethodPost, http.MethodPut, http.MethodTrace:
		default:
			panic("unsupported HTTP method: " + method)
		}
	}

	e.endpoints = append(e.endpoints, endpoint{
		path:    path,
		methods: methods,
	})
}

func (e *AllowedEndpoints) matchRoute(r *http.Request, endpoint *endpoint) bool {
	route := strings.FieldsFunc(endpoint.path, isPathSeparator)
	visited := strings.FieldsFunc(r.URL.Path, isPathSeparator)

	// The API endpoint URIs follow a rigid structure and can only match if the pattern matches exactly.
	if len(route) != len(visited) {
		return false
	}
	for x := 0; x < len(route); x++ {
		if route[x] != "%s" && route[x] != visited[x] {
			return false
		}
	}

	// Check if method is allowed.
	for _, method := range endpoint.methods {
		if method == r.Method {
			return true
		}
	}
	return false
}

func isPathSeparator(r rune) bool {
	return r == '/'
}

// endpoint is a route for a given whitelisted pattern and its allowed methods.
type endpoint struct {
	methods []string
	path    string
}
