package frontend

import (
	"net/http"
)

// HomepageHandler handles the / page.
type HomepageHandler struct {
	Frontend *Frontend
}

func (h *HomepageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := &HTMLResponse{
		Template: "homepage",
		Title:    "homepage",
		Breadcrumb: []Breadcrumb{
			{Text: "Computers", Link: "/c/computers"},
			{Text: "Displays", Link: "/c/computers-displays"},
			{Text: "4K", Link: "/c/computers-displays-4k"},
			{Text: "LG Ultrafine 24\" 4K", Active: true},
		},
	}
	h.Frontend.Respond(w, r, resp)
}
