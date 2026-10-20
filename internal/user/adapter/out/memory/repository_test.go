package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rzfhlv/go-std/internal/user/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/user/domain"
)

func seedUser(t *testing.T, repo *memory.UserRepository, id, name, email string) {
	t.Helper()

	user, err := domain.NewUser(domain.UserID(id), name, email)
	if err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatalf("save user %s: %v", id, err)
	}
}

func TestRepositoryDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		delete  string
		wantErr error
	}{
		{name: "existing user", delete: "u1"},
		{name: "unknown user", delete: "missing", wantErr: domain.ErrUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := memory.NewUserRepository()
			ctx := context.Background()
			seedUser(t, repo, "u1", "Ada", "ada@example.com")

			err := repo.Delete(ctx, domain.UserID(tt.delete))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Delete() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if _, err := repo.FindByID(ctx, "u1"); !errors.Is(err, domain.ErrUserNotFound) {
				t.Fatalf("FindByID() after delete error = %v, want ErrUserNotFound", err)
			}
		})
	}
}

func TestRepositoryFindByEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		lookup  string
		wantID  domain.UserID
		wantErr error
	}{
		{name: "exact match", lookup: "ada@example.com", wantID: "u1"},
		{name: "normalizes case and space", lookup: "  ADA@EXAMPLE.COM ", wantID: "u1"},
		{name: "unknown email", lookup: "nobody@example.com", wantErr: domain.ErrUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := memory.NewUserRepository()
			seedUser(t, repo, "u1", "Ada", "ada@example.com")

			user, err := repo.FindByEmail(context.Background(), tt.lookup)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("FindByEmail() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if user.ID() != tt.wantID {
				t.Fatalf("FindByEmail() id = %q, want %q", user.ID(), tt.wantID)
			}
		})
	}
}
