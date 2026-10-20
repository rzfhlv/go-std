package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rzfhlv/go-std/internal/user/adapter/in/rest"
	"github.com/rzfhlv/go-std/internal/user/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/user/app"
)

func newHandler() http.Handler {
	users := app.NewUserService(memory.NewUserRepository())
	return rest.NewServer(users).Handler()
}

func TestCreateThenGetUser(t *testing.T) {
	t.Parallel()

	h := newHandler()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/users", strings.NewReader(`{"id":"u1","name":"Ada","email":"ada@example.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /users status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/users/u1", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /users/u1 status = %d, want 200", rec.Code)
	}

	var got struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "u1" || got.Email != "ada@example.com" {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestCreateUserInvalidEmail(t *testing.T) {
	t.Parallel()

	h := newHandler()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/users", strings.NewReader(`{"id":"u1","name":"Ada","email":"bad"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetUserMissing(t *testing.T) {
	t.Parallel()

	h := newHandler()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/users/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
