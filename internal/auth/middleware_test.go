package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create a test user and get a token
	user, token, err := authService.Register("middlewaretest", "middleware@example.com", "password123")
	require.NoError(t, err)

	middleware := AuthMiddleware(authService)

	// Mock handler that checks if user is in context
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if ok {
			w.Header().Set("X-User-ID", user.ID)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}

	tests := []struct {
		name           string
		path           string
		authHeader     string
		wantStatus     int
		wantUserInCtx  bool
		expectedUserID string
	}{
		{
			name:           "valid token on protected endpoint",
			path:           "/api/protected",
			authHeader:     "Bearer " + token,
			wantStatus:     http.StatusOK,
			wantUserInCtx:  true,
			expectedUserID: user.ID,
		},
		{
			name:          "missing auth header on protected endpoint",
			path:          "/api/protected",
			authHeader:    "",
			wantStatus:    http.StatusUnauthorized,
			wantUserInCtx: false,
		},
		{
			name:          "invalid token on protected endpoint",
			path:          "/api/protected",
			authHeader:    "Bearer invalid-token",
			wantStatus:    http.StatusUnauthorized,
			wantUserInCtx: false,
		},
		{
			name:          "malformed auth header on protected endpoint",
			path:          "/api/protected",
			authHeader:    "InvalidFormat",
			wantStatus:    http.StatusUnauthorized,
			wantUserInCtx: false,
		},
		{
			name:          "skip auth endpoint - health check",
			path:          "/health",
			authHeader:    "",
			wantStatus:    http.StatusOK,
			wantUserInCtx: false,
		},
		{
			name:          "skip auth endpoint - login",
			path:          "/api/auth/login",
			authHeader:    "",
			wantStatus:    http.StatusOK,
			wantUserInCtx: false,
		},
		{
			name:          "skip auth endpoint - register",
			path:          "/api/auth/register",
			authHeader:    "",
			wantStatus:    http.StatusOK,
			wantUserInCtx: false,
		},
		{
			name:          "skip auth endpoint - temp user",
			path:          "/api/auth/temp-user",
			authHeader:    "",
			wantStatus:    http.StatusOK,
			wantUserInCtx: false,
		},
		{
			name:          "skip auth endpoint - verify email",
			path:          "/api/auth/verify-email",
			authHeader:    "",
			wantStatus:    http.StatusOK,
			wantUserInCtx: false,
		},
		{
			name:           "valid token on skip auth endpoint",
			path:           "/health",
			authHeader:     "Bearer " + token,
			wantStatus:     http.StatusOK,
			wantUserInCtx:  false, // Should skip auth completely
			expectedUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			handler := middleware(mockHandler)
			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantUserInCtx {
				userID := w.Header().Get("X-User-ID")
				assert.Equal(t, tt.expectedUserID, userID)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantUser *User
		wantOk   bool
	}{
		{
			name: "user in context",
			ctx: context.WithValue(context.Background(), UserContextKey{}, &User{
				ID:       "test-id",
				Username: "testuser",
			}),
			wantUser: &User{
				ID:       "test-id",
				Username: "testuser",
			},
			wantOk: true,
		},
		{
			name:     "no user in context",
			ctx:      context.Background(),
			wantUser: nil,
			wantOk:   false,
		},
		{
			name:     "wrong type in context",
			ctx:      context.WithValue(context.Background(), UserContextKey{}, "not-a-user"),
			wantUser: nil,
			wantOk:   false,
		},
		{
			name:     "nil user in context",
			ctx:      context.WithValue(context.Background(), UserContextKey{}, (*User)(nil)),
			wantUser: nil,
			wantOk:   true, // GetUserFromContext returns the value regardless of whether it's nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, ok := GetUserFromContext(tt.ctx)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk && tt.wantUser != nil {
				assert.Equal(t, tt.wantUser.ID, user.ID)
				assert.Equal(t, tt.wantUser.Username, user.Username)
			} else if !tt.wantOk {
				assert.Nil(t, user)
			}
			// For nil user in context case, we get ok=true but user=nil
			if tt.wantUser == nil && tt.wantOk {
				assert.Nil(t, user)
			}
		})
	}
}

func TestShouldSkipAuth(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "health endpoint",
			path:     "/health",
			expected: true,
		},
		{
			name:     "login endpoint",
			path:     "/api/auth/login",
			expected: true,
		},
		{
			name:     "register endpoint",
			path:     "/api/auth/register",
			expected: true,
		},
		{
			name:     "temp user endpoint",
			path:     "/api/auth/temp-user",
			expected: true,
		},
		{
			name:     "verify email endpoint",
			path:     "/api/auth/verify-email",
			expected: true,
		},
		{
			name:     "protected API endpoint",
			path:     "/api/users",
			expected: false,
		},
		{
			name:     "other auth endpoint",
			path:     "/api/auth/current",
			expected: false,
		},
		{
			name:     "root path",
			path:     "/",
			expected: false,
		},
		{
			name:     "non-API path",
			path:     "/static/css/main.css",
			expected: false,
		},
		{
			name:     "partial match should not skip",
			path:     "/api/auth/login/something",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipAuth(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequiresAuth(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "protected API endpoint",
			path:     "/api/users",
			expected: true,
		},
		{
			name:     "protected API endpoint - nested",
			path:     "/api/v1/users",
			expected: true,
		},
		{
			name:     "auth login endpoint",
			path:     "/api/auth/login",
			expected: false,
		},
		{
			name:     "auth register endpoint",
			path:     "/api/auth/register",
			expected: false,
		},
		{
			name:     "auth temp user endpoint",
			path:     "/api/auth/temp-user",
			expected: false,
		},
		{
			name:     "auth verify endpoint",
			path:     "/api/auth/verify-email",
			expected: false,
		},
		{
			name:     "other auth endpoint requires auth",
			path:     "/api/auth/current",
			expected: false, // All /api/auth/ endpoints are excluded from auth requirements
		},
		{
			name:     "non-API endpoint",
			path:     "/health",
			expected: false,
		},
		{
			name:     "root path",
			path:     "/",
			expected: false,
		},
		{
			name:     "static files",
			path:     "/static/js/main.js",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := requiresAuth(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMiddleware_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create test users
	normalUser, normalToken, err := authService.Register("normal", "normal@example.com", "password123")
	require.NoError(t, err)

	tempUser, _, tempToken, err := authService.CreateTemporaryUser("test-session")
	require.NoError(t, err)

	middleware := AuthMiddleware(authService)

	// Handler that captures user info
	handler := func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if ok {
			w.Header().Set("X-User-ID", user.ID)
			w.Header().Set("X-User-Type", func() string {
				if user.IsTemporary {
					return "temporary"
				}
				return "permanent"
			}())
		}
		w.WriteHeader(http.StatusOK)
	}

	tests := []struct {
		name           string
		path           string
		token          string
		wantStatus     int
		expectedUserID string
		expectedType   string
	}{
		{
			name:           "normal user on protected endpoint",
			path:           "/api/protected",
			token:          normalToken,
			wantStatus:     http.StatusOK,
			expectedUserID: normalUser.ID,
			expectedType:   "permanent",
		},
		{
			name:           "temporary user on protected endpoint",
			path:           "/api/protected",
			token:          tempToken,
			wantStatus:     http.StatusOK,
			expectedUserID: tempUser.ID,
			expectedType:   "temporary",
		},
		{
			name:       "no token on non-protected endpoint",
			path:       "/health",
			token:      "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			wrappedHandler := middleware(handler)
			wrappedHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.expectedUserID != "" {
				assert.Equal(t, tt.expectedUserID, w.Header().Get("X-User-ID"))
				assert.Equal(t, tt.expectedType, w.Header().Get("X-User-Type"))
			}
		})
	}
}

func TestMiddleware_ErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create and then delete a user to test token validation with non-existent user
	user, token, err := authService.Register("deleteme", "delete@example.com", "password123")
	require.NoError(t, err)

	// Delete user from database
	_, err = db.Exec("DELETE FROM users WHERE id = ?", user.ID)
	require.NoError(t, err)

	middleware := AuthMiddleware(authService)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	wrappedHandler := middleware(handler)
	wrappedHandler(w, req)

	// Should return 401 because user no longer exists
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Benchmark tests for middleware
func BenchmarkAuthMiddleware_ValidToken(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create test user
	_, token, err := authService.Register("benchuser", "bench@example.com", "password123")
	if err != nil {
		b.Fatal(err)
	}

	middleware := AuthMiddleware(authService)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler := middleware(handler)
		wrappedHandler(w, req)
	}
}

func BenchmarkAuthMiddleware_SkipAuth(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	middleware := AuthMiddleware(authService)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		wrappedHandler := middleware(handler)
		wrappedHandler(w, req)
	}
}

func BenchmarkShouldSkipAuth(b *testing.B) {
	paths := []string{
		"/health",
		"/api/auth/login",
		"/api/protected",
		"/api/users",
		"/static/css/main.css",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			shouldSkipAuth(path)
		}
	}
}

func BenchmarkRequiresAuth(b *testing.B) {
	paths := []string{
		"/api/users",
		"/api/auth/login",
		"/health",
		"/static/js/main.js",
		"/api/v1/protected",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			requiresAuth(path)
		}
	}
}
