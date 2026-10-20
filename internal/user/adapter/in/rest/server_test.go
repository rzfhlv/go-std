package rest_test

import (
	"context"
	"encoding/json"
	"errors"
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

func createUser(t *testing.T, h http.Handler, id, name, email string) {
	t.Helper()

	body := `{"id":"` + id + `","name":"` + name + `","email":"` + email + `"}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /users status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
}

func TestRenameUserRoute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		body       string
		seed       bool
		wantStatus int
	}{
		{name: "renames existing", id: "u1", body: `{"name":"Grace"}`, seed: true, wantStatus: http.StatusOK},
		{name: "blank name is invalid", id: "u1", body: `{"name":"   "}`, seed: true, wantStatus: http.StatusBadRequest},
		{name: "unknown user", id: "missing", body: `{"name":"Grace"}`, wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newHandler()
			if tt.seed {
				createUser(t, h, "u1", "Ada", "ada@example.com")
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/users/"+tt.id, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("PATCH /users/%s status = %d, want %d; body=%s", tt.id, rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var got struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Name != "Grace" || got.ID != "u1" || got.Email != "ada@example.com" {
				t.Fatalf("unexpected renamed user: %+v", got)
			}
		})
	}
}

func TestDeleteUserRoute(t *testing.T) {
	t.Parallel()

	h := newHandler()
	createUser(t, h, "u1", "Ada", "ada@example.com")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/users/u1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /users/u1 status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/users/u1", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("second DELETE status = %d, want 404", rec.Code)
	}
}

func TestCreateUserDuplicateEmailConflict(t *testing.T) {
	t.Parallel()

	h := newHandler()
	createUser(t, h, "u1", "Ada", "ada@example.com")
	createUser(t, h, "u2", "Grace", "grace@example.com")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/users", strings.NewReader(`{"id":"u3","name":"Ada","email":"ADA@example.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate email status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}
}

func TestListUsersCountRoute(t *testing.T) {
	t.Parallel()

	h := newHandler()
	createUser(t, h, "u1", "Ada", "ada@example.com")
	createUser(t, h, "u2", "Grace", "grace@example.com")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/users/count", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /users/count status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var got struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Count != 2 {
		t.Fatalf("count = %d, want 2", got.Count)
	}
}

// TestCountRouteWinsOverIDPattern pins that the literal /users/count route is not
// captured by /users/{id} even when no user named "count" exists.
func TestCountRouteWinsOverIDPattern(t *testing.T) {
	t.Parallel()

	h := newHandler()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/users/count", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /users/count status = %d, want 200 (not captured by /users/{id}); body=%s", rec.Code, rec.Body.String())
	}

	var got struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Count != 0 {
		t.Fatalf("count = %d, want 0 on an empty store", got.Count)
	}
}

// failingUsers is an app.UserUseCases whose CountUsers and RenameUser always
// fail, so the new handlers' error branches can be exercised. The memory-backed
// service never fails, so a stub is required here.
type failingUsers struct {
	err error
}

func (f failingUsers) CreateUser(context.Context, app.CreateUserInput) (app.UserOutput, error) {
	return app.UserOutput{}, f.err
}

func (f failingUsers) GetUser(context.Context, string) (app.UserOutput, error) {
	return app.UserOutput{}, f.err
}

func (f failingUsers) ListUsers(context.Context) ([]app.UserOutput, error) {
	return nil, f.err
}

func (f failingUsers) CountUsers(context.Context) (int, error) { return 0, f.err }

func (f failingUsers) RenameUser(context.Context, string, string) (app.UserOutput, error) {
	return app.UserOutput{}, f.err
}

func (f failingUsers) DeleteUser(context.Context, string) error { return f.err }

// TestRenameUserInvalidBody covers the new rename handler's 400 branch for a
// body that is not valid JSON.
func TestRenameUserInvalidBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"name":`},
		{name: "not an object", body: `"nope"`},
	}

	h := newHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/users/u1", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("PATCH /users/u1 status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

// TestCountAndRenameUpstreamError covers the new handlers' writeDomainError
// branch when the use case itself fails with an unknown error (-> 500).
func TestCountAndRenameUpstreamError(t *testing.T) {
	t.Parallel()

	h := rest.NewServer(failingUsers{err: errors.New("boom")}).Handler()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "count fails", method: http.MethodGet, path: "/users/count", wantStatus: http.StatusInternalServerError},
		{name: "rename fails", method: http.MethodPatch, path: "/users/u1", body: `{"name":"Grace"}`, wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s %s status = %d, want %d; body=%s", tt.method, tt.path, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
