package auth

import (
	"context"
	"net/http"
	"strings"
)

// UserContextKey is the key used to store user in request context
type UserContextKey struct{}

// AuthMiddleware creates middleware that validates JWT tokens and adds user to context
func AuthMiddleware(authService *AuthService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain endpoints
			if shouldSkipAuth(r.URL.Path) {
				next(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// For endpoints that require authentication, return 401
				if requiresAuth(r.URL.Path) {
					http.Error(w, "Authorization header required", http.StatusUnauthorized)
					return
				}
				// For optional auth endpoints, continue without user
				next(w, r)
				return
			}

			token, err := ExtractTokenFromHeader(authHeader)
			if err != nil {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

			// Validate token
			user, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), UserContextKey{}, user)
			next(w, r.WithContext(ctx))
		}
	}
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey{}).(*User)
	return user, ok
}

// shouldSkipAuth returns true if the endpoint should skip authentication
func shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/health",
		"/api/auth/login",
		"/api/auth/register",
		"/api/auth/temp-user",
		"/api/auth/verify-email",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// requiresAuth returns true if the endpoint requires authentication
func requiresAuth(path string) bool {
	// All API endpoints except auth endpoints require authentication
	return strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/api/auth/")
}
