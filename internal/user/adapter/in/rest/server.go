// Package rest exposes the user module's use cases as a JSON REST API on top of
// the standard library net/http server.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rzfhlv/go-std/internal/user/app"
	"github.com/rzfhlv/go-std/internal/user/domain"
)

// Server adapts HTTP requests to the user use cases.
type Server struct {
	users app.UserUseCases
}

// NewServer builds a REST adapter around the given use cases.
func NewServer(users app.UserUseCases) *Server {
	return &Server{users: users}
}

// Register mounts the user routes onto mux so a command root can compose several
// modules onto one server.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", s.listUsers)
	mux.HandleFunc("POST /users", s.createUser)
	mux.HandleFunc("GET /users/{id}", s.getUser)
}

// Handler returns a standalone handler with the user routes registered. It is
// convenient for tests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.Register(mux)
	return mux
}

type createUserRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func toResponse(user app.UserOutput) userResponse {
	return userResponse{ID: user.ID, Name: user.Name, Email: user.Email}
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	out, err := s.users.CreateUser(r.Context(), app.CreateUserInput{
		ID:    req.ID,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(out))
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	out, err := s.users.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(out))
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.users.ListUsers(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response := make([]userResponse, 0, len(users))
	for _, user := range users {
		response = append(response, toResponse(user))
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The response is already committed; an encode error cannot be surfaced.
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidUser):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
