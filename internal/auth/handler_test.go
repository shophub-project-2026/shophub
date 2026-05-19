package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shophub-project-2026/shophub/internal/auth"
)

type mockRepo struct {
	users map[string]*auth.User
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[string]*auth.User)}
}

func (m *mockRepo) Create(_ context.Context, email, hash string) (*auth.User, error) {
	if _, exists := m.users[email]; exists {
		return nil, auth.ErrEmailTaken
	}
	u := &auth.User{Email: email, PasswordHash: hash}
	m.users[email] = u
	return u, nil
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*auth.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, auth.ErrNotFound
	}
	return u, nil
}

func TestHandler_Register(t *testing.T) {
	svc := auth.NewService(newMockRepo(), "secret-key-for-testing-purposes!")
	h := auth.NewHandler(svc)

	t.Run("valid registration returns 201", func(t *testing.T) {
		body, _ := json.Marshal(auth.RegisterInput{Email: "bob@example.com", Password: "password123"})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("short password returns 400", func(t *testing.T) {
		body, _ := json.Marshal(auth.RegisterInput{Email: "carol@example.com", Password: "short"})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("duplicate email returns 409", func(t *testing.T) {
		body, _ := json.Marshal(auth.RegisterInput{Email: "bob@example.com", Password: "password123"})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestHandler_Login(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo, "secret-key-for-testing-purposes!")
	h := auth.NewHandler(svc)

	_, _ = svc.Register(context.Background(), auth.RegisterInput{
		Email: "dave@example.com", Password: "password123",
	})

	t.Run("valid credentials return 200 with token", func(t *testing.T) {
		body, _ := json.Marshal(auth.LoginInput{Email: "dave@example.com", Password: "password123"})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Login(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]string
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["token"] == "" {
			t.Error("expected token in response body")
		}
	})

	t.Run("wrong password returns 401", func(t *testing.T) {
		body, _ := json.Marshal(auth.LoginInput{Email: "dave@example.com", Password: "wrongpass"})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Login(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
