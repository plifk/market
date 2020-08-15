package frontend

import (
	"net/http"

	"github.com/plifk/market/internal/services"
)

// AccountHandler for the application.
type AccountHandler struct {
	Frontend *Frontend
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := services.UserFromRequest(r)
	if user == nil {
		h.Frontend.HTTPError(w, r, http.StatusNotFound)
		return
	}

	resp := &HTMLResponse{
		Template: "account",
		Title:    "Your Account",
	}
	h.Frontend.Respond(w, r, resp)
}
