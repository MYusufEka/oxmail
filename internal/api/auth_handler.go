package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	apiMiddleware "github.com/MYusufEka/oxmail/internal/api/middleware"
	"github.com/MYusufEka/oxmail/internal/domain"
)

// UserLookup fetches a user by email for password verification.
type UserLookup interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, id int64, req domain.UpdateUserRequest) (*domain.User, error)
	List(ctx context.Context, params domain.UserListParams) ([]domain.User, int, error)
}

// PasswordChangeMailConfigApplier regenerates Dovecot passdb after password change.
type PasswordChangeMailConfigApplier interface {
	ApplyUserConfig(users []domain.User) error
}

type AuthHandler struct {
	jwtSecret     string
	adminPassword string
	rateLimiter   *loginRateLimiter
	userLookup    UserLookup
	mailConfig    PasswordChangeMailConfigApplier
}

type loginRateLimiter struct {
	mu          sync.Mutex
	attempts    map[string][]time.Time
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

func NewAuthHandler(jwtSecret, adminPassword string, userLookup UserLookup, mailConfig PasswordChangeMailConfigApplier) *AuthHandler {
	return &AuthHandler{
		jwtSecret:     jwtSecret,
		adminPassword: adminPassword,
		rateLimiter:   newLoginRateLimiter(5, time.Minute),
		userLookup:    userLookup,
		mailConfig:    mailConfig,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/auth/login", h.handleLogin)
	r.Post("/api/auth/logout", h.handleLogout)
	if h.userLookup != nil {
		r.Post("/api/auth/change-password", h.handleChangePassword)
	}
}

func (h *AuthHandler) RegisterProtectedRoutes(r chi.Router) {
	r.Get("/api/auth/me", h.handleMe)
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

	authenticatedEmail, role, mustChangePassword, ok := h.authenticateLogin(r.Context(), req)
	if !ok {
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
		"sub":   authenticatedEmail,
		"email": authenticatedEmail,
		"role":  role,
		"exp":   jwt.NewNumericDate(expiresAt),
		"iat":   jwt.NewNumericDate(time.Now()),
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

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    signed,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":             "ok",
		"email":              authenticatedEmail,
		"role":               role,
		"mustChangePassword": mustChangePassword,
	})
}

func (h *AuthHandler) authenticateLogin(ctx context.Context, req domain.LoginRequest) (string, string, bool, bool) {
	if req.Email == "admin" && req.Password == h.adminPassword {
		mustChangePassword := false
		if h.userLookup != nil {
			if user, err := h.userLookup.GetByEmail(ctx, "admin"); err == nil {
				mustChangePassword = user.MustChangePassword
			}
		}
		return "admin", "admin", mustChangePassword, true
	}

	if h.userLookup == nil {
		return "", "", false, false
	}

	user, err := h.userLookup.GetByEmail(ctx, req.Email)
	if err != nil || !user.Active {
		return "", "", false, false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", "", false, false
	}

	return user.Email, "user", user.MustChangePassword, true
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	email, ok := apiMiddleware.UserEmailFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
			Error: domain.ErrorDetail{
				Code:    "UNAUTHENTICATED",
				Message: "Authentication required",
			},
		})
		return
	}

	mustChangePassword := false
	role := "admin"
	if h.userLookup != nil {
		if user, err := h.userLookup.GetByEmail(r.Context(), email); err == nil {
			mustChangePassword = user.MustChangePassword
			role = "user"
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"email":              email,
		"role":               role,
		"mustChangePassword": mustChangePassword,
	})
}

func (h *AuthHandler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req domain.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.ErrorDetail{Code: "INVALID_BODY", Message: "Invalid request body"},
		})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.CurrentPassword == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.ErrorDetail{Code: "INVALID_BODY", Message: "Email, currentPassword, and newPassword are required"},
		})
		return
	}

	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.ErrorDetail{Code: "WEAK_PASSWORD", Message: "New password must be at least 8 characters"},
		})
		return
	}

	user, err := h.userLookup.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error: domain.ErrorDetail{Code: "USER_NOT_FOUND", Message: "User not found"},
			})
			return
		}
		slog.Error("failed to lookup user for password change", "error", err)
		writeJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Internal server error"},
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
			Error: domain.ErrorDetail{Code: "INVALID_CREDENTIALS", Message: "Current password is incorrect"},
		})
		return
	}

	if _, err := h.userLookup.Update(r.Context(), user.ID, domain.UpdateUserRequest{Password: &req.NewPassword}); err != nil {
		slog.Error("failed to update password", "error", err, "email", req.Email)
		writeJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to update password"},
		})
		return
	}

	if h.mailConfig != nil {
		users, _, err := h.userLookup.List(r.Context(), domain.UserListParams{Page: 1, Limit: 10000})
		if err != nil {
			slog.Error("failed to list users for dovecot config after password change", "error", err)
		} else if err := h.mailConfig.ApplyUserConfig(users); err != nil {
			slog.Error("failed to apply dovecot config after password change", "error", err)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
}

func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
