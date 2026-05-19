package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func testRouter(jwtSecret string) *chi.Mux {
	router := chi.NewRouter()
	router.Use(JWTAuth(jwtSecret))
	router.Get("/api/domains", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
	return router
}

func generateValidToken(secret string) string {
	claims := jwt.MapClaims{
		"sub": "admin",
		"exp": float64(9999999999),
		"iat": float64(1000000000),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestJWTAuthRejectsUnauthenticated(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	router := testRouter("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuthAcceptsValidToken(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	secret := "test-secret"
	router := testRouter(secret)
	token := generateValidToken(secret)

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTAuthRejectsInvalidToken(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	router := testRouter("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuthRejectsExpiredToken(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	secret := "test-secret"
	router := testRouter(secret)

	claims := jwt.MapClaims{
		"sub": "admin",
		"exp": float64(1000000000), // expired long ago
		"iat": float64(999999999),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuthSkipsInDevMode(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "dev")
	defer os.Unsetenv("OXMAIL_MODE")

	router := testRouter("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTAuthRejectsWrongSignature(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	router := testRouter("test-secret")
	token := generateValidToken("different-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
