package api_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStatsTest(t *testing.T) (*api.StatsHandler, *domain.StatsService) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	svc := domain.NewStatsService(db.Conn)
	handler := api.NewStatsHandler(svc)
	return handler, svc
}

func TestStatsHandler_GetStats(t *testing.T) {
	t.Run("returns empty stats", func(t *testing.T) {
		h, _ := setupStatsTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var stats []domain.DailyStat
		err := json.NewDecoder(rec.Body).Decode(&stats)
		require.NoError(t, err)
		assert.Empty(t, stats)
	})

	t.Run("returns stats with data", func(t *testing.T) {
		h, svc := setupStatsTest(t)
		err := svc.IncrementSent(context.Background())
		require.NoError(t, err)
		err = svc.IncrementReceived(context.Background())
		require.NoError(t, err)
		err = svc.IncrementBounced(context.Background())
		require.NoError(t, err)
		err = svc.IncrementSpamCaught(context.Background())
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var stats []domain.DailyStat
		err = json.NewDecoder(rec.Body).Decode(&stats)
		require.NoError(t, err)
		require.Len(t, stats, 1)
		statDate, err := time.Parse(time.RFC3339, stats[0].Date)
		require.NoError(t, err)
		assert.Equal(t, time.Now().UTC().Format(time.DateOnly), statDate.Format(time.DateOnly))
		assert.Equal(t, int64(1), stats[0].Sent)
		assert.Equal(t, int64(1), stats[0].Received)
		assert.Equal(t, int64(1), stats[0].Bounced)
		assert.Equal(t, int64(1), stats[0].SpamCaught)
	})

	t.Run("handles days parameter", func(t *testing.T) {
		h, svc := setupStatsTest(t)
		err := svc.IncrementSent(context.Background())
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/stats?days=7", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var stats []domain.DailyStat
		err = json.NewDecoder(rec.Body).Decode(&stats)
		require.NoError(t, err)
		assert.NotEmpty(t, stats)
	})

	t.Run("returns 400 for invalid days", func(t *testing.T) {
		h, _ := setupStatsTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/stats?days=0", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp domain.ErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "invalid_days", resp.Error.Code)
	})

	t.Run("returns 400 for negative days", func(t *testing.T) {
		h, _ := setupStatsTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/stats?days=-1", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp domain.ErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "invalid_days", resp.Error.Code)
	})

	t.Run("returns 400 for non-numeric days", func(t *testing.T) {
		h, _ := setupStatsTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/stats?days=abc", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp domain.ErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "invalid_days", resp.Error.Code)
	})
}

func TestStatsHandler_GetSummary(t *testing.T) {
	t.Run("returns empty summary", func(t *testing.T) {
		h, _ := setupStatsTest(t)

		req := httptest.NewRequest(http.MethodGet, "/api/stats/summary", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var summary domain.StatSummary
		err := json.NewDecoder(rec.Body).Decode(&summary)
		require.NoError(t, err)
		assert.Equal(t, int64(0), summary.TotalSent)
	})

	t.Run("returns summary with data", func(t *testing.T) {
		h, svc := setupStatsTest(t)
		err := svc.IncrementSent(context.Background())
		require.NoError(t, err)
		err = svc.IncrementSent(context.Background())
		require.NoError(t, err)
		err = svc.IncrementBounced(context.Background())
		require.NoError(t, err)
		err = svc.IncrementSpamCaught(context.Background())
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/stats/summary", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var summary domain.StatSummary
		err = json.NewDecoder(rec.Body).Decode(&summary)
		require.NoError(t, err)
		assert.Equal(t, int64(2), summary.TotalSent)
		assert.Equal(t, int64(1), summary.TotalBounced)
		assert.Equal(t, int64(1), summary.TotalSpamCaught)
		assert.Equal(t, int64(0), summary.TotalReceived)
	})
}
