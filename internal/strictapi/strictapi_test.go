package strictapi

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestAllowedEndpoints(t *testing.T) {
	var aa AllowedEndpoints
	aa.Add("/", http.MethodGet)
	aa.Add("/products", http.MethodGet)
	aa.Add("/products/%s", http.MethodGet, http.MethodPost, http.MethodDelete)
	aa.Add("/products/%s/images", http.MethodGet)
	aa.Add("/products/%s/reviews", http.MethodGet)
	aa.Add("/products/%s/reviews/%s", http.MethodGet, http.MethodDelete)

	testCases := []struct {
		method string
		path   string
		want   bool
	}{
		{
			method: "GET",
			path:   "/",
			want:   true,
		},
		{
			method: "POST",
			path:   "/",
			want:   false,
		},
		{
			method: "GET",
			path:   "/products",
			want:   true,
		},
		{
			method: "DELETE",
			path:   "/products",
			want:   false,
		},
		{
			method: "GET",
			path:   "/products/foo/reviews",
			want:   true,
		},
		{
			method: "DELETE",
			path:   "/products/foo/reviews",
			want:   false,
		},
		{
			method: "DELETE",
			path:   "/products/foo/reviews/bar",
			want:   true,
		},
		{
			method: "DELETE",
			path:   "/products/foo/reviews/bar/invalid",
			want:   false,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.method, tc.path), func(t *testing.T) {
			safe := aa.Check(&http.Request{
				Method: tc.method,
				URL: &url.URL{
					Path: tc.path,
				},
			})
			if safe != tc.want {
				t.Errorf("wanted Check(request) = %v, got %v instead", tc.want, safe)
			}
		})
	}
}

func TestAllowedEndpointsInvalid(t *testing.T) {
	var aa AllowedEndpoints
	defer func() {
		r := recover()
		want := "unsupported HTTP method: xyz"
		if r == nil || r.(string) != want {
			t.Errorf("expected panic not met: %v, got %v instead", want, r)
		}
	}()
	aa.Add("/", http.MethodGet, "xyz") // This is going to panic because we expect uppercase HTTP methods only.

}
