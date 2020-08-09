package market

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/henvic/httpretty"
	"github.com/plifk/market/internal/api"
	"github.com/plifk/market/internal/webpages"
)

// ServerHTTP handles HTTP requests to the market system.
// It serves API, admin, and web pages based on the request Host address.
func (s *System) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Request-ID") == "" {
		r.Header.Set("X-Request-ID", uuid.New().String())
	}
	var h http.Handler

	switch {
	case strings.HasPrefix(r.Host, "api."):
		h = s.api
	case strings.HasPrefix(r.Host, "www."):
		h = s.webpages
	default:
		h = s.notFound
	}
	h.ServeHTTP(w, r)
}

func (s *System) httpHandlers() {
	settings := s.core.Settings
	s.httpServer = &http.Server{
		Addr:    settings.HTTPAddress,
		Handler: s.core.CSRFProtection, // This CSRF middleware injects the HTTP entrypoint.
	}
	if settings.Debug {
		s.httpServer.Handler = httpLogger().Middleware(s.httpServer.Handler)
	}

	s.api = &api.Router{}
	s.api.Load(s.Modules)

	s.webpages = &webpages.Router{}
	s.webpages.Load(s.Modules)
}

func httpLogger() *httpretty.Logger {
	l := &httpretty.Logger{
		Colors:     true,
		Formatters: []httpretty.Formatter{&httpretty.JSONFormatter{}},
	}
	l.SetFilter(func(req *http.Request) (skip bool, err error) {
		if !strings.HasPrefix(req.URL.Host, "api.") {
			return true, nil
		}
		return false, nil
	})
	return l
}

type notFoundHandler struct{}

func (h notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "market server: page or endpoint not found", 404)
}
