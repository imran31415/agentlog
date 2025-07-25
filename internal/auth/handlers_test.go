package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandlersTest(t *testing.T) (*AuthHandlers, *AuthService) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })

	authService := NewAuthService(db, "test-secret")
	handlers := NewAuthHandlers(authService)

	return handlers, authService
}

func TestNewAuthHandlers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	handlers := NewAuthHandlers(authService)
	assert.NotNil(t, handlers)
	assert.Equal(t, authService, handlers.authService)
}

func TestAuthHandlers_LoginHandler(t *testing.T) {
	handlers, authService := setupHandlersTest(t)

	// Create a test user
	_, _, err := authService.Register("testuser", "test@example.com", "password123")
	require.NoError(t, err)

	tests := []struct {
		name       string
		method     string
		body       interface{}
		wantStatus int
		wantToken  bool
	}{
		{
			name:   "successful login",
			method: http.MethodPost,
			body: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:   "invalid credentials",
			method: http.MethodPost,
			body: LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name:   "user not found",
			method: http.MethodPost,
			body: LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/auth/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.LoginHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantToken {
				var response LoginResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.NotNil(t, response.User)
				assert.NotZero(t, response.ExpiresAt)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestAuthHandlers_RegisterHandler(t *testing.T) {
	handlers, authService := setupHandlersTest(t)

	// Create a user to test conflicts
	_, _, err := authService.Register("existing", "existing@example.com", "password123")
	require.NoError(t, err)

	tests := []struct {
		name       string
		method     string
		body       interface{}
		wantStatus int
		wantToken  bool
	}{
		{
			name:   "successful registration",
			method: http.MethodPost,
			body: RegisterRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:   "successful registration without email",
			method: http.MethodPost,
			body: RegisterRequest{
				Username: "newuser2",
				Email:    "",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:   "duplicate username",
			method: http.MethodPost,
			body: RegisterRequest{
				Username: "existing",
				Email:    "different@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:   "duplicate email",
			method: http.MethodPost,
			body: RegisterRequest{
				Username: "different",
				Email:    "existing@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/auth/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.RegisterHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantToken {
				var response RegisterResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.NotNil(t, response.User)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestAuthHandlers_CreateTemporaryUserHandler(t *testing.T) {
	handlers, _ := setupHandlersTest(t)

	tests := []struct {
		name       string
		method     string
		body       interface{}
		wantStatus int
		wantUser   bool
	}{
		{
			name:   "successful creation with session ID",
			method: http.MethodPost,
			body: CreateTemporaryUserRequest{
				SessionID: "test-session-123",
			},
			wantStatus: http.StatusOK,
			wantUser:   true,
		},
		{
			name:   "successful creation without session ID",
			method: http.MethodPost,
			body: CreateTemporaryUserRequest{
				SessionID: "",
			},
			wantStatus: http.StatusOK,
			wantUser:   true,
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
			wantUser:   false,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
			wantUser:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/auth/temp-user", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.CreateTemporaryUserHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantUser {
				var response CreateTemporaryUserResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotNil(t, response.User)
				assert.NotEmpty(t, response.TemporaryPassword)
				assert.NotEmpty(t, response.Token)
				assert.True(t, response.User.IsTemporary)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestAuthHandlers_SaveTemporaryAccountHandler(t *testing.T) {
	handlers, authService := setupHandlersTest(t)

	// Create a temporary user
	tempUser, tempPassword, _, err := authService.CreateTemporaryUser("test-session")
	require.NoError(t, err)

	tests := []struct {
		name       string
		method     string
		body       interface{}
		user       *User
		wantStatus int
		wantUser   bool
	}{
		{
			name:   "successful save",
			method: http.MethodPost,
			body: SaveTemporaryAccountRequest{
				Email:           "save@example.com",
				CurrentPassword: tempPassword,
			},
			user:       tempUser,
			wantStatus: http.StatusOK,
			wantUser:   true,
		},
		{
			name:   "wrong password",
			method: http.MethodPost,
			body: SaveTemporaryAccountRequest{
				Email:           "save2@example.com",
				CurrentPassword: "wrongpassword",
			},
			user:       tempUser,
			wantStatus: http.StatusBadRequest,
			wantUser:   false,
		},
		{
			name:   "no authentication",
			method: http.MethodPost,
			body: SaveTemporaryAccountRequest{
				Email:           "save3@example.com",
				CurrentPassword: tempPassword,
			},
			user:       nil,
			wantStatus: http.StatusUnauthorized,
			wantUser:   false,
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       "invalid json",
			user:       tempUser,
			wantStatus: http.StatusBadRequest,
			wantUser:   false,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			user:       tempUser,
			wantStatus: http.StatusMethodNotAllowed,
			wantUser:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/auth/save-temp", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Add user to context if provided
			if tt.user != nil {
				ctx := context.WithValue(req.Context(), UserContextKey{}, tt.user)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handlers.SaveTemporaryAccountHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantUser {
				var response SaveTemporaryAccountResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotNil(t, response.User)
				assert.False(t, response.User.IsTemporary)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestAuthHandlers_VerifyEmailHandler(t *testing.T) {
	handlers, authService := setupHandlersTest(t)

	// Create a user and set up verification token
	user, _, err := authService.Register("verifytest", "verify@example.com", "password123")
	require.NoError(t, err)

	// Add verification token directly to database
	verificationToken := "test-verification-token"
	expiresAt := time.Now().Add(time.Hour)
	_, err = authService.db.Exec(`
		UPDATE users 
		SET email_verification_token = ?, email_verification_expires_at = ?
		WHERE id = ?
	`, verificationToken, expiresAt, user.ID)
	require.NoError(t, err)

	tests := []struct {
		name       string
		method     string
		body       interface{}
		wantStatus int
		wantUser   bool
	}{
		{
			name:   "successful verification",
			method: http.MethodPost,
			body: VerifyEmailRequest{
				Token: verificationToken,
			},
			wantStatus: http.StatusOK,
			wantUser:   true,
		},
		{
			name:   "invalid token",
			method: http.MethodPost,
			body: VerifyEmailRequest{
				Token: "invalid-token",
			},
			wantStatus: http.StatusBadRequest,
			wantUser:   false,
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
			wantUser:   false,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
			wantUser:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/auth/verify-email", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.VerifyEmailHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantUser {
				var response VerifyEmailResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotNil(t, response.User)
				assert.True(t, response.Verified)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestAuthHandlers_GetCurrentUserHandler(t *testing.T) {
	handlers, authService := setupHandlersTest(t)

	// Create a test user
	user, _, err := authService.Register("currenttest", "current@example.com", "password123")
	require.NoError(t, err)

	tests := []struct {
		name       string
		method     string
		user       *User
		wantStatus int
		wantUser   bool
	}{
		{
			name:       "successful get current user",
			method:     http.MethodGet,
			user:       user,
			wantStatus: http.StatusOK,
			wantUser:   true,
		},
		{
			name:       "no authentication",
			method:     http.MethodGet,
			user:       nil,
			wantStatus: http.StatusUnauthorized,
			wantUser:   false,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			user:       user,
			wantStatus: http.StatusMethodNotAllowed,
			wantUser:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/auth/current", nil)

			// Add user to context if provided
			if tt.user != nil {
				ctx := context.WithValue(req.Context(), UserContextKey{}, tt.user)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handlers.GetCurrentUserHandler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantUser {
				var response GetCurrentUserResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotNil(t, response.User)
				assert.Equal(t, user.ID, response.User.ID)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

// Benchmark tests for handlers
func BenchmarkLoginHandler(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()
	authService := NewAuthService(db, "test-secret")
	handlers := NewAuthHandlers(authService)

	// Create test user
	_, _, err := authService.Register("benchuser", "bench@example.com", "password123")
	if err != nil {
		b.Fatal(err)
	}

	body := LoginRequest{
		Username: "benchuser",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.LoginHandler(w, req)
	}
}

func BenchmarkCreateTemporaryUserHandler(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()
	authService := NewAuthService(db, "test-secret")
	handlers := NewAuthHandlers(authService)

	body := CreateTemporaryUserRequest{
		SessionID: "bench-session",
	}
	reqBody, _ := json.Marshal(body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/temp-user", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreateTemporaryUserHandler(w, req)
	}
}
