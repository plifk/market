package frontend

import (
	"log"
	"net/http"
	"strings"

	"github.com/plifk/market/internal/services"
)

// Frontend helpers.
type Frontend struct {
	Modules       *services.Modules
	staticHandler http.Handler
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
	Frontend *Frontend

	homepageHandler *HomepageHandler
	loginHandler    *LoginHandler
	logoutHandler   *LogoutHandler
	staticHandler   *StaticHandler
	searchHandler   *SearchHandler
	productHandler  *ProductHandler
	accountHandler  *AccountHandler
	adminHandler    *AdminHandler
}

// Load HTTP handlers.
func (rh *Router) Load(modules *services.Modules) {
	frontend := &Frontend{
		Modules: modules,
	}
	rh.staticHandler = &StaticHandler{
		Frontend: frontend,
	}
	rh.Frontend = frontend
	rh.Frontend.staticHandler = rh.staticHandler
	rh.homepageHandler = &HomepageHandler{Frontend: frontend}
	rh.loginHandler = &LoginHandler{Frontend: frontend}
	rh.logoutHandler = &LogoutHandler{Frontend: frontend}
	rh.searchHandler = &SearchHandler{Frontend: frontend}
	rh.productHandler = &ProductHandler{Frontend: frontend}
	rh.accountHandler = &AccountHandler{Frontend: frontend}
	rh.adminHandler = &AdminHandler{Frontend: frontend}
	rh.adminHandler.Load()
}

func (rh *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	modules := rh.Frontend.Modules
	session, err := modules.Sessions.Read(w, r)
	if err != nil {
		log.Printf("request %s failed to get session: %v\n", r.Header.Get("X-Request-ID"), err)
	}
	if session != nil {
		ctx := services.SessionContext(r.Context(), session)
		if session.UserID != "" {
			// TODO(henvic): Add caching layer.
			u, err := modules.Accounts.GetUserByID(r.Context(), session.UserID)
			if err != nil {
				log.Printf("request %s failed to get user: %v\n", r.Header.Get("X-Request-ID"), err)
			}
			ctx = services.UserContext(ctx, u)
		}
		r = r.Clone(ctx)
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
		handler = rh.homepageHandler
	case route.is("/s"):
		handler = rh.searchHandler
	case strings.HasPrefix(path, "/p/"):
		handler = rh.productHandler
	case route.is("/account"):
		handler = rh.accountHandler
	case route.is("/admin"):
		handler = rh.adminHandler
	case route.is("/login"):
		handler = rh.loginHandler
	case route.is("/logout"):
		handler = rh.logoutHandler
	}
	if handler == nil {
		handler = rh.staticHandler
	}
	handler.ServeHTTP(w, r)
}

// dirRouter makes it less repetitive to use a directory-style routing style.
type dirRouter string

func (p dirRouter) is(route string) bool {
	return route == string(p) || route == string(p+"/")
}

func (p dirRouter) within(route string) bool {
	return route == string(p) || strings.HasPrefix(string(p), string(route)+"/")
}
