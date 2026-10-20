package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rzfhlv/go-std/internal/user/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/user/app"
	"github.com/rzfhlv/go-std/internal/user/domain"
	"github.com/rzfhlv/go-std/internal/user/port"
)

func newService() *app.UserService {
	return app.NewUserService(memory.NewUserRepository())
}

// errRepo is a port.UserRepository whose FindByEmail and Save return a
// programmable error, so the use-case error paths can be exercised. The in-memory
// adapter only ever succeeds, so a stub is required here.
type errRepo struct {
	findByEmailErr error
	saveErr        error
}

var _ port.UserRepository = (*errRepo)(nil)

func (r *errRepo) Save(context.Context, domain.User) error { return r.saveErr }

func (r *errRepo) FindByID(context.Context, domain.UserID) (domain.User, error) {
	user, err := domain.NewUser("u1", "Ada", "ada@example.com")
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (r *errRepo) FindByEmail(context.Context, string) (domain.User, error) {
	return domain.User{}, r.findByEmailErr
}

func (r *errRepo) List(context.Context) ([]domain.User, error) { return nil, nil }

func (r *errRepo) Delete(context.Context, domain.UserID) error { return nil }

func TestCreateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      app.CreateUserInput
		wantErr error
	}{
		{
			name: "valid",
			in:   app.CreateUserInput{ID: "u1", Name: "Ada", Email: "ada@example.com"},
		},
		{
			name:    "invalid email",
			in:      app.CreateUserInput{ID: "u2", Name: "Ada", Email: "bad"},
			wantErr: domain.ErrInvalidUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService()
			if _, err := svc.CreateUser(context.Background(), tt.in); !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUserNotFound(t *testing.T) {
	t.Parallel()

	svc := newService()
	if _, err := svc.GetUser(context.Background(), "missing"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("GetUser() error = %v, want ErrUserNotFound", err)
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	svc := newService()
	ctx := context.Background()
	if _, err := svc.CreateUser(ctx, app.CreateUserInput{ID: "u1", Name: "Ada", Email: "ada@example.com"}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	users, err := svc.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("ListUsers() len = %d, want 1", len(users))
	}
}

func TestCountUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		seed int
		want int
	}{
		{name: "empty", seed: 0, want: 0},
		{name: "three", seed: 3, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService()
			ctx := context.Background()
			for i := range tt.seed {
				in := app.CreateUserInput{
					ID:    fmt.Sprintf("u%d", i),
					Name:  "Ada",
					Email: fmt.Sprintf("ada%d@example.com", i),
				}
				if _, err := svc.CreateUser(ctx, in); err != nil {
					t.Fatalf("seed user %d: %v", i, err)
				}
			}

			n, err := svc.CountUsers(ctx)
			if err != nil {
				t.Fatalf("CountUsers() error = %v", err)
			}
			if n != tt.want {
				t.Fatalf("CountUsers() = %d, want %d", n, tt.want)
			}
		})
	}
}

func TestRenameUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		newName string
		wantErr error
	}{
		{name: "renames existing", id: "u1", newName: "Grace"},
		{name: "blank name is invalid", id: "u1", newName: "   ", wantErr: domain.ErrInvalidUser},
		{name: "unknown user", id: "missing", newName: "Grace", wantErr: domain.ErrUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService()
			ctx := context.Background()
			if _, err := svc.CreateUser(ctx, app.CreateUserInput{ID: "u1", Name: "Ada", Email: "ada@example.com"}); err != nil {
				t.Fatalf("seed user: %v", err)
			}

			out, err := svc.RenameUser(ctx, tt.id, tt.newName)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("RenameUser() error = %v, want %v", err, tt.wantErr)
				}
				if tt.wantErr == domain.ErrInvalidUser {
					stored, getErr := svc.GetUser(ctx, tt.id)
					if getErr != nil {
						t.Fatalf("GetUser() after rejected rename: %v", getErr)
					}
					if stored.Name != "Ada" {
						t.Errorf("stored name = %q, want %q after a rejected rename", stored.Name, "Ada")
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("RenameUser() unexpected error: %v", err)
			}
			if out.Name != tt.newName {
				t.Errorf("Name = %q, want %q", out.Name, tt.newName)
			}
			if out.ID != "u1" || out.Email != "ada@example.com" {
				t.Errorf("rename changed id/email: %+v", out)
			}

			reloaded, err := svc.GetUser(ctx, "u1")
			if err != nil {
				t.Fatalf("GetUser() error = %v", err)
			}
			if reloaded.Name != tt.newName {
				t.Errorf("stored name = %q, want %q", reloaded.Name, tt.newName)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	svc := newService()
	ctx := context.Background()
	if _, err := svc.CreateUser(ctx, app.CreateUserInput{ID: "u1", Name: "Ada", Email: "ada@example.com"}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	if err := svc.DeleteUser(ctx, "u1"); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if _, err := svc.GetUser(ctx, "u1"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("GetUser() after delete error = %v, want ErrUserNotFound", err)
	}
	if err := svc.DeleteUser(ctx, "u1"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("second DeleteUser() error = %v, want ErrUserNotFound", err)
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		secondMail string
		wantErr    error
	}{
		{name: "exact duplicate", secondMail: "ada@example.com", wantErr: domain.ErrEmailTaken},
		{name: "case and space variant", secondMail: "  ADA@EXAMPLE.COM ", wantErr: domain.ErrEmailTaken},
		{name: "different email is allowed", secondMail: "grace@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService()
			ctx := context.Background()
			if _, err := svc.CreateUser(ctx, app.CreateUserInput{ID: "u1", Name: "Ada", Email: "ada@example.com"}); err != nil {
				t.Fatalf("seed user: %v", err)
			}

			_, err := svc.CreateUser(ctx, app.CreateUserInput{ID: "u2", Name: "Ada", Email: tt.secondMail})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateUser() error = %v, want %v", err, tt.wantErr)
			}

			n, err := svc.CountUsers(ctx)
			if err != nil {
				t.Fatalf("CountUsers() error = %v", err)
			}
			wantCount := 1
			if tt.wantErr == nil {
				wantCount = 2
			}
			if n != wantCount {
				t.Fatalf("CountUsers() = %d, want %d", n, wantCount)
			}
		})
	}
}

// TestCreateUserRepositoryError covers the new FindByEmail guard's third
// outcome: a repository error that is neither a hit nor ErrUserNotFound must be
// propagated (wrapped), not mistaken for "email is free".
func TestCreateUserRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("boom")

	tests := []struct {
		name       string
		findErr    error
		wantRepoEr bool
	}{
		{name: "unexpected find error propagates", findErr: repoErr, wantRepoEr: true},
		{name: "not found means free", findErr: domain.ErrUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := app.NewUserService(&errRepo{findByEmailErr: tt.findErr})
			_, err := svc.CreateUser(context.Background(), app.CreateUserInput{ID: "u1", Name: "Ada", Email: "ada@example.com"})

			if tt.wantRepoEr {
				if !errors.Is(err, repoErr) {
					t.Fatalf("CreateUser() error = %v, want it to wrap %v", err, repoErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateUser() error = %v, want nil", err)
			}
		})
	}
}

// TestRenameUserSaveError covers the new RenameUser save-failure branch.
func TestRenameUserSaveError(t *testing.T) {
	t.Parallel()

	saveErr := errors.New("disk full")
	svc := app.NewUserService(&errRepo{saveErr: saveErr})

	_, err := svc.RenameUser(context.Background(), "u1", "Grace")
	if !errors.Is(err, saveErr) {
		t.Fatalf("RenameUser() error = %v, want it to wrap %v", err, saveErr)
	}
}
