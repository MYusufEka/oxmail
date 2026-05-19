package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MYusufEka/oxmail/internal/domain"
)

func setupAuthRouter(adminPassword, jwtSecret string) *chi.Mux {
	os.Setenv("OXMAIL_ADMIN_PASSWORD", adminPassword)
	os.Setenv("OXMAIL_JWT_SECRET", jwtSecret)

	router := chi.NewRouter()
	handler := NewAuthHandler(jwtSecret, adminPassword)
	handler.RegisterRoutes(router)
	return router
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

	// Make 5 failed attempts
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "attempt %d should be 401", i+1)
	}

	// 6th attempt should be rate limited
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
