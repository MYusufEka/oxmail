package api_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/MYusufEka/oxmail/internal/api"
	apiMiddleware "github.com/MYusufEka/oxmail/internal/api/middleware"
	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type signatureTestFixture struct {
	db      *database.DB
	handler *api.SignatureHandler
}

func setupSignatureHandlerTest(t *testing.T) *signatureTestFixture {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	svc := domain.NewSignatureService(db.Conn)
	return &signatureTestFixture{
		db:      db,
		handler: api.NewSignatureHandler(svc),
	}
}

func (fixture *signatureTestFixture) Router() *chi.Mux {
	return fixture.handler.Router()
}

func TestSignatureHandler_GetEmpty(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/mail/signature/alice@example.com", nil)
	rec := httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var signature domain.Signature
	err := json.NewDecoder(rec.Body).Decode(&signature)
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", signature.Email)
	assert.Equal(t, "", signature.Content)
	assert.False(t, signature.Enabled)
}

func TestSignatureHandler_URLDecodedEmail(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)

	body := `{"content":"<p>Alias</p>","enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/mail/signature/alice%2Btag@example.com", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/mail/signature/alice%2Btag@example.com", nil)
	rec = httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var signature domain.Signature
	err := json.NewDecoder(rec.Body).Decode(&signature)
	require.NoError(t, err)
	assert.Equal(t, "alice+tag@example.com", signature.Email)
	assert.Equal(t, "<p>Alias</p>", signature.Content)
	assert.True(t, signature.Enabled)
}

func TestSignatureHandler_Upsert(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)

	body := `{"content":"<p>Regards</p>","enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/mail/signature/alice@example.com", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var signature domain.Signature
	err := json.NewDecoder(rec.Body).Decode(&signature)
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", signature.Email)
	assert.Equal(t, "<p>Regards</p>", signature.Content)
	assert.True(t, signature.Enabled)
}

func TestSignatureHandler_InvalidJSON(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/mail/signature/alice@example.com", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid_request")
}

func TestSignatureHandler_UpsertReplacesExistingRow(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)

	requests := []string{
		`{"content":"<p>First</p>","enabled":true}`,
		`{"content":"<p>Second</p>","enabled":false}`,
	}
	for _, body := range requests {
		req := httptest.NewRequest(http.MethodPost, "/api/mail/signature/alice@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		fixture.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}

	var signatureCount int
	err := fixture.db.Conn.QueryRow(`SELECT COUNT(*) FROM signatures WHERE user_email = ?`, "alice@example.com").Scan(&signatureCount)
	require.NoError(t, err)
	assert.Equal(t, 1, signatureCount)

	req := httptest.NewRequest(http.MethodGet, "/api/mail/signature/alice@example.com", nil)
	rec := httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	var signature domain.Signature
	err = json.NewDecoder(rec.Body).Decode(&signature)
	require.NoError(t, err)
	assert.Equal(t, "<p>Second</p>", signature.Content)
	assert.False(t, signature.Enabled)
}

func TestSignatureHandler_Delete(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)

	body := `{"content":"<p>Bye</p>","enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/mail/signature/alice@example.com", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/mail/signature/alice@example.com", nil)
	rec = httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "", rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/api/mail/signature/alice@example.com", nil)
	rec = httptest.NewRecorder()
	fixture.Router().ServeHTTP(rec, req)

	var signature domain.Signature
	err := json.NewDecoder(rec.Body).Decode(&signature)
	require.NoError(t, err)
	assert.Equal(t, "", signature.Content)
	assert.False(t, signature.Enabled)
}

func TestSignatureHandler_ProtectedRouteRequiresAuth(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)
	jwtSecret := "signature-test-secret"
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(apiMiddleware.JWTAuth(jwtSecret))
		fixture.handler.RegisterRoutes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/mail/signature/alice@example.com", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSignatureHandler_ProtectedRouteAcceptsBearer(t *testing.T) {
	fixture := setupSignatureHandlerTest(t)
	jwtSecret := "signature-test-secret"
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(apiMiddleware.JWTAuth(jwtSecret))
		fixture.handler.RegisterRoutes(r)
	})
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "admin@example.com",
		"role":  "admin",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
		"iat":   jwt.NewNumericDate(time.Now()),
	})
	signedToken, err := token.SignedString([]byte(jwtSecret))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/mail/signature/alice@example.com", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
