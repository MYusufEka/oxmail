package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MYusufEka/oxmail/internal/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChecker struct {
	healthy   bool
	latencyMs int
}

func (m *mockChecker) Check(_ context.Context) (bool, int) {
	return m.healthy, m.latencyMs
}

func TestHealthHandler_AllHealthy(t *testing.T) {
	checkers := map[string]health.ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 2},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 5},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := health.NewService(checkers, "0.1.0")
	handler := newHealthHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var result health.Result
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, "0.1.0", result.Version)
	assert.NotEmpty(t, result.Uptime)
	require.Len(t, result.Services, 4)
}

func TestHealthHandler_Degraded(t *testing.T) {
	checkers := map[string]health.ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 2},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: false, latencyMs: 0},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := health.NewService(checkers, "0.1.0")
	handler := newHealthHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result health.Result
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "degraded", result.Status)
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	checkers := map[string]health.ServiceChecker{
		"postfix": &mockChecker{healthy: false, latencyMs: 0},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 5},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := health.NewService(checkers, "0.1.0")
	handler := newHealthHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var result health.Result
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "unhealthy", result.Status)
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	checkers := map[string]health.ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 12},
		"dovecot": &mockChecker{healthy: true, latencyMs: 3},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 45},
		"redis":   &mockChecker{healthy: true, latencyMs: 7},
	}

	svc := health.NewService(checkers, "0.1.0")
	handler := newHealthHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var result health.Result
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	// Verify all required fields present
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.Version)
	assert.NotEmpty(t, result.Uptime)
	assert.NotEmpty(t, result.Services)

	// Verify each service has required fields
	for _, svc := range result.Services {
		assert.NotEmpty(t, svc.Name)
		assert.NotEmpty(t, svc.Status)
		assert.GreaterOrEqual(t, svc.LatencyMs, 0)
	}
}
