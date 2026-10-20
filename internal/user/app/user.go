package app

import (
	"context"
	"fmt"

	"github.com/rzfhlv/go-std/internal/user/domain"
	"github.com/rzfhlv/go-std/internal/user/port"
)

// CreateUserInput is the input DTO for creating a user.
type CreateUserInput struct {
	ID    string
	Name  string
	Email string
}

// UserOutput is the output DTO representing a user.
type UserOutput struct {
	ID    string
	Name  string
	Email string
}

// UserUseCases is the inbound port implemented by UserService and consumed by
// the module's inbound adapters (REST, gRPC, CLI). It is defined here, next to
// its consumer.
type UserUseCases interface {
	CreateUser(ctx context.Context, in CreateUserInput) (UserOutput, error)
	GetUser(ctx context.Context, id string) (UserOutput, error)
	ListUsers(ctx context.Context) ([]UserOutput, error)
	CountUsers(ctx context.Context) (int, error)
}

// UserService implements the user use cases.
type UserService struct {
	repo port.UserRepository
}

// NewUserService wires a UserService to its repository.
func NewUserService(repo port.UserRepository) *UserService {
	return &UserService{repo: repo}
}

var _ UserUseCases = (*UserService)(nil)

// CreateUser validates the input through the domain and persists the user.
func (s *UserService) CreateUser(ctx context.Context, in CreateUserInput) (UserOutput, error) {
	user, err := domain.NewUser(domain.UserID(in.ID), in.Name, in.Email)
	if err != nil {
		return UserOutput{}, fmt.Errorf("create user: %w", err)
	}
	if err := s.repo.Save(ctx, user); err != nil {
		return UserOutput{}, fmt.Errorf("create user: %w", err)
	}
	return toOutput(user), nil
}

// GetUser returns a single user by ID.
func (s *UserService) GetUser(ctx context.Context, id string) (UserOutput, error) {
	user, err := s.repo.FindByID(ctx, domain.UserID(id))
	if err != nil {
		return UserOutput{}, fmt.Errorf("get user %s: %w", id, err)
	}
	return toOutput(user), nil
}

// ListUsers returns every user.
func (s *UserService) ListUsers(ctx context.Context) ([]UserOutput, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	out := make([]UserOutput, 0, len(users))
	for _, user := range users {
		out = append(out, toOutput(user))
	}
	return out, nil
}

// CountUsers returns how many users are stored.
func (s *UserService) CountUsers(ctx context.Context) (int, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return len(users), nil
}

func toOutput(user domain.User) UserOutput {
	return UserOutput{
		ID:    string(user.ID()),
		Name:  user.Name(),
		Email: user.Email(),
	}
}
