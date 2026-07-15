package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultRateLimitSend is requests per minute for POST /api/mail/send.
	DefaultRateLimitSend = 20
	// DefaultRateLimitLogin is requests per minute for POST /api/auth/login.
	DefaultRateLimitLogin = 10
	// DefaultRedisURL is the in-compose Redis address.
	DefaultRedisURL = "redis://redis:6379"
	// DefaultRateLimitWindow is the sliding window duration.
	DefaultRateLimitWindow = time.Minute
)

// slidingWindowScript atomically counts requests in a sliding window and records
// the current request when under the limit.
// KEYS[1] = rate limit key
// ARGV[1] = now (unix ms)
// ARGV[2] = window start (unix ms)
// ARGV[3] = limit
// ARGV[4] = member id
// ARGV[5] = key TTL seconds
// Returns: {allowed (0|1), count, retry_after_ms}
var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window_start = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]
local ttl = tonumber(ARGV[5])

redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
local count = redis.call('ZCARD', key)

if count >= limit then
  local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
  local retry_after_ms = 0
  if oldest[2] then
    retry_after_ms = math.max(0, tonumber(oldest[2]) + (now - window_start) - now)
  end
  return {0, count, retry_after_ms}
end

redis.call('ZADD', key, now, member)
redis.call('EXPIRE', key, ttl)
return {1, count + 1, 0}
`)

// RateLimiter enforces per-client sliding-window limits backed by Redis.
// Fail-open: if Redis is unavailable, the request is allowed and a warning is logged.
type RateLimiter struct {
	client *redis.Client
	logger *slog.Logger
	limit  int
	window time.Duration
	name   string
}

// RateLimitConfig configures a named rate limiter.
type RateLimitConfig struct {
	// Name is used in Redis keys (e.g. "login", "send").
	Name string
	// Limit is max requests per window (defaults applied by caller).
	Limit int
	// Window is the sliding window duration (default 1 minute).
	Window time.Duration
	// Logger receives fail-open warnings; optional.
	Logger *slog.Logger
}

// NewRedisClient parses OXMAIL_REDIS_URL (or DefaultRedisURL) and returns a client.
func NewRedisClient() (*redis.Client, error) {
	redisURL := os.Getenv("OXMAIL_REDIS_URL")
	if redisURL == "" {
		redisURL = DefaultRedisURL
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse OXMAIL_REDIS_URL: %w", err)
	}
	return redis.NewClient(opts), nil
}

// EnvRateLimitSend returns OXMAIL_RATE_LIMIT_SEND or DefaultRateLimitSend.
func EnvRateLimitSend() int {
	return envInt("OXMAIL_RATE_LIMIT_SEND", DefaultRateLimitSend)
}

// EnvRateLimitLogin returns OXMAIL_RATE_LIMIT_LOGIN or DefaultRateLimitLogin.
func EnvRateLimitLogin() int {
	return envInt("OXMAIL_RATE_LIMIT_LOGIN", DefaultRateLimitLogin)
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

// NewRateLimiter builds a Redis sliding-window limiter for one route family.
func NewRateLimiter(client *redis.Client, cfg RateLimitConfig) *RateLimiter {
	window := cfg.Window
	if window <= 0 {
		window = DefaultRateLimitWindow
	}
	limit := cfg.Limit
	if limit <= 0 {
		limit = 1
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &RateLimiter{
		client: client,
		logger: logger,
		limit:  limit,
		window: window,
		name:   cfg.Name,
	}
}

// Middleware returns chi-compatible middleware that rate-limits by client IP.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rl == nil || rl.client == nil {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := clientIPFromRequest(r)
			allowed, retryAfter, err := rl.Allow(r.Context(), clientIP)
			if err != nil {
				// Fail open: Redis down must not block mail/auth.
				rl.logger.Warn("rate limiter redis unavailable, allowing request",
					"name", rl.name,
					"client_ip", clientIP,
					"error", err,
				)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				secs := int(retryAfter.Seconds())
				if secs < 1 {
					secs = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(secs))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "rate limit exceeded, try again later",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Allow checks and records one request for key within the sliding window.
// Returns allowed, suggested Retry-After duration, and any Redis error.
func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, time.Duration, error) {
	now := time.Now()
	nowMs := now.UnixMilli()
	windowStartMs := now.Add(-rl.window).UnixMilli()
	ttlSec := int(rl.window.Seconds()) + 1
	if ttlSec < 2 {
		ttlSec = 2
	}

	redisKey := fmt.Sprintf("oxmail:rl:%s:%s", rl.name, key)
	member := uuid.NewString()

	result, err := slidingWindowScript.Run(ctx, rl.client, []string{redisKey},
		nowMs, windowStartMs, rl.limit, member, ttlSec,
	).Int64Slice()
	if err != nil {
		return false, 0, err
	}

	if len(result) < 3 {
		return false, 0, fmt.Errorf("unexpected redis script result length %d", len(result))
	}

	allowed := result[0] == 1
	retryAfterMs := result[2]
	if retryAfterMs < 0 {
		retryAfterMs = 0
	}
	return allowed, time.Duration(retryAfterMs) * time.Millisecond, nil
}

func clientIPFromRequest(r *http.Request) string {
	// RealIP middleware rewrites RemoteAddr when present; still parse safely.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
