package domain_test

import (
	"errors"
	"testing"

	"github.com/rzfhlv/go-std/internal/user/domain"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      domain.UserID
		inName  string
		email   string
		wantErr error
	}{
		{name: "valid", id: "u1", inName: "Ada", email: "ada@example.com"},
		{name: "trims and lowercases", id: "u2", inName: "  Grace  ", email: " GRACE@EXAMPLE.COM "},
		{name: "missing id", id: "", inName: "Ada", email: "ada@example.com", wantErr: domain.ErrInvalidUser},
		{name: "blank name", id: "u3", inName: "   ", email: "ada@example.com", wantErr: domain.ErrInvalidUser},
		{name: "malformed email", id: "u4", inName: "Ada", email: "not-an-email", wantErr: domain.ErrInvalidUser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, err := domain.NewUser(tt.id, tt.inName, tt.email)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewUser() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewUser() unexpected error: %v", err)
			}
			if user.ID() != tt.id {
				t.Errorf("ID() = %q, want %q", user.ID(), tt.id)
			}
		})
	}
}

func TestUserRename(t *testing.T) {
	t.Parallel()

	user, err := domain.NewUser("u1", "Ada", "ada@example.com")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	renamed, err := user.Rename(" Grace ")
	if err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if renamed.Name() != "Grace" {
		t.Errorf("Name() = %q, want %q", renamed.Name(), "Grace")
	}

	if _, err := user.Rename("  "); !errors.Is(err, domain.ErrInvalidUser) {
		t.Errorf("Rename(blank) error = %v, want ErrInvalidUser", err)
	}
}
