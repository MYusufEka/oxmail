package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDKIMRouter(t *testing.T) (http.Handler, *domain.DKIMService) {
	t.Helper()
	tempDir := t.TempDir()
	dkimService := domain.NewDKIMService(nil, tempDir)

	r := chi.NewRouter()
	api.RegisterDKIMRoutes(r, dkimService)
	return r, dkimService
}

func TestDKIMHandler_Generate(t *testing.T) {
	router, _ := setupDKIMRouter(t)

	t.Run("creates DKIM key and returns 201", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/domains/local.test/dkim", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var body domain.DKIMKey
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "local.test", body.Domain)
		assert.Equal(t, "default", body.Selector)
		assert.NotEmpty(t, body.PublicKey)
		assert.Contains(t, body.DNSRecord, "v=DKIM1; k=rsa; p=")
		assert.False(t, body.CreatedAt.IsZero())
	})

	t.Run("returns application/json content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/domains/example.com/dkim", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})
}

func TestDKIMHandler_Get(t *testing.T) {
	router, dkimService := setupDKIMRouter(t)

	t.Run("returns 200 with key info", func(t *testing.T) {
		_, err := dkimService.Generate("local.test", "default")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/domains/local.test/dkim", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body domain.DKIMKey
		err = json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "local.test", body.Domain)
		assert.Equal(t, "default", body.Selector)
		assert.NotEmpty(t, body.PublicKey)
		assert.Contains(t, body.DNSRecord, "v=DKIM1; k=rsa; p=")
	})

	t.Run("returns 404 for non-existent domain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/domains/nonexistent.com/dkim", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body domain.ErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "not_found", body.Error.Code)
	})
}

func TestDKIMHandler_Delete(t *testing.T) {
	router, dkimService := setupDKIMRouter(t)

	t.Run("deletes existing key and returns 200", func(t *testing.T) {
		_, err := dkimService.Generate("delete-test.com", "default")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/domains/delete-test.com/dkim", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]string
		err = json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "deleted", body["status"])
	})

	t.Run("returns 404 for non-existent key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/domains/nonexistent.com/dkim", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDKIMHandler_Rotate(t *testing.T) {
	router, dkimService := setupDKIMRouter(t)

	t.Run("rotates existing key and returns 200 with warning", func(t *testing.T) {
		_, err := dkimService.Generate("rotate-test.com", "default")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/domains/rotate-test.com/dkim/rotate", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)

		assert.NotEmpty(t, body["publicKey"])
		assert.Contains(t, body["dnsRecord"], "v=DKIM1; k=rsa; p=")
		assert.NotEmpty(t, body["message"])
	})

	t.Run("returns 404 when no key to rotate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/domains/no-key.com/dkim/rotate", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
