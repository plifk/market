package webpages

import (
	"path"
	"strings"
)

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

// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
// Source: https://github.com/golang/go/blob/187a41dbf730117bd52f871009466a9679d6b718/src/net/http/example_filesystem_test.go
func containsDotFile(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}
