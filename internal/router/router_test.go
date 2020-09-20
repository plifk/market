package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestRouteMatch(t *testing.T) {
	t.Parallel()
	r := Route{
		Pattern: "/products/:product_id",
		Methods: []string{http.MethodGet, http.MethodPost},
	}

	testCases := []struct {
		method string
		path   string

		ok          bool
		wantProduct string
	}{
		{method: http.MethodGet, path: "", ok: false},
		{method: http.MethodGet, path: "/foo", ok: false},
		{method: http.MethodGet, path: "/foo/abc", ok: false},
		{method: http.MethodGet, path: "/products", ok: false},
		{method: http.MethodGet, path: "/products/abc", ok: true, wantProduct: "abc"},
		{method: http.MethodPost, path: "/products/abc", ok: true, wantProduct: "abc"},
		{method: http.MethodDelete, path: "/products/abc", ok: false},
		{method: http.MethodGet, path: "/products/abc/def", ok: false},
	}
	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			params, got := r.Match(tc.method, tc.path)
			if got != tc.ok {
				t.Errorf("route mismatch: %s", tc.path)
			}
			if tc.ok {
				gotProduct := params.Get("product_id")
				if gotProduct != tc.wantProduct {
					t.Errorf("wanted product %s, got %v instead (for path %s)", tc.wantProduct, gotProduct, tc.path)
				}
			}
		})
	}
}

func TestRouteMatchMultipleFields(t *testing.T) {
	t.Parallel()
	r := Route{
		Pattern: "/foo/:x/bar/:y",
		Methods: []string{http.MethodGet},
	}
	params, ok := r.Match(http.MethodGet, "/foo/abc/bar/def")
	if !ok {
		t.Errorf("route failed: %q", r)
	}
	want := Params{
		"x": "abc",
		"y": "def",
	}
	if !reflect.DeepEqual(params, want) {
		t.Errorf("wanted params to be %+v, got %+v instead", want, params)
	}
	if want, got := "abc", params.Get("x"); got != want {
		t.Errorf("wanted param x to be %q, got %v instead", want, got)
	}
	if want, got := "def", params.Get("y"); got != want {
		t.Errorf("wanted param y to be %q, got %v instead", want, got)
	}

	ctxWithParams := WithParams(context.Background(), params)
	paramsFromContext := ReadParams(ctxWithParams)
	if !reflect.DeepEqual(params, paramsFromContext) {
		t.Errorf("wanted params from context to be the same as original")
	}
}

func TestRouteMatchDuplicatedFields(t *testing.T) {
	t.Parallel()
	r := Route{
		Methods: []string{http.MethodGet},
		Pattern: "/foo/:x/bar/:x",
	}
	defer func() {
		want := "route contains duplicated field"
		if r := recover(); r != want {
			t.Errorf("missing expected panic, got %q instead of %q", r, want)
		}
	}()
	r.Match(http.MethodGet, "/foo/abc/bar/def")
}

func TestReadMissingParams(t *testing.T) {
	t.Parallel()
	if got := ReadParams(context.Background()); got == nil {
		t.Error("params must not be nil")
	}
}

func TestMux(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("expected mux to be valid, got panic %v", r)
		}
	}()

	mux := &Mux{
		Routes: []Route{
			{
				Pattern: "/valid",
				Methods: []string{http.MethodGet},
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if ReadParams(r.Context()) == nil {
						t.Error("unexpected nil value returned by ReadParams")
					}
					if params := ReadParams(r.Context()); len(params) != 0 {
						t.Errorf("unexpected values returned by ReadParams: %+v", params)
					}
				}),
			},
			{
				Pattern: "/products/:id",
				Methods: []string{http.MethodGet},
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusAccepted)
					params := ReadParams(r.Context())
					if want, got := 1, len(params); want != got {
						t.Errorf("expected %d values to be returned by ReadParams, got %d instead", want, got)
					}
					if want, got := "xyz", params["id"]; want != got {
						t.Errorf("expected id param to be %s, got %s instead", want, got)
					}
				}),
			},
		},
	}
	mux.Validate()

	testCases := []struct {
		link       string
		statusCode int
	}{
		{link: "http://example.com/valid", statusCode: http.StatusOK},
		{link: "http://example.com/valid/", statusCode: http.StatusPermanentRedirect},
		{link: "http://example.com/products/xyz", statusCode: http.StatusAccepted},
		{link: "http://example.com/not-found", statusCode: http.StatusNotFound},
	}
	for _, tc := range testCases {
		t.Run(tc.link, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.link, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if want, got := tc.statusCode, w.Result().StatusCode; want != got {
				t.Errorf("expected status code to be %d, got %d instead", want, got)
			}
		})
	}
}

func TestMuxDefaultHandler(t *testing.T) {
	mux := &Mux{
		DefaultHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		Routes: []Route{
			{
				Pattern: "/do-not-call-me",
				Methods: []string{http.MethodGet},
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fatal("unexpected call to route handler")
				}),
			},
		},
	}
	mux.Validate()

	req := httptest.NewRequest("GET", "http://example.com/no-match", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Check if we are calling http.Error or the defined default handler.
	if want, got := http.StatusCreated, w.Result().StatusCode; want != got {
		t.Errorf("expected status code to be %d, got %d instead", want, got)
	}
}

func TestMuxInvalid(t *testing.T) {
	t.Parallel()
	defer func() {
		want := "unsupported HTTP method: invalid-method"
		if r := recover(); r != want {
			t.Errorf("expected mux panicking %v, got %v instead", want, r)
		}
	}()
	mux := &Mux{
		Routes: []Route{
			{
				Pattern: "/valid",
				Methods: []string{http.MethodGet},
			},
			{
				Pattern: "/foo",
				Methods: []string{"invalid-method"},
			},
		},
	}
	mux.Validate()
}

func TestNormalizeRedirect(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc    string
		request *http.Request

		redirected bool
		location   string
		statusCode int
	}{
		{
			desc: "GET (invalid: empty)",
			request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path: "",
				},
			},

			redirected: true,
			location:   "/",
			statusCode: http.StatusPermanentRedirect,
		},
		{
			desc: "GET (invalid: prefix must be /)",
			request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path: "abc",
				},
			},

			redirected: true,
			location:   "/abc",
			statusCode: http.StatusPermanentRedirect,
		},
		{
			desc: "GET /",
			request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path: "/",
				},
			},

			redirected: false,
		},
		{
			desc: "GET /foo/x",
			request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path: "/foo/x",
				},
			},

			redirected: false,
		},
		{
			desc: "GET /foo/x/",
			request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path: "/foo/x/",
				},
			},

			redirected: true,
			location:   "/foo/x",
			statusCode: http.StatusPermanentRedirect,
		},
		{
			desc: "GET /bar/x/",
			request: &http.Request{
				Method: http.MethodPut,
				URL: &url.URL{
					Path: "/bar/x/",
				},
			},

			redirected: true,
			location:   "/bar/x",
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			desc: "GET /bar//x/",
			request: &http.Request{
				Method: http.MethodPut,
				URL: &url.URL{
					Path: "/bar//x/",
				},
			},

			redirected: true,
			location:   "/bar/x",
			statusCode: http.StatusTemporaryRedirect,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			rec := httptest.NewRecorder()

			if got := normalizeRedirect(rec, tc.request); got != tc.redirected {
				t.Errorf("expected redirected = %v, got %v instead", tc.redirected, got)
			}

			if !tc.redirected {
				return
			}
			res := rec.Result()
			location, err := res.Location()
			if err != nil {
				t.Errorf("unexpected error getting location: %v", err)
			}
			if location.String() != tc.location {
				t.Errorf("expected location to be %s, got %s instead", location.String(), tc.location)
			}
			if res.StatusCode != tc.statusCode {
				t.Errorf("expected redirected status code = %v, got %v instead", tc.statusCode, rec.Result().StatusCode)
			}
		})
	}
}
