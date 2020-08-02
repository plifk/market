package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/nyaruka/phonenumbers"
	"github.com/plifk/market/internal/passwords"
	"golang.org/x/crypto/bcrypt"
)

// User structure
type User struct {
	UserID    string
	Name      string
	Email     string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Access    Authorization
}

// Authorization role levels.
type Authorization string

var (
	// UserAuthorization role.
	UserAuthorization Authorization = "user"

	// AdminAuthorization role.
	AdminAuthorization Authorization = "admin"
)

// NewUserParams to create a new user.
type NewUserParams struct {
	Name   string
	Email  string
	Phone  string
	Access Authorization
}

// ValidateAndNormalize user params.
func (p *NewUserParams) ValidateAndNormalize() error {
	if len(p.Name) == 0 {
		return errors.New("missing name")
	}
	if len(p.Name) > 150 {
		return errors.New("name must be at most 150 chars")
	}
	if err := validateEmail(p.Email); err != nil {
		return err
	}
	if p.Phone != "" {
		// TODO(henvic): Set default region and normalize phone.
		num, err := phonenumbers.Parse(p.Phone, "US")
		if err != nil {
			return fmt.Errorf("invalid phone number: %w", err)
		}
		p.Phone = num.String()
	}
	return nil
}

// Accounts services.
type Accounts struct {
	core *Core
}

// FromRequest gets an user account from a given request.
func (a *Accounts) FromRequest(r *http.Request) *User {
	session := GetSession(r)
	if session == nil && session.UserID != "" {
		return nil
	}
	// TODO(henvic): Add caching layer.
	u, _ := a.GetUserByID(r.Context(), session.UserID)
	return u
}

// NewUser creates a new user.
func (a *Accounts) NewUser(ctx context.Context, p NewUserParams) (id string, err error) {
	if err = p.ValidateAndNormalize(); err != nil {
		return "", err
	}
	id = new11RandomID()
	pg := a.core.Postgres
	if p.Access == "" {
		p.Access = UserAuthorization
	}
	sql := `INSERT INTO users ("user_id", "name", "email", "phone", "created_at", "access") VALUES ($1, $2, $3, $4, NOW(), $5)`
	c, err := pg.Exec(ctx, sql, id, p.Name, p.Phone, p.Email, string(p.Access))
	if err == nil && c.RowsAffected() == 0 {
		err = errors.New("cannot save to database")
	}
	if err != nil {
		return id, fmt.Errorf("cannot create user %q: %w", id, err)
	}
	return id, nil
}

// GetUserByID and return user object.
func (a *Accounts) GetUserByID(ctx context.Context, userID string) (*User, error) {
	pg := a.core.Postgres
	const sql = `SELECT "user_id", "name", "email", "phone", "created_at", "updated_at", "access" FROM users WHERE "user_id" = $1 LIMIT 1`
	row := pg.QueryRow(ctx, sql, userID)
	var u User
	err := row.Scan(&u.UserID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt, &u.UpdatedAt, &u.Access)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &u, err
}

// ErrUserNotFound occurs when no user is found.
var ErrUserNotFound = errors.New("user not found")

// GetUserByEmail and return user object.
func (a *Accounts) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	pg := a.core.Postgres
	const sql = `SELECT "user_id", "name", "email", "phone", "created_at", "updated_at", "access" FROM users WHERE "email" = $1 LIMIT 1`
	row := pg.QueryRow(ctx, sql, email)
	var u User
	err := row.Scan(&u.UserID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt, &u.UpdatedAt, &u.Access)
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return &u, err
}

// NewAdminParams required to create a new admin user.
type NewAdminParams struct {
	NewUserParams
	Password string
}

// NewAdmin creates a new admin user.
func (a *Accounts) NewAdmin(ctx context.Context, p NewAdminParams) (id string, err error) {
	if err = passwords.Validate(p.Password); err != nil {
		return "", err
	}
	id, err = a.NewUser(ctx, p.NewUserParams)
	if err != nil {
		return "", err
	}

	if err := a.SetCredentials(ctx, SetPasswordParams{
		UserID:   id,
		Password: p.Password,
	}); err != nil {
		return id, err
	}
	return id, nil
}

// SetPasswordParams for a given user.
type SetPasswordParams struct {
	UserID   string
	Password string
}

// SetCredentials for user.
func (a *Accounts) SetCredentials(ctx context.Context, p SetPasswordParams) error {
	if err := passwords.Validate(p.Password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("cannot encrypt password: %w", err)
	}

	pg := a.core.Postgres
	const sql = `INSERT INTO users_credentials ("user_id", "password_hash") VALUES($1, $2) ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash`
	switch c, err := pg.Exec(ctx, sql, p.UserID, hash); {
	case err != nil:
		return fmt.Errorf("error setting a credential for user %q: %w", p.UserID, err)
	case c.RowsAffected() == 0:
		return fmt.Errorf("error setting a credential for user %q: no rows affected", p.UserID)
	}
	return nil
}

type checkPasswordRow struct {
	PasswordHash string
	UpdatedAt    time.Time
}

// ErrWrongPassword is used after failing to verify password.
var ErrWrongPassword = errors.New("wrong password")

// CheckPassword for user.
func (a *Accounts) CheckPassword(ctx context.Context, userID, password string) error {
	if password == "" {
		return errors.New("password is empty")
	}
	if len(password) > passwords.MaxPasswordLength {
		return errors.New("password is longer than acceptable")
	}

	pg := a.core.Postgres
	const sql = `SELECT "password_hash", "updated_at" FROM users_credentials WHERE user_id = $1 LIMIT 1`
	row := pg.QueryRow(ctx, sql, userID)
	var r checkPasswordRow
	if err := row.Scan(&r.PasswordHash, &r.UpdatedAt); err != nil {
		return fmt.Errorf("cannot check password: %w", err)
	}

	err := bcrypt.CompareHashAndPassword([]byte(r.PasswordHash), []byte(password))
	if err == nil {
		return nil
	}
	if err != bcrypt.ErrMismatchedHashAndPassword {
		log.Printf("cannot compare password for user %s: %v", userID, err)
	}
	return ErrWrongPassword
}

func validateEmail(address string) error {
	if address == "" {
		return errors.New("missing email address")
	}
	if len(address) > 255 {
		return errors.New("email address is too long")
	}
	// ParseAddress takes Name <email>, we compare the received address to the Address field below
	// to verify if we received only the email address, without a name.
	if a, err := mail.ParseAddress(address); err != nil || address != a.Address {
		return errors.New("invalid email address")
	}
	return nil
}
