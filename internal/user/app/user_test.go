package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rzfhlv/go-std/internal/user/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/user/app"
	"github.com/rzfhlv/go-std/internal/user/domain"
)

func newService() *app.UserService {
	return app.NewUserService(memory.NewUserRepository())
}

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
					Email: "ada@example.com",
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
