package webpages

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/plifk/market/internal/services"
)

// StaticHandler checks permission to see if the user is able to access static content, and serve them.
// This is useful to provide files such as favicon.ico, or other stuff that aren't supposed to be stored in the CDN.
type StaticHandler struct {
	Modules  *services.Modules
	Frontend *Frontend
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/admin/") { // TODO(henvic): Verify if user has access.
		h.Frontend.HTTPError(w, r, http.StatusNotFound)
		return
	}
	path := filepath.Join(h.Modules.Settings.StaticDirectory, r.URL.Path)
	switch _, err := os.Stat(path); {
	case os.IsNotExist(err):
		h.Frontend.HTTPError(w, r, http.StatusNotFound)
		return
	case err != nil:
		log.Printf("cannot serve static page %q: %v\n", path, err)
		h.Frontend.HTTPError(w, r, http.StatusInternalServerError)
		return
	}

	if containsDotFile(r.URL.Path) {
		h.Frontend.HTTPError(w, r, http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, path)
}

// dirRouter makes it less repetitive to use a directory-style routing style.
type dirRouter string

func (p dirRouter) is(route string) bool {
	return route == string(p) || strings.HasPrefix(string(p), string(route)+"/")
}
