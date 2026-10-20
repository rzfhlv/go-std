package cli_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rzfhlv/go-std/internal/user/adapter/in/cli"
	"github.com/rzfhlv/go-std/internal/user/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/user/app"
	"github.com/rzfhlv/go-std/internal/user/domain"
)

// stubUsers is a fake app.UserUseCases with programmable per-method results.
type stubUsers struct {
	renameOut app.UserOutput
	renameErr error
	deleteErr error
	getErr    error
}

func newCLI(stub *stubUsers) (*cli.CLI, *strings.Builder) {
	var out strings.Builder
	return cli.New(stub, &out), &out
}

func (s *stubUsers) CreateUser(context.Context, app.CreateUserInput) (app.UserOutput, error) {
	return app.UserOutput{}, errors.New("not implemented")
}

func (s *stubUsers) GetUser(context.Context, string) (app.UserOutput, error) {
	return app.UserOutput{}, s.getErr
}

func (s *stubUsers) ListUsers(context.Context) ([]app.UserOutput, error) {
	return nil, errors.New("not implemented")
}

func (s *stubUsers) CountUsers(context.Context) (int, error) {
	return 0, errors.New("not implemented")
}

func (s *stubUsers) RenameUser(context.Context, string, string) (app.UserOutput, error) {
	return s.renameOut, s.renameErr
}

func (s *stubUsers) DeleteUser(context.Context, string) error {
	return s.deleteErr
}

func TestRenameCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		stub      *stubUsers
		wantErr   bool
		wantInOut string
	}{
		{
			name:      "renames a user",
			args:      []string{"rename", "-id", "u1", "-name", "Grace"},
			stub:      &stubUsers{renameOut: app.UserOutput{ID: "u1", Name: "Grace", Email: "ada@example.com"}},
			wantInOut: "Grace",
		},
		{
			name:    "blank name is rejected",
			args:    []string{"rename", "-id", "u1", "-name", "  "},
			stub:    &stubUsers{renameErr: domain.ErrInvalidUser},
			wantErr: true,
		},
		{
			name:    "unknown user propagates",
			args:    []string{"rename", "-id", "missing", "-name", "Grace"},
			stub:    &stubUsers{renameErr: domain.ErrUserNotFound},
			wantErr: true,
		},
		{
			name:    "missing id is a usage error",
			args:    []string{"rename", "-name", "Grace"},
			stub:    &stubUsers{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, out := newCLI(tt.stub)
			err := c.Run(context.Background(), tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Run error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if !strings.Contains(out.String(), tt.wantInOut) {
				t.Fatalf("output = %q, want it to contain %q", out.String(), tt.wantInOut)
			}
		})
	}
}

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		stub    *stubUsers
		wantErr bool
	}{
		{
			name: "deletes a user",
			args: []string{"delete", "-id", "u1"},
			stub: &stubUsers{},
		},
		{
			name:    "deleting again fails",
			args:    []string{"delete", "-id", "u1"},
			stub:    &stubUsers{deleteErr: domain.ErrUserNotFound},
			wantErr: true,
		},
		{
			name:    "missing id is a usage error",
			args:    []string{"delete"},
			stub:    &stubUsers{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, out := newCLI(tt.stub)
			err := c.Run(context.Background(), tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Run error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if !strings.Contains(out.String(), "u1") {
				t.Fatalf("output = %q, want it to mention u1", out.String())
			}
		})
	}
}

func TestGetErrorMessage(t *testing.T) {
	t.Parallel()

	c, _ := newCLI(&stubUsers{getErr: domain.ErrUserNotFound})
	if err := c.Run(context.Background(), []string{"get", "missing"}); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("Run error = %v, want ErrUserNotFound", err)
	}
}

// TestCreateCommandAgainstMemory exercises the real wiring end to end.
func TestCreateCommandAgainstMemory(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	svc := app.NewUserService(memory.NewUserRepository())
	c := cli.New(svc, &out)

	if err := c.Run(context.Background(), []string{"create", "-id", "u1", "-name", "Ada", "-email", "ada@example.com"}); err != nil {
		t.Fatalf("create error = %v", err)
	}
	if err := c.Run(context.Background(), []string{"rename", "-id", "u1", "-name", "Grace"}); err != nil {
		t.Fatalf("rename error = %v", err)
	}
	if err := c.Run(context.Background(), []string{"delete", "-id", "u1"}); err != nil {
		t.Fatalf("delete error = %v", err)
	}
	if err := c.Run(context.Background(), []string{"count"}); err != nil {
		t.Fatalf("count error = %v", err)
	}
}

// TestRenameAndDeleteFlagParseError covers the new rename/delete handlers' flag
// parse error branch.
func TestRenameAndDeleteFlagParseError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "rename unknown flag", args: []string{"rename", "-nope", "x"}},
		{name: "delete unknown flag", args: []string{"delete", "-nope", "x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := newCLI(&stubUsers{})
			err := c.Run(context.Background(), tt.args)
			if err == nil {
				t.Fatal("Run error = nil, want parse error")
			}
			if !strings.Contains(err.Error(), "parse flags") {
				t.Fatalf("Run error = %v, want it to mention %q", err, "parse flags")
			}
		})
	}
}
