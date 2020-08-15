package frontend

import (
	"net/http"
)

// ProductHandler for the application.
type ProductHandler struct {
	Frontend *Frontend
}

func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := &HTMLResponse{
		Template: "product",
		Title:    "Product",
	}
	h.Frontend.Respond(w, r, resp)
}
