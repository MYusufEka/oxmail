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

func setupBouncesTest(t *testing.T) (*api.BouncesHandler, *domain.BounceService) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	svc := domain.NewBounceService(db.Conn)
	handler := api.NewBouncesHandler(svc)
	return handler, svc
}

func TestBouncesHandler_List(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/bounces", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Bounces []domain.Bounce `json:"bounces"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Empty(t, resp.Bounces)
	})

	t.Run("returns bounces with data", func(t *testing.T) {
		h, svc := setupBouncesTest(t)
		_, err := svc.RecordBounce(context.Background(), "user@example.com", "sender@example.com", "Test", "hard", "User unknown")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/bounces", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Bounces []domain.Bounce `json:"bounces"`
		}
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Bounces, 1)
		assert.Equal(t, "user@example.com", resp.Bounces[0].Recipient)
		assert.Equal(t, "hard", resp.Bounces[0].BounceType)
	})

	t.Run("filters by recipient", func(t *testing.T) {
		h, svc := setupBouncesTest(t)
		svc.RecordBounce(context.Background(), "alice@example.com", "s@e.com", "Subj", "hard", "err")
		svc.RecordBounce(context.Background(), "bob@example.com", "s@e.com", "Subj", "soft", "err")

		req := httptest.NewRequest(http.MethodGet, "/api/bounces?recipient=alice@example.com", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Bounces []domain.Bounce `json:"bounces"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Bounces, 1)
		assert.Equal(t, "alice@example.com", resp.Bounces[0].Recipient)
	})

	t.Run("handles limit and offset", func(t *testing.T) {
		h, svc := setupBouncesTest(t)
		for i := 0; i < 5; i++ {
			svc.RecordBounce(context.Background(), fmt.Sprintf("u%d@e.com", i), "s@e.com", "Subj", "hard", "err")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/bounces?limit=2&offset=0", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Bounces []domain.Bounce `json:"bounces"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp.Bounces, 2)
	})

	t.Run("returns 400 for negative limit", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/bounces?limit=-1", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for invalid limit", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/bounces?limit=abc", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for negative offset", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/bounces?offset=-1", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestBouncesHandler_Get(t *testing.T) {
	t.Run("returns existing bounce", func(t *testing.T) {
		h, svc := setupBouncesTest(t)
		bounce, err := svc.RecordBounce(context.Background(), "get@example.com", "s@e.com", "Subj", "hard", "Err msg")
		require.NoError(t, err)

		path := fmt.Sprintf("/api/bounces/%d", bounce.ID)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result domain.Bounce
		err = json.NewDecoder(rec.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, bounce.ID, result.ID)
		assert.Equal(t, "get@example.com", result.Recipient)
	})

	t.Run("returns 404 for non-existent bounce", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/bounces/99999", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 400 for invalid ID", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/bounces/abc", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestBouncesHandler_Delete(t *testing.T) {
	t.Run("deletes existing bounce", func(t *testing.T) {
		h, svc := setupBouncesTest(t)
		bounce, err := svc.RecordBounce(context.Background(), "del@example.com", "s@e.com", "Subj", "soft", "Err")
		require.NoError(t, err)

		path := fmt.Sprintf("/api/bounces/%d", bounce.ID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "deleted", resp["status"])
	})

	t.Run("returns 404 for non-existent bounce", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/bounces/99999", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 400 for invalid ID", func(t *testing.T) {
		h, _ := setupBouncesTest(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/bounces/abc", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
