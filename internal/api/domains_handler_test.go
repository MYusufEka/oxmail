package api_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDomainsTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func newDomainsHandler(t *testing.T) *api.DomainsHandler {
	t.Helper()
	db := setupDomainsTestDB(t)
	svc := domain.NewDomainService(db)
	configPath := filepath.Join(t.TempDir(), "postfix", "virtual_domains")
	return api.NewDomainsHandler(svc, configPath, nil)
}

func TestDomainsHandler_Create(t *testing.T) {
	t.Run("creates domain successfully", func(t *testing.T) {
		h := newDomainsHandler(t)

		body := `{"name":"example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var created domain.Domain
		err := json.NewDecoder(rec.Body).Decode(&created)
		require.NoError(t, err)
		assert.Equal(t, "example.com", created.Name)
		assert.True(t, created.Active)
		assert.NotZero(t, created.ID)
	})

	t.Run("returns 409 for duplicate domain", func(t *testing.T) {
		h := newDomainsHandler(t)

		body := `{"name":"dup.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		req = httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("returns 400 for invalid domain", func(t *testing.T) {
		h := newDomainsHandler(t)

		body := `{"name":"not valid!"}`
		req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		h := newDomainsHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestDomainsHandler_List(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		h := newDomainsHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp api.DomainListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Empty(t, resp.Data)
		assert.Equal(t, 0, resp.Pagination.Total)
	})

	t.Run("returns domains with pagination", func(t *testing.T) {
		h := newDomainsHandler(t)

		// Create domains
		for _, name := range []string{"a.com", "b.com", "c.com"} {
			body := `{"name":"` + name + `"}`
			req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.Router().ServeHTTP(rec, req)
			require.Equal(t, http.StatusCreated, rec.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/domains?page=1&limit=2", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp api.DomainListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, 3, resp.Pagination.Total)
	})
}

func TestDomainsHandler_Get(t *testing.T) {
	t.Run("gets existing domain", func(t *testing.T) {
		h := newDomainsHandler(t)

		body := `{"name":"getme.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		req = httptest.NewRequest(http.MethodGet, "/api/domains/getme.com", nil)
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got domain.Domain
		err := json.NewDecoder(rec.Body).Decode(&got)
		require.NoError(t, err)
		assert.Equal(t, "getme.com", got.Name)
	})

	t.Run("returns 404 for missing domain", func(t *testing.T) {
		h := newDomainsHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/domains/nope.com", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDomainsHandler_Delete(t *testing.T) {
	t.Run("deletes existing domain", func(t *testing.T) {
		h := newDomainsHandler(t)

		body := `{"name":"deleteme.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		req = httptest.NewRequest(http.MethodDelete, "/api/domains/deleteme.com", nil)
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns 404 for missing domain", func(t *testing.T) {
		h := newDomainsHandler(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/domains/ghost.com", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDomainsHandler_ConfigGeneration(t *testing.T) {
	t.Run("generates config file after create", func(t *testing.T) {
		db := setupDomainsTestDB(t)
		svc := domain.NewDomainService(db)
		configPath := filepath.Join(t.TempDir(), "postfix", "virtual_domains")
		h := api.NewDomainsHandler(svc, configPath, nil)

		body := `{"name":"configured.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "configured.com")
	})
}
