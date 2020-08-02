package webpages

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/plifk/market/internal/services"
)

// Frontend helpers.
type Frontend struct {
	Modules *services.Modules
}

// HTTPErrorHandlerTemplate to use when rendering 'Internal Server Error' pages and the like.
type HTTPErrorHandlerTemplate struct {
	StatusCode int
	StatusText string
	Errors     []error
}

// HTTPError renders an user-friendly error page.
func (f *Frontend) HTTPError(w http.ResponseWriter, r *http.Request, code int, errs ...error) {
	resp := &HTMLResponse{
		Template: "http-error",
		Title:    http.StatusText(code),
		Content: HTTPErrorHandlerTemplate{
			StatusCode: code,
			StatusText: http.StatusText(code),
			Errors:     errs,
		},
	}
	w.WriteHeader(code)
	f.Respond(w, r, resp)
}

// Router for the webpages.
type Router struct {
	HomepageHandler *HomepageHandler
	LoginHandler    *LoginHandler
	LogoutHandler   *LogoutHandler
	StaticHandler   *StaticHandler
	SearchHandler   *SearchHandler

	modules *services.Modules
}

// Load HTTP handlers.
func (rh *Router) Load(modules *services.Modules) {
	rh.modules = modules
	frontend := &Frontend{
		Modules: modules,
	}

	rh.HomepageHandler = &HomepageHandler{
		Modules:  modules,
		Frontend: frontend,
	}
	rh.LoginHandler = &LoginHandler{
		Modules:  modules,
		Frontend: frontend,
	}
	rh.LogoutHandler = &LogoutHandler{
		Modules:  modules,
		Frontend: frontend,
	}
	rh.StaticHandler = &StaticHandler{
		Modules:  modules,
		Frontend: frontend,
	}
	rh.SearchHandler = &SearchHandler{
		Modules:  modules,
		Frontend: frontend,
	}
}

func (rh *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := rh.modules.Sessions.Read(w, r)
	if err != nil {
		log.Printf("request %s failed to get session: %v\n", r.Header.Get("X-Request-ID"), err)
	}
	if session != nil {
		r = r.Clone(services.SessionContext(r.Context(), session))
	}

	path := cleanPath(r.URL.Path)
	if r.URL.Path != path {
		url := *r.URL
		url.Path = path
		http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
		return
	}

	var handler http.Handler
	switch route := dirRouter(path); {
	case path == "/":
		handler = rh.HomepageHandler
	case path == "/s" || strings.HasPrefix(path, "/s/"):
		handler = rh.SearchHandler
	case strings.HasPrefix(path, "/p/"):
		fmt.Fprintln(w, "product page")
	case route.is("/admin"):
	case route.is("/login"):
		handler = rh.LoginHandler
	case route.is("/logout"):
		handler = rh.LogoutHandler
	case route.is("/account"):
	}
	if handler == nil {
		handler = rh.StaticHandler
	}
	handler.ServeHTTP(w, r)
}
