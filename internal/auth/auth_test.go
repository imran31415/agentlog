package auth

import (
	"database/sql"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create users table
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		password_hash TEXT NOT NULL,
		email_verified BOOLEAN DEFAULT FALSE,
		email_verification_token TEXT,
		email_verification_expires_at DATETIME,
		is_temporary BOOLEAN DEFAULT FALSE,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		last_login_at DATETIME
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestNewAuthService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		jwtSecret string
		wantLen   int
	}{
		{
			name:      "with provided secret",
			jwtSecret: "test-secret-key",
			wantLen:   len("test-secret-key"),
		},
		{
			name:      "with empty secret generates random",
			jwtSecret: "",
			wantLen:   64, // hex encoded 32 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := NewAuthService(db, tt.jwtSecret)

			assert.NotNil(t, authService)
			assert.Equal(t, db, authService.db)
			assert.Len(t, authService.jwtSecret, tt.wantLen)
			assert.Equal(t, 24*time.Hour, authService.tokenExpiry)
		})
	}
}

func TestAuthService_CreateTemporaryUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
	}{
		{
			name:      "successful creation",
			sessionID: "test-session-123",
			wantErr:   false,
		},
		{
			name:      "successful creation with empty session",
			sessionID: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, tempPassword, token, err := authService.CreateTemporaryUser(tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.NotEmpty(t, tempPassword)
			assert.NotEmpty(t, token)
			assert.True(t, user.IsTemporary)
			assert.Contains(t, user.Username, "temp_")
			assert.Equal(t, 12, len(tempPassword))

			// Verify user was created in database
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", user.ID).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count)

			// Verify token is valid
			parsedUser, err := authService.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, user.ID, parsedUser.ID)
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "successful registration",
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "successful registration without email",
			username: "testuser2",
			email:    "",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "duplicate username",
			username: "testuser",
			email:    "different@example.com",
			password: "password123",
			wantErr:  true,
			errMsg:   "username already exists",
		},
		{
			name:     "duplicate email",
			username: "differentuser",
			email:    "test@example.com",
			password: "password123",
			wantErr:  true,
			errMsg:   "email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := authService.Register(tt.username, tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.NotEmpty(t, token)
			assert.Equal(t, tt.username, user.Username)
			assert.False(t, user.IsTemporary)
			assert.False(t, user.EmailVerified)

			if tt.email != "" {
				assert.NotNil(t, user.Email)
				assert.Equal(t, tt.email, *user.Email)
			} else {
				assert.Nil(t, user.Email)
			}

			// Verify token is valid
			parsedUser, err := authService.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, user.ID, parsedUser.ID)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create a test user first
	testUser, _, err := authService.Register("logintest", "login@example.com", "password123")
	require.NoError(t, err)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "successful login",
			username: "logintest",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "wrong password",
			username: "logintest",
			password: "wrongpassword",
			wantErr:  true,
			errMsg:   "invalid credentials",
		},
		{
			name:     "non-existent user",
			username: "nonexistent",
			password: "password123",
			wantErr:  true,
			errMsg:   "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := authService.Login(tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.NotEmpty(t, token)
			assert.Equal(t, testUser.ID, user.ID)
			assert.NotNil(t, user.LastLoginAt)

			// Verify token is valid
			parsedUser, err := authService.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, user.ID, parsedUser.ID)
		})
	}
}

func TestAuthService_SaveTemporaryAccount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	tests := []struct {
		name      string
		setupFunc func() (string, string, string) // returns userID, email, password
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful save",
			setupFunc: func() (string, string, string) {
				tempUser, tempPassword, _, err := authService.CreateTemporaryUser("test-session-1")
				require.NoError(t, err)
				return tempUser.ID, "save@example.com", tempPassword
			},
			wantErr: false,
		},
		{
			name: "wrong password",
			setupFunc: func() (string, string, string) {
				tempUser, _, _, err := authService.CreateTemporaryUser("test-session-2")
				require.NoError(t, err)
				return tempUser.ID, "save2@example.com", "wrongpassword"
			},
			wantErr: true,
			errMsg:  "invalid current password",
		},
		{
			name: "email already exists",
			setupFunc: func() (string, string, string) {
				// Create a permanent user to test email conflict
				_, _, err := authService.Register("permanent", "existing@example.com", "password123")
				require.NoError(t, err)

				tempUser, tempPassword, _, err := authService.CreateTemporaryUser("test-session-3")
				require.NoError(t, err)
				return tempUser.ID, "existing@example.com", tempPassword
			},
			wantErr: true,
			errMsg:  "email already exists",
		},
		{
			name: "non-existent user",
			setupFunc: func() (string, string, string) {
				return "non-existent-id", "new@example.com", "password"
			},
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, email, password := tt.setupFunc()
			user, err := authService.SaveTemporaryAccount(userID, email, password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.False(t, user.IsTemporary)
			assert.Equal(t, email, *user.Email)
		})
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create a test user
	testUser, _, err := authService.Register("gettest", "get@example.com", "password123")
	require.NoError(t, err)

	tests := []struct {
		name    string
		userID  string
		want    *User
		wantErr bool
		errMsg  string
	}{
		{
			name:    "existing user",
			userID:  testUser.ID,
			want:    testUser,
			wantErr: false,
		},
		{
			name:    "non-existent user",
			userID:  "non-existent-id",
			want:    nil,
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := authService.GetUserByID(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.ID, user.ID)
			assert.Equal(t, tt.want.Username, user.Username)
			assert.Equal(t, tt.want.Email, user.Email)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create a test user and get a token
	testUser, token, err := authService.Register("tokentest", "token@example.com", "password123")
	require.NoError(t, err)

	// Create an invalid token with wrong secret
	invalidAuthService := NewAuthService(db, "wrong-secret")
	invalidToken, err := invalidAuthService.generateToken(testUser)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "invalid token signature",
			token:   invalidToken,
			wantErr: true,
			errMsg:  "invalid token",
		},
		{
			name:    "malformed token",
			token:   "invalid.token.format",
			wantErr: true,
			errMsg:  "invalid token",
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errMsg:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := authService.ValidateToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testUser.ID, user.ID)
			assert.Equal(t, testUser.Username, user.Username)
		})
	}
}

func TestAuthService_generateToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	user := &User{
		ID:          "test-id",
		Username:    "testuser",
		IsTemporary: false,
	}

	token, err := authService.generateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return authService.jwtSecret, nil
	})
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*Claims)
	require.True(t, ok)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.IsTemporary, claims.IsTemp)
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid bearer token",
			authHeader: "Bearer abc123",
			want:       "abc123",
			wantErr:    false,
		},
		{
			name:       "empty header",
			authHeader: "",
			want:       "",
			wantErr:    true,
			errMsg:     "no authorization header",
		},
		{
			name:       "invalid format - no bearer",
			authHeader: "abc123",
			want:       "",
			wantErr:    true,
			errMsg:     "invalid authorization header format",
		},
		{
			name:       "invalid format - wrong prefix",
			authHeader: "Basic abc123",
			want:       "",
			wantErr:    true,
			errMsg:     "invalid authorization header format",
		},
		{
			name:       "invalid format - too many parts",
			authHeader: "Bearer abc 123",
			want:       "",
			wantErr:    true,
			errMsg:     "invalid authorization header format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.authHeader)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, token)
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 8", 8},
		{"length 12", 12},
		{"length 16", 16},
		{"length 0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRandomString(tt.length)
			assert.Equal(t, tt.length, len(result))

			// Verify it contains only valid characters
			charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, char := range result {
				assert.Contains(t, charset, string(char))
			}
		})
	}

	// Test uniqueness
	t.Run("generates unique strings", func(t *testing.T) {
		str1 := generateRandomString(10)
		str2 := generateRandomString(10)
		assert.NotEqual(t, str1, str2)
	})
}

func TestGenerateRandomSecret(t *testing.T) {
	secret1 := generateRandomSecret()
	secret2 := generateRandomSecret()

	assert.Equal(t, 64, len(secret1)) // 32 bytes hex encoded
	assert.Equal(t, 64, len(secret2))
	assert.NotEqual(t, secret1, secret2)

	// Verify it's valid hex
	for _, char := range secret1 {
		assert.Contains(t, "0123456789abcdef", string(char))
	}
}

func TestAuthService_VerifyEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	authService := NewAuthService(db, "test-secret")

	// Create a user and set up email verification token
	user, _, err := authService.Register("verifytest", "verify@example.com", "password123")
	require.NoError(t, err)

	// Add verification token directly to database for testing
	verificationToken := "test-verification-token"
	expiresAt := time.Now().Add(time.Hour)
	_, err = db.Exec(`
		UPDATE users 
		SET email_verification_token = ?, email_verification_expires_at = ?
		WHERE id = ?
	`, verificationToken, expiresAt, user.ID)
	require.NoError(t, err)

	// Create an expired token for another user
	expiredUser, _, err := authService.Register("expiredtest", "expired@example.com", "password123")
	require.NoError(t, err)
	expiredToken := "expired-token"
	expiredTime := time.Now().Add(-time.Hour)
	_, err = db.Exec(`
		UPDATE users 
		SET email_verification_token = ?, email_verification_expires_at = ?
		WHERE id = ?
	`, expiredToken, expiredTime, expiredUser.ID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			token:   verificationToken,
			wantErr: false,
		},
		{
			name:    "expired token",
			token:   expiredToken,
			wantErr: true,
			errMsg:  "verification token expired",
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: true,
			errMsg:  "invalid verification token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifiedUser, err := authService.VerifyEmail(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, verifiedUser)
			assert.True(t, verifiedUser.EmailVerified)
			assert.Equal(t, user.ID, verifiedUser.ID)
		})
	}
}
