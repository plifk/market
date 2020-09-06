package router

import (
	"context"
	"reflect"
	"testing"
)

func TestRouteMatch(t *testing.T) {
	r := Route("/products/:product_id")

	testCases := []struct {
		path string

		ok      bool
		product string
	}{
		{path: "", ok: false},
		{path: "/foo", ok: false},
		{path: "/foo/abc", ok: false},
		{path: "/products", ok: false},
		{path: "/products/abc", ok: true, product: "abc"},
		{path: "/products/abc/def", ok: false},
	}
	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			params, got := r.Match(tc.path)
			if got != tc.ok {
				t.Errorf("route mismatch: %s", tc.path)
			}
			if tc.ok {
				gotProduct := params.Get("product_id")
				if gotProduct != tc.product {
					t.Errorf("wanted product %s, got %v instead (for path %s)", tc.product, gotProduct, tc.path)
				}
			}
		})
	}
}

func TestRouteMatchMultipleFields(t *testing.T) {
	r := Route("/foo/:x/bar/:y")
	params, ok := r.Match("/foo/abc/bar/def")
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
	r := Route("/foo/:x/bar/:x")
	defer func() {
		want := "route contains duplicated field"
		if r := recover(); r != want {
			t.Errorf("missing expected panic, got %q instead of %q", r, want)
		}
	}()
	r.Match("/foo/abc/bar/def")
}

func TestReadMissingParams(t *testing.T) {
	if got := ReadParams(context.Background()); got == nil {
		t.Error("params must not be nil")
	}
}
