package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors returned by domain operations. Callers match them with
// errors.Is.
var (
	// ErrInvalidUser is returned when a user fails validation.
	ErrInvalidUser = errors.New("invalid user")
	// ErrUserNotFound is returned when a user cannot be located.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailTaken is returned when a user is created with an email that
	// already belongs to another user.
	ErrEmailTaken = errors.New("email already in use")
)

// UserID identifies a User.
type UserID string

// User is the user aggregate root. Its fields are unexported so every mutation
// goes through a method that enforces the aggregate's invariants.
type User struct {
	id    UserID
	name  string
	email string
}

// NormalizeEmail canonicalizes an email address the way the module stores it:
// surrounding whitespace is trimmed and the address is lowercased. It is the
// single source of truth for email normalization, so lookups and storage agree.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NewUser constructs a validated User. Name is trimmed and email is normalized
// to lowercase.
func NewUser(id UserID, name, email string) (User, error) {
	u := User{
		id:    id,
		name:  strings.TrimSpace(name),
		email: NormalizeEmail(email),
	}
	if err := u.validate(); err != nil {
		return User{}, err
	}
	return u, nil
}

func (u User) validate() error {
	if u.id == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidUser)
	}
	if u.name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidUser)
	}
	if !strings.Contains(u.email, "@") {
		return fmt.Errorf("%w: email %q is malformed", ErrInvalidUser, u.email)
	}
	return nil
}

// ID returns the user identifier.
func (u User) ID() UserID { return u.id }

// Name returns the user's display name.
func (u User) Name() string { return u.name }

// Email returns the user's normalized email address.
func (u User) Email() string { return u.email }

// Rename returns a copy of the user with a new name, re-checking invariants.
func (u User) Rename(name string) (User, error) {
	updated := u
	updated.name = strings.TrimSpace(name)
	if updated.name == "" {
		return u, fmt.Errorf("%w: name is required", ErrInvalidUser)
	}
	return updated, nil
}
