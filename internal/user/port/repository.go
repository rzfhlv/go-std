// Package port defines the user module's outbound interfaces (what the module
// needs from infrastructure). It contains interfaces only.
package port

import (
	"context"

	"github.com/rzfhlv/go-std/internal/user/domain"
)

// UserRepository is an outbound port for persisting users. Implementations live
// under the module's internal/adapter/out.
type UserRepository interface {
	// Save persists a user, overwriting any entry with the same ID.
	Save(ctx context.Context, user domain.User) error
	// FindByID returns the user with the given ID or domain.ErrUserNotFound.
	FindByID(ctx context.Context, id domain.UserID) (domain.User, error)
	// FindByEmail returns the user whose email matches the given address,
	// compared through domain.NormalizeEmail, or domain.ErrUserNotFound.
	// Implementations must apply domain.NormalizeEmail rather than assume the
	// caller already did, so every store agrees on what "same email" means.
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	// List returns every stored user.
	List(ctx context.Context) ([]domain.User, error)
	// Delete removes the user with the given ID or returns
	// domain.ErrUserNotFound.
	Delete(ctx context.Context, id domain.UserID) error
}
