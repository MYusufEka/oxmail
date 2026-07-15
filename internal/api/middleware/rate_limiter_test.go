package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLimiter(t *testing.T, limit int) (*RateLimiter, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(client, RateLimitConfig{
		Name:   "test",
		Limit:  limit,
		Window: time.Minute,
	})
	return rl, mr
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl, _ := newTestLimiter(t, 3)

	for i := 0; i < 3; i++ {
		allowed, _, err := rl.Allow(t.Context(), "client1")
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_BlocksAtLimit(t *testing.T) {
	rl, _ := newTestLimiter(t, 2)

	for i := 0; i < 2; i++ {
		allowed, _, err := rl.Allow(t.Context(), "client1")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	allowed, retryAfter, err := rl.Allow(t.Context(), "client1")
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.GreaterOrEqual(t, retryAfter, time.Duration(0))
}

func TestRateLimiter_SeparateKeysPerClient(t *testing.T) {
	rl, _ := newTestLimiter(t, 1)

	allowed1, _, err := rl.Allow(t.Context(), "client1")
	require.NoError(t, err)
	assert.True(t, allowed1)

	allowed2, _, err := rl.Allow(t.Context(), "client2")
	require.NoError(t, err)
	assert.True(t, allowed2)

	blocked, _, err := rl.Allow(t.Context(), "client1")
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl, mr := newTestLimiter(t, 1)

	allowed, _, err := rl.Allow(t.Context(), "client1")
	require.NoError(t, err)
	assert.True(t, allowed)

	blocked, _, err := rl.Allow(t.Context(), "client1")
	require.NoError(t, err)
	assert.False(t, blocked)

	mr.FastForward(2 * time.Minute)

	allowed2, _, err := rl.Allow(t.Context(), "client1")
	require.NoError(t, err)
	assert.True(t, allowed2)
}

func TestRateLimiter_Middleware_Returns429WithRetryAfter(t *testing.T) {
	rl, _ := newTestLimiter(t, 1)

	router := chi.NewRouter()
	router.With(rl.Middleware()).Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	makeRequest := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	first := makeRequest()
	assert.Equal(t, http.StatusOK, first.Code)

	second := makeRequest()
	assert.Equal(t, http.StatusTooManyRequests, second.Code)

	retryAfter := second.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfter)
	secs, err := strconv.Atoi(retryAfter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, secs, 1)
}

func TestRateLimiter_Middleware_FailOpenWhenRedisDown(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(client, RateLimitConfig{
		Name:   "test",
		Limit:  1,
		Window: time.Minute,
	})

	router := chi.NewRouter()
	router.With(rl.Middleware()).Post("/api/mail/send", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mr.Close()

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/mail/send", nil)
		req.RemoteAddr = "10.0.0.2:9999"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "fail-open: request %d must pass when redis down", i+1)
	}
}

func TestEnvRateLimitSend_Default(t *testing.T) {
	t.Setenv("OXMAIL_RATE_LIMIT_SEND", "")
	assert.Equal(t, DefaultRateLimitSend, EnvRateLimitSend())
}

func TestEnvRateLimitSend_Custom(t *testing.T) {
	t.Setenv("OXMAIL_RATE_LIMIT_SEND", "42")
	assert.Equal(t, 42, EnvRateLimitSend())
}

func TestEnvRateLimitLogin_Default(t *testing.T) {
	t.Setenv("OXMAIL_RATE_LIMIT_LOGIN", "")
	assert.Equal(t, DefaultRateLimitLogin, EnvRateLimitLogin())
}

func TestEnvRateLimitLogin_Invalid(t *testing.T) {
	t.Setenv("OXMAIL_RATE_LIMIT_LOGIN", "notanumber")
	assert.Equal(t, DefaultRateLimitLogin, EnvRateLimitLogin())
}

func TestClientIPFromRequest_ParsesRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:4321"
	assert.Equal(t, "192.168.1.1", clientIPFromRequest(req))
}

func TestClientIPFromRequest_BareAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"
	assert.Equal(t, "192.168.1.1", clientIPFromRequest(req))
}
