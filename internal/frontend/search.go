package frontend

import (
	"net/http"
)

// SearchHandler for the application.
type SearchHandler struct {
	Frontend *Frontend
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := &HTMLResponse{
		Template: "search",
		Title:    "Search",
	}
	h.Frontend.Respond(w, r, resp)
}
