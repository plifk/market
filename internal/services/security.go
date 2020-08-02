package services

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// Security module.
type Security struct {
	csrfProtection *CSRFProtection
}

// RegenerateCSRFToken on a given request. Should be called during login/logout operations.
func (s *Security) RegenerateCSRFToken(w http.ResponseWriter, r *http.Request) string {
	return s.csrfProtection.RegenerateToken(w, r)
}

// CSRFProtection protects requests against Cross-Site Request Forgery attacks.
// See https://owasp.org/www-community/attacks/csrf
// It uses https://github.com/justinas/nosurf behind the scenes.
type CSRFProtection struct {
	nosurf *nosurf.CSRFHandler
}

func (c *CSRFProtection) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.nosurf.ServeHTTP(w, r)
}

// RegenerateToken on a given request. Should be called during login/logout operations.
func (c *CSRFProtection) RegenerateToken(w http.ResponseWriter, r *http.Request) string {
	return c.nosurf.RegenerateToken(w, r)
}

// SetFailureHandler for when requests fail.
func (c *CSRFProtection) SetFailureHandler(handler http.Handler) {
	c.nosurf.SetFailureHandler(handler)
}

// ExemptFunc to bypass CSRF protection for a given request.
// This should only be used when there is already another CSRF protection in place, such as by the use of other types of tokens,
// and to allow HTTP connections from webhooks or non-browser clients that doesn't require or support CSRF protection.
// Please remember to always protect endpoints accordingly, and consider the case of browsers accessing them directly without CSRF protection.
func (c *CSRFProtection) ExemptFunc(fn func(r *http.Request) bool) {
	c.nosurf.ExemptFunc(fn)
}

// NewCSRFProtection middleware.
func NewCSRFProtection(handler http.Handler) *CSRFProtection {
	n := nosurf.New(handler)
	n.SetBaseCookie(http.Cookie{
		Path: "/",

		Secure:   true,
		HttpOnly: true,
		// Cannot use Strict Mode as this would break passing the session cookie in situations like
		// typing the URL or redirect from an external page.
		SameSite: http.SameSiteLaxMode,
	})
	return &CSRFProtection{
		nosurf: n,
	}
}

// CSRFToken from a request.
func CSRFToken(r *http.Request) string {
	return nosurf.Token(r)
}
