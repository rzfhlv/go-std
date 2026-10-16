// Package memory provides the user module's in-memory implementation of the user
// repository outbound port. It is intended for tests and local development.
package memory

import (
	"context"
	"sync"

	"github.com/rzfhlv/go-std/internal/user/domain"
	"github.com/rzfhlv/go-std/internal/user/port"
)

// UserRepository stores users in memory and is safe for concurrent use.
type UserRepository struct {
	mu    sync.RWMutex
	users map[domain.UserID]domain.User
}

// NewUserRepository returns an empty in-memory repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{users: make(map[domain.UserID]domain.User)}
}

var _ port.UserRepository = (*UserRepository)(nil)

// Save implements port.UserRepository.
func (r *UserRepository) Save(_ context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID()] = user
	return nil
}

// FindByID implements port.UserRepository.
func (r *UserRepository) FindByID(_ context.Context, id domain.UserID) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

// List implements port.UserRepository.
func (r *UserRepository) List(_ context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}
