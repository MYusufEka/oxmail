package api

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	jwtSecret     string
	adminPassword string
	rateLimiter   *loginRateLimiter
}

type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

func newLoginRateLimiter(maxAttempts int, window time.Duration) *loginRateLimiter {
	return &loginRateLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

func (rl *loginRateLimiter) isRateLimited(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	attempts := rl.attempts[ip]
	valid := attempts[:0]
	for _, t := range attempts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.attempts[ip] = valid

	return len(valid) >= rl.maxAttempts
}

func (rl *loginRateLimiter) recordAttempt(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.attempts[ip] = append(rl.attempts[ip], time.Now())
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(jwtSecret, adminPassword string) *AuthHandler {
	return &AuthHandler{
		jwtSecret:     jwtSecret,
		adminPassword: adminPassword,
		rateLimiter:   newLoginRateLimiter(5, time.Minute),
	}
}

// RegisterRoutes registers auth routes on the router.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/auth/login", h.handleLogin)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := extractIP(r)

	if h.rateLimiter.isRateLimited(ip) {
		writeJSON(w, http.StatusTooManyRequests, domain.ErrorResponse{
			Error: domain.ErrorDetail{
				Code:    "RATE_LIMITED",
				Message: "Too many login attempts. Try again later.",
			},
		})
		return
	}

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.ErrorDetail{
				Code:    "INVALID_BODY",
				Message: "Invalid request body",
			},
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.ErrorDetail{
				Code:    "INVALID_BODY",
				Message: "Email and password are required",
			},
		})
		return
	}

	if req.Email != "admin" || req.Password != h.adminPassword {
		h.rateLimiter.recordAttempt(ip)
		writeJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
			Error: domain.ErrorDetail{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid email or password",
			},
		})
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"sub": "admin",
		"exp": jwt.NewNumericDate(expiresAt),
		"iat": jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.ErrorDetail{
				Code:    "TOKEN_ERROR",
				Message: "Failed to generate token",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, domain.LoginResponse{
		Token:     signed,
		ExpiresAt: expiresAt,
	})
}

func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}


