package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	srv := api.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	// With real checkers that can't connect, status will be unhealthy (503)
	// Just verify the response format is correct
	var body health.Result
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body.Status)
	assert.Equal(t, "0.1.0", body.Version)
	assert.NotEmpty(t, body.Uptime)
	assert.Len(t, body.Services, 4)
}

func TestHealthEndpoint_ContentType(t *testing.T) {
	srv := api.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
