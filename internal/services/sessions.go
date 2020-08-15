package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	// PersistentSession maintain the user logged in for up to 365 days of inactivity.
	// The 'remember me' option when signing in uses this.
	PersistentSession = "persistent"

	// EphemeralSession maintain the user logged in for up to 30 days of inactivity.
	// This work as long as the browser isn't closed or browser session isn't restored.
	// It is important to have in mind that cookie expiration might be undetermined at the client-side,
	// but at the server-side, there is always a session expiration (lower in case of an 'ephemeral' session).
	EphemeralSession = "ephemeral"
)

// Session data.
type Session struct {
	ID         string
	StickyID   string
	CreatedAt  time.Time
	Expire     time.Time
	State      string
	UserID     string
	RememberMe bool
}

// SessionFromRequest extracts the session data from a request.
func SessionFromRequest(r *http.Request) *Session {
	ctx := r.Context()
	if session, ok := ctx.Value(sessionKey{}).(*Session); ok {
		return session
	}
	return nil
}

// SessionContext adds a session to a given context.
func SessionContext(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, session)
}

type sessionKey struct{}

// Sessions services.
//
// See https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
type Sessions struct {
	core *Core
}

// SessionIDCookieName is the cookie name where the session id is stored on the browser.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#Cookie_prefixes
const SessionIDCookieName = "__Host-Market-SID"

const sessionIDLength = 343 // base64 encoded 64 bytes sticky id + comma + base64 encoded 192 bytes

// Read session cookie from request or create a new one.
func (s *Sessions) Read(w http.ResponseWriter, r *http.Request) (*Session, error) {
	var session *Session
	now := time.Now()
	cookie, err := r.Cookie(SessionIDCookieName)
	if err == nil && len(cookie.Value) == sessionIDLength {
		session, err = s.get(r.Context(), cookie.Value)
		if err != nil {
			log.Printf("request %s failed to get session from database: %v\n", r.Header.Get("X-Request-ID"), err)
		}
	}

	// Check if there is an active session.
	if err != nil || session == nil || now.After(session.Expire) || session.State != "active" {
		return s.startNewSession(w, r, makeSession(nil))
	}
	// Check if token needs to be renewed.
	if now.Before(session.CreatedAt.Add(time.Hour)) {
		return session, nil
	}
	newSession, err := s.startNewSession(w, r, makeSession(&sessionParams{
		ID:         regenerateSessionID(session.StickyID),
		StickyID:   session.StickyID,
		UserID:     session.UserID,
		RememberMe: session.RememberMe,
	}))
	if err != nil {
		log.Printf("request %s failed to renew session: %v\n", r.Header.Get("X-Request-ID"), err)
		return session, nil
	}
	return newSession, nil
}

func (s *Sessions) startNewSession(w http.ResponseWriter, r *http.Request, session *Session) (*Session, error) {
	if err := s.save(r.Context(), session); err != nil {
		return nil, fmt.Errorf("cannot write cookie: %w", err)
	}
	http.SetCookie(w, makeSessionCookie(session))
	return session, nil
}

// LoginParams to control cookie persistence, and etc.
type LoginParams struct {
	// RememberMe defines whether to set a persistent cookie that survives closing the browser or not.
	RememberMe bool
}

// Login logs out of any existing session and logs in again.
func (s *Sessions) Login(w http.ResponseWriter, r *http.Request, userID string, p LoginParams) (*Session, error) {
	oldSession := SessionFromRequest(r)
	sessionID, stickyID := newSessionID()
	session := makeSession(&sessionParams{
		ID:         sessionID,
		StickyID:   stickyID,
		UserID:     userID,
		RememberMe: p.RememberMe,
	})
	if err := s.save(r.Context(), session); err != nil {
		return nil, fmt.Errorf("cannot write cookie: %w", err)
	}
	http.SetCookie(w, makeSessionCookie(session))
	go s.expireOldSessionAfterLogin(userID, oldSession)
	return session, nil
}

// expireOldSessionAfterLogin sets an unauthenticated session to be expired one minute after login.
func (s *Sessions) expireOldSessionAfterLogin(userID string, session *Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.expireInOneMinute(ctx, session.ID); err != nil {
		log.Printf("cannot expire old session after user %q logged in: %v", userID, err)
	}
}

type sessionParams struct {
	ID         string
	StickyID   string
	UserID     string
	RememberMe bool
}

// makeSession object.
func makeSession(p *sessionParams) *Session {
	if p == nil {
		p = &sessionParams{}
		p.ID, p.StickyID = newSessionID()
	}
	var inactivity = 30 // days of inactivity
	var rememberMe bool
	if p.RememberMe {
		inactivity = 365 // days of inactivity
		rememberMe = true
	}
	return &Session{
		ID:         p.ID,
		StickyID:   p.StickyID,
		Expire:     time.Now().AddDate(0, 0, inactivity),
		State:      "active",
		UserID:     p.UserID,
		RememberMe: rememberMe,
	}
}

// makeSessionCookie object.
func makeSessionCookie(session *Session) *http.Cookie {
	cookie := &http.Cookie{
		Name:  SessionIDCookieName,
		Value: session.ID,

		Path: "/",

		Secure:   true,
		HttpOnly: true,
		// Cannot use Strict Mode as this would break passing the session cookie in situations like
		// typing the URL or redirect from an external page.
		SameSite: http.SameSiteLaxMode,
	}
	if session.RememberMe {
		cookie.Expires = session.Expire
	}
	return cookie
}

func newSessionID() (id, stickyID string) {
	var r = make([]byte, 256)
	if _, err := rand.Read(r); err != nil {
		panic(err)
	}
	enc := base64.RawURLEncoding
	sticky := enc.EncodeToString(r[:64])
	return sticky + "," + enc.EncodeToString(r[64:]), sticky
}

// regenerateSessionID based on an existing sticky ID.
// The regular session ID we use has 256 bytes of entropy.
// The first 64 bytes are carried on from the first generated session ID.
// This rotation of session ID improves security by reducing the risk of a cookie replay attack.
// It introduces a random value akin to using a forward secrecy strategy.
// We can use the sticky ID for auditing purposes, but should not use it to establish security in any form.
func regenerateSessionID(stickyID string) string {
	enc := base64.RawURLEncoding
	var nonSticky = make([]byte, 256-enc.DecodedLen(len(stickyID)))
	if _, err := rand.Read(nonSticky); err != nil {
		panic(err)
	}
	return stickyID + "," + enc.EncodeToString(nonSticky)
}

// Get session from database.
func (s *Sessions) get(ctx context.Context, sessionID string) (*Session, error) {
	var session Session
	pg := s.core.Postgres
	const sql = `SELECT "id", "sticky_id", "created_at", "expiration", "state", "user_id", "type" FROM http_sessions WHERE id = $1 LIMIT 1`
	row := pg.QueryRow(ctx, sql, sessionID)
	var t string
	switch err := row.Scan(&session.ID, &session.StickyID, &session.CreatedAt, &session.Expire, &session.State, &session.UserID, &t); {
	case err == pgx.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("error getting session: %w", err)
	}
	if t == PersistentSession {
		session.RememberMe = true
	}
	return &session, nil
}

// Save session to database.
// TODO(henvic): Cache session to Redis with a shorter lifetime like 5min.
func (s *Sessions) save(ctx context.Context, session *Session) error {
	var t = EphemeralSession
	if session.RememberMe {
		t = PersistentSession
	}
	pg := s.core.Postgres
	const sql = `INSERT INTO http_sessions ("id", "sticky_id", "created_at", "expiration", "state", "user_id", "type") VALUES ($1, $2, NOW(), $3, $4, $5, $6)`
	if _, err := pg.Exec(ctx, sql, session.ID, session.StickyID, session.Expire, session.State, session.UserID, t); err != nil {
		return fmt.Errorf("cannot save session: %w", err)
	}
	return nil
}

// expireInOneMinute an existing session that was renewed.
func (s *Sessions) expireInOneMinute(ctx context.Context, sessionID string) error {
	pg := s.core.Postgres
	const sql = `UPDATE http_sessions SET expiration = NOW() + INTERVAL '1 MINUTE' WHERE id = $1 AND expiration > NOW() AND state = 'active'`
	if _, err := pg.Exec(ctx, sql, sessionID); err != nil {
		return fmt.Errorf("error setting context to expire: %w", err)
	}
	return nil
}

// Close session.
// Revokes its sticky session id to make any existing cookie associated to it invalid.
func (s *Sessions) Close(ctx context.Context, stickyID string) error {
	pg := s.core.Postgres
	const sql = `UPDATE http_sessions SET state = 'expired' WHERE sticky_id = $1`
	if _, err := pg.Exec(ctx, sql, stickyID); err != nil {
		return fmt.Errorf("error closing session (with sticky id %q): %w", stickyID, err)
	}
	return nil
}

// CloseExpired sessions changes the state of expired sessions to mark them as expired.
// It should be called on a schedule.
func (s *Sessions) CloseExpired(ctx context.Context) (int, error) {
	panic("not implemented")
}
