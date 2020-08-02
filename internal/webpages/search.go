package webpages

import (
	"net/http"

	"github.com/plifk/market/internal/services"
)

// SearchHandler for the application.
type SearchHandler struct {
	Modules  *services.Modules
	Frontend *Frontend
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := &HTMLResponse{
		Template: "search",
		Title:    "Search",
	}
	h.Frontend.Respond(w, r, resp)
}
