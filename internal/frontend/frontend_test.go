package frontend

import (
	"testing"
)

func TestDirRouter(t *testing.T) {
	visited := "/p/N3oCS85HvpY"
	r := dirRouter(visited)

	if route := "/p"; r.is(route) {
		t.Errorf("unexpected r.is(%q) = true", route)
	}
	if route := "/p/"; r.is(route) {
		t.Errorf("unexpected r.is(%q) = true", route)
	}
	if route := "/p/N3oCS85HvpY/x"; r.is(route) {
		t.Errorf("unexpected r.is(%q) = true", route)
	}
	if route := "/p/N3oCS85HvpY"; !r.is(route) {
		t.Errorf("unexpected r.is(%q) = false", route)
	}
	if route := "/p/N3oCS85HvpY/"; !r.is(route) {
		t.Errorf("unexpected r.is(%q) = false", route)
	}

	if route := "/p"; !r.within(route) {
		t.Errorf("unexpected r.within(%q) = false", route)
	}
	if route := "/p/"; !r.within(route) {
		t.Errorf("unexpected r.within(%q) = false", route)
	}
	if route := "/p/N3oCS85HvpY/x"; r.within(route) {
		t.Errorf("unexpected r.within(%q) = true", route)
	}
	if route := "/p/N3oCS85HvpY"; !r.within(route) {
		t.Errorf("unexpected r.within(%q) = false", route)
	}
	if route := "/p/N3oCS85HvpY/"; !r.within(route) {
		t.Errorf("unexpected r.within(%q) = false", route)
	}
}
