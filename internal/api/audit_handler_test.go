package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type auditListResponse struct {
	Entries []domain.AuditEntry `json:"entries"`
	Total   int                 `json:"total"`
}

func setupAuditTest(t *testing.T) (*api.AuditHandler, *domain.AuditService) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	svc := domain.NewAuditService(db.Conn)
	handler := api.NewAuditHandler(svc)
	return handler, svc
}

func TestAuditHandler_List(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		h, _ := setupAuditTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/audit", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Empty(t, resp.Entries)
		assert.Equal(t, 0, resp.Total)
	})

	t.Run("returns entries with data", func(t *testing.T) {
		h, svc := setupAuditTest(t)
		ctx := context.Background()
		err := svc.Log(ctx, "admin@test.com", "create", "domain", "domain-1", `{"name":"test.com"}`)
		require.NoError(t, err)
		err = svc.Log(ctx, "admin@test.com", "delete", "user", "user-2", `{"email":"u@t.com"}`)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/audit", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Entries, 2)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, "admin@test.com", resp.Entries[0].Actor)
		assert.Equal(t, "delete", resp.Entries[0].Action)
		assert.Equal(t, "user", resp.Entries[0].TargetType)
		assert.Equal(t, "user-2", resp.Entries[0].TargetID)
		assert.JSONEq(t, `{"email":"u@t.com"}`, resp.Entries[0].Detail)
	})

	t.Run("filters by actor", func(t *testing.T) {
		h, svc := setupAuditTest(t)
		svc.Log(context.Background(), "alice", "create", "domain", "1", `{}`)
		svc.Log(context.Background(), "bob", "delete", "user", "2", `{}`)

		req := httptest.NewRequest(http.MethodGet, "/api/audit?actor=alice", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		assert.Equal(t, 1, resp.Total)
		assert.Equal(t, "alice", resp.Entries[0].Actor)
	})

	t.Run("filters by action", func(t *testing.T) {
		h, svc := setupAuditTest(t)
		svc.Log(context.Background(), "admin", "create", "domain", "1", `{}`)
		svc.Log(context.Background(), "admin", "delete", "user", "2", `{}`)
		svc.Log(context.Background(), "admin", "create", "alias", "3", `{}`)

		req := httptest.NewRequest(http.MethodGet, "/api/audit?action=create", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Entries, 2)
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("filters by both actor and action", func(t *testing.T) {
		h, svc := setupAuditTest(t)
		svc.Log(context.Background(), "admin", "create", "domain", "1", `{}`)
		svc.Log(context.Background(), "admin", "delete", "user", "2", `{}`)
		svc.Log(context.Background(), "ops", "create", "alias", "3", `{}`)

		req := httptest.NewRequest(http.MethodGet, "/api/audit?actor=admin&action=delete", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		assert.Equal(t, 1, resp.Total)
		assert.Equal(t, "admin", resp.Entries[0].Actor)
		assert.Equal(t, "delete", resp.Entries[0].Action)
	})

	t.Run("handles pagination", func(t *testing.T) {
		h, svc := setupAuditTest(t)
		ctx := context.Background()
		for entryNumber := 0; entryNumber < 5; entryNumber++ {
			err := svc.Log(ctx, "admin", "create", "domain", fmt.Sprintf("domain-%d", entryNumber), `{}`)
			require.NoError(t, err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/audit?limit=2&offset=0", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp.Entries, 2)
		assert.Equal(t, 5, resp.Total)
	})

	t.Run("paginates filtered action results", func(t *testing.T) {
		h, svc := setupAuditTest(t)
		ctx := context.Background()
		for entryNumber := 0; entryNumber < 3; entryNumber++ {
			err := svc.Log(ctx, "admin", "create", "domain", fmt.Sprintf("domain-%d", entryNumber), `{}`)
			require.NoError(t, err)
		}
		err := svc.Log(ctx, "admin", "delete", "domain", "domain-3", `{}`)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/audit?action=create&limit=2&offset=1", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Entries, 2)
		assert.Equal(t, 3, resp.Total)
		for _, auditEntry := range resp.Entries {
			assert.Equal(t, "create", auditEntry.Action)
		}
	})

	t.Run("invalid pagination falls back to defaults", func(t *testing.T) {
		h, svc := setupAuditTest(t)
		ctx := context.Background()
		for entryNumber := 0; entryNumber < 3; entryNumber++ {
			err := svc.Log(ctx, "admin", "create", "domain", fmt.Sprintf("domain-%d", entryNumber), `{}`)
			require.NoError(t, err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/audit?limit=-2&offset=-1", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp auditListResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp.Entries, 3)
		assert.Equal(t, 3, resp.Total)
	})
}
