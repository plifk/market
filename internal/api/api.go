package api

import (
	"fmt"
	"net/http"

	"github.com/plifk/market/internal/services"
)

// Router for the API.
type Router struct {
	modules *services.Modules
}

// Load API.
func (rh *Router) Load(modules *services.Modules) {
	rh.modules = modules
}

func (rh *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "API")
}
