package frontend

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
	Frontend *Frontend
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/admin/") {
		u := services.UserFromRequest(r)
		if u.Access != services.AdminAuthorization {
			h.Frontend.HTTPError(w, r, http.StatusNotFound)
			return
		}
	}
	modules := h.Frontend.Modules
	path := filepath.Join(modules.Settings.StaticDirectory, r.URL.Path)
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
