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
	router.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
		email, ok := UserEmailFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(email))
	})
	return router
}

func generateValidToken(secret string) string {
	return generateToken(secret, jwt.MapClaims{
		"sub": "admin",
		"exp": float64(9999999999),
		"iat": float64(1000000000),
	})
}

func generateToken(secret string, claims jwt.MapClaims) string {
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

func TestJWTAuth_CookieToken(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	secret := "test-secret"
	router := testRouter(secret)
	token := generateToken(secret, jwt.MapClaims{
		"email": "admin@example.com",
		"sub":   "admin",
		"exp":   float64(9999999999),
		"iat":   float64(1000000000),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "admin@example.com", rec.Body.String())
}

func TestJWTAuth_CookiePrecedesHeader(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	secret := "test-secret"
	router := testRouter(secret)
	cookieToken := generateToken(secret, jwt.MapClaims{
		"sub": "cookie-user@example.com",
		"exp": float64(9999999999),
		"iat": float64(1000000000),
	})
	headerToken := generateToken(secret, jwt.MapClaims{
		"sub": "header-user@example.com",
		"exp": float64(9999999999),
		"iat": float64(1000000000),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: cookieToken})
	req.Header.Set("Authorization", "Bearer "+headerToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "cookie-user@example.com", rec.Body.String())
}

func TestJWTAuth_NeitherCookieNorHeader(t *testing.T) {
	os.Setenv("OXMAIL_MODE", "production")
	defer os.Unsetenv("OXMAIL_MODE")

	router := testRouter("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: ""})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"MISSING_TOKEN","message":"Authorization header required"}}`, rec.Body.String())
}
