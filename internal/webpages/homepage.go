package webpages

import (
	"net/http"

	"github.com/plifk/market/internal/services"
)

// HomepageHandler handles the / page.
type HomepageHandler struct {
	Modules  *services.Modules
	Frontend *Frontend
}

func (h *HomepageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := &HTMLResponse{
		Template: "homepage",
		Title:    "homepage",
	}
	h.Frontend.Respond(w, r, resp)
}
