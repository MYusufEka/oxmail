package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type authContextKey struct{}

func UserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(authContextKey{}).(string)
	return email, ok && email != ""
}

// JWTAuth returns middleware that validates JWT Bearer tokens.
// In dev mode (OXMAIL_MODE=dev), all requests are allowed through.
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if os.Getenv("OXMAIL_MODE") == "dev" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString, ok := tokenFromCookie(r)
			if !ok {
				var authErr *authError
				tokenString, authErr = tokenFromBearerHeader(r.Header.Get("Authorization"))
				if authErr != nil {
					writeAuthError(w, http.StatusUnauthorized, authErr.code, authErr.message)
					return
				}
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				writeAuthError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
				return
			}

			ctx := context.WithValue(r.Context(), authContextKey{}, emailFromClaims(token.Claims))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type authError struct {
	code    string
	message string
}

func tokenFromCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", false
	}

	tokenString := strings.TrimSpace(cookie.Value)
	return tokenString, tokenString != ""
}

func tokenFromBearerHeader(authHeader string) (string, *authError) {
	if authHeader == "" {
		return "", &authError{code: "MISSING_TOKEN", message: "Authorization header required"}
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", &authError{code: "INVALID_TOKEN", message: "Invalid authorization format"}
	}

	return parts[1], nil
}

func emailFromClaims(claims jwt.Claims) string {
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	if email, ok := mapClaims["email"].(string); ok && email != "" {
		return email
	}

	if subject, ok := mapClaims["sub"].(string); ok {
		return subject
	}

	return ""
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
