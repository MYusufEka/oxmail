package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/MYusufEka/oxmail/internal/domain"
)

type mockUserLookup struct {
	users       map[string]*domain.User
	listErr     error
	getErr      error
	updateErr   error
	updatedUser *domain.User
}

func (m *mockUserLookup) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	user, ok := m.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserLookup) Update(_ context.Context, id int64, req domain.UpdateUserRequest) (*domain.User, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	for _, u := range m.users {
		if u.ID == id {
			if req.Password != nil {
				u.PasswordHash = *req.Password
			}
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserLookup) List(_ context.Context, _ domain.UserListParams) ([]domain.User, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var users []domain.User
	for _, u := range m.users {
		users = append(users, *u)
	}
	return users, len(users), nil
}

func setupAuthRouter(adminPassword, jwtSecret string) *chi.Mux {
	os.Setenv("OXMAIL_ADMIN_PASSWORD", adminPassword)
	os.Setenv("OXMAIL_JWT_SECRET", jwtSecret)

	router := chi.NewRouter()
	handler := NewAuthHandler(jwtSecret, adminPassword, nil, nil)
	handler.RegisterRoutes(router)
	return router
}

func setupAuthRouterWithUserLookup(adminPassword, jwtSecret string, lookup UserLookup) *chi.Mux {
	router := chi.NewRouter()
	handler := NewAuthHandler(jwtSecret, adminPassword, lookup, nil)
	handler.RegisterRoutes(router)
	return router
}

func hashPasswordForTest(pw string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), 4)
	return string(hash)
}

func TestLoginSuccess(t *testing.T) {
	router := setupAuthRouter("secret123", "test-jwt-secret")
	defer os.Unsetenv("OXMAIL_ADMIN_PASSWORD")
	defer os.Unsetenv("OXMAIL_JWT_SECRET")

	body := domain.LoginRequest{
		Email:    "admin",
		Password: "secret123",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.LoginResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.False(t, resp.ExpiresAt.IsZero())
}

func TestLoginInvalidPassword(t *testing.T) {
	router := setupAuthRouter("secret123", "test-jwt-secret")
	defer os.Unsetenv("OXMAIL_ADMIN_PASSWORD")
	defer os.Unsetenv("OXMAIL_JWT_SECRET")

	body := domain.LoginRequest{
		Email:    "admin",
		Password: "wrongpassword",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp domain.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", resp.Error.Code)
}

func TestLoginEmptyBody(t *testing.T) {
	router := setupAuthRouter("secret123", "test-jwt-secret")
	defer os.Unsetenv("OXMAIL_ADMIN_PASSWORD")
	defer os.Unsetenv("OXMAIL_JWT_SECRET")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLoginRateLimit(t *testing.T) {
	router := setupAuthRouter("secret123", "test-jwt-secret")
	defer os.Unsetenv("OXMAIL_ADMIN_PASSWORD")
	defer os.Unsetenv("OXMAIL_JWT_SECRET")

	body := domain.LoginRequest{
		Email:    "admin",
		Password: "wrongpassword",
	}
	payload, _ := json.Marshal(body)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "attempt %d should be 401", i+1)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestLoginInvalidJSON(t *testing.T) {
	router := setupAuthRouter("secret123", "test-jwt-secret")
	defer os.Unsetenv("OXMAIL_ADMIN_PASSWORD")
	defer os.Unsetenv("OXMAIL_JWT_SECRET")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestChangePasswordSuccess(t *testing.T) {
	mock := &mockUserLookup{
		users: map[string]*domain.User{
			"alice@local.test": {
				ID:           1,
				Email:        "alice@local.test",
				PasswordHash: hashPasswordForTest("OldPass123!"),
			},
		},
	}
	router := setupAuthRouterWithUserLookup("secret123", "test-jwt-secret", mock)

	body := domain.ChangePasswordRequest{
		Email:           "alice@local.test",
		CurrentPassword: "OldPass123!",
		NewPassword:     "NewPass456!",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "password_changed", resp["status"])

	user := mock.users["alice@local.test"]
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("NewPass456!")))
}

func TestChangePasswordWrongCurrentPassword(t *testing.T) {
	mock := &mockUserLookup{
		users: map[string]*domain.User{
			"alice@local.test": {
				ID:           1,
				Email:        "alice@local.test",
				PasswordHash: hashPasswordForTest("OldPass123!"),
			},
		},
	}
	router := setupAuthRouterWithUserLookup("secret123", "test-jwt-secret", mock)

	body := domain.ChangePasswordRequest{
		Email:           "alice@local.test",
		CurrentPassword: "WrongPass999!",
		NewPassword:     "NewPass456!",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp domain.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", resp.Error.Code)
}

func TestChangePasswordUserNotFound(t *testing.T) {
	mock := &mockUserLookup{
		users: map[string]*domain.User{},
	}
	router := setupAuthRouterWithUserLookup("secret123", "test-jwt-secret", mock)

	body := domain.ChangePasswordRequest{
		Email:           "nobody@local.test",
		CurrentPassword: "AnyPass123!",
		NewPassword:     "NewPass456!",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp domain.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "USER_NOT_FOUND", resp.Error.Code)
}

func TestChangePasswordMissingFields(t *testing.T) {
	mock := &mockUserLookup{users: map[string]*domain.User{}}
	router := setupAuthRouterWithUserLookup("secret123", "test-jwt-secret", mock)

	tests := []struct {
		name string
		body domain.ChangePasswordRequest
	}{
		{"empty email", domain.ChangePasswordRequest{Email: "", CurrentPassword: "a", NewPassword: "b"}},
		{"empty current password", domain.ChangePasswordRequest{Email: "a@b.com", CurrentPassword: "", NewPassword: "b"}},
		{"empty new password", domain.ChangePasswordRequest{Email: "a@b.com", CurrentPassword: "a", NewPassword: ""}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestChangePasswordWeakNewPassword(t *testing.T) {
	mock := &mockUserLookup{
		users: map[string]*domain.User{
			"alice@local.test": {
				ID:           1,
				Email:        "alice@local.test",
				PasswordHash: hashPasswordForTest("OldPass123!"),
			},
		},
	}
	router := setupAuthRouterWithUserLookup("secret123", "test-jwt-secret", mock)

	body := domain.ChangePasswordRequest{
		Email:           "alice@local.test",
		CurrentPassword: "OldPass123!",
		NewPassword:     "short",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp domain.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "WEAK_PASSWORD", resp.Error.Code)
}

func TestChangePasswordInvalidJSON(t *testing.T) {
	mock := &mockUserLookup{users: map[string]*domain.User{}}
	router := setupAuthRouterWithUserLookup("secret123", "test-jwt-secret", mock)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
