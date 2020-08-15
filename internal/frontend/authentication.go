package frontend

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/plifk/market/internal/services"
	"github.com/plifk/market/internal/validator"
)

// LoginHandler handles the user authentication.
type LoginHandler struct {
	Frontend *Frontend
}

// redirectAfterLogin checks whether the redirect parameter is a path to a safe internal URL.
// It tests if it is safe to redirect. If not, returns to the home page (/).
func redirectAfterLogin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirect := q.Get("redirect_uri")
	// We want Scheme, Opaque, User, Host, RawQuery, and Fragment not to be defined.
	// Testing for non parsing error and path = redirect should be enough.
	if u, err := url.Parse(redirect); err != nil || u.Path != redirect {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session := services.SessionFromRequest(r)
	if session != nil && session.UserID != "" {
		switch r.Method {
		// Post is accepted because if the user opens up several login windows,
		// we want to let them go to the right place after logging in.
		case http.MethodHead, http.MethodGet, http.MethodPost:
			redirectAfterLogin(w, r)
			return
		default:
			h.Frontend.HTTPError(w, r, http.StatusMethodNotAllowed)
			return
		}
	}
	if r.Method == http.MethodPost {
		h.loginPostHandler(w, r)
		return
	}
	h.loginGetHandler(w, r, nil)
}

// LoginForm for /login.
type LoginForm struct {
	Email      string
	RememberMe bool

	Error error
}

func (h *LoginHandler) loginGetHandler(w http.ResponseWriter, r *http.Request, err error) {
	rememberMe := true
	if r.Method == http.MethodPost {
		rememberMe = r.PostFormValue("remember_me") == "on"
	}
	form := LoginForm{
		Email:      r.PostFormValue("email"),
		RememberMe: rememberMe,
		Error:      err,
	}
	resp := &HTMLResponse{
		Template: "account-login",
		Title:    "Login",
		Content:  form,
	}
	h.Frontend.Respond(w, r, resp)
}

func (h *LoginHandler) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	rememberMe := r.PostFormValue("remember_me") == "on"

	var fe validator.FormError
	if len(email) > 255 {
		h.loginGetHandler(w, r, fe.Append("email", errors.New("email address is too long")))
		return
	}

	modules := h.Frontend.Modules
	accounts := modules.Accounts
	u, err := accounts.GetUserByEmail(r.Context(), email)
	switch {
	case err == services.ErrUserNotFound:
		h.loginGetHandler(w, r, fe.Append("email", err))
		return
	case err != nil:
		log.Printf("cannot get user by email on /login: %v", err)
		h.Frontend.HTTPError(w, r, http.StatusInternalServerError)
		return
	}

	err = accounts.CheckPassword(r.Context(), u.UserID, password)
	switch {
	case err == services.ErrWrongPassword:
		h.loginGetHandler(w, r, fe.Append("password", err))
		return
	case err != nil:
		h.Frontend.HTTPError(w, r, http.StatusInternalServerError)
		return
	}

	if session := services.SessionFromRequest(r); session != nil {
		if err := modules.Sessions.Close(r.Context(), session.StickyID); err != nil {
			log.Printf("cannot close session with sticky id %q: %v", session.StickyID, err)
			h.Frontend.HTTPError(w, r, http.StatusInternalServerError)
			return
		}
	}

	params := services.LoginParams{
		RememberMe: rememberMe,
	}
	modules.Security.RegenerateCSRFToken(w, r) // Create new CSRF token.
	if session, err := modules.Sessions.Login(w, r, u.UserID, params); err != nil {
		log.Printf("cannot login user %q (with sticky id %q): %v", u.UserID, session.StickyID, err)
		h.Frontend.HTTPError(w, r, http.StatusInternalServerError)
		return
	}
	redirectAfterLogin(w, r)
}

// LogoutHandler manages the logout endpoint.
type LogoutHandler struct {
	Frontend *Frontend
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session := services.SessionFromRequest(r)
	if session == nil || session.UserID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		resp := &HTMLResponse{
			Template: "account-logout",
			Title:    "Logout",
		}
		h.Frontend.Respond(w, r, resp)
		return
	}

	modules := h.Frontend.Modules
	modules.Security.RegenerateCSRFToken(w, r) // We want to make sure we invalidate CSRF tokens immediately.
	if session := services.SessionFromRequest(r); session != nil {
		if err := modules.Sessions.Close(r.Context(), session.StickyID); err != nil {
			log.Printf("cannot close session with sticky id %q: %v", session.StickyID, err)
			h.Frontend.HTTPError(w, r, http.StatusInternalServerError)
			return
		}
	}
	http.SetCookie(w, expireSessionCookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

var expireSessionCookie = &http.Cookie{
	Name:   services.SessionIDCookieName,
	Value:  "",
	Path:   "/",
	MaxAge: -1,
}
