package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID            string     `json:"id"`
	Username      string     `json:"username"`
	Email         *string    `json:"email,omitempty"`
	EmailVerified bool       `json:"email_verified"`
	IsTemporary   bool       `json:"is_temporary"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsTemp   bool   `json:"is_temp"`
	jwt.RegisteredClaims
}

// AuthService handles authentication and user management
type AuthService struct {
	db          *sql.DB
	jwtSecret   []byte
	tokenExpiry time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	if jwtSecret == "" {
		// Generate a random secret if none provided
		jwtSecret = generateRandomSecret()
		log.Printf("🔐 Generated random JWT secret")
	}

	return &AuthService{
		db:          db,
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: 24 * time.Hour, // 24 hours
	}
}

// CreateTemporaryUser creates a temporary user for anonymous access
func (as *AuthService) CreateTemporaryUser(sessionID string) (*User, string, string, error) {
	// Generate temporary username
	tempUsername := fmt.Sprintf("temp_%s", generateRandomString(8))

	// Generate temporary password
	tempPassword := generateRandomString(12)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New().String()
	now := time.Now()

	// Insert user into database
	query := `
		INSERT INTO users (id, username, password_hash, is_temporary, created_at, updated_at)
		VALUES (?, ?, ?, TRUE, ?, ?)
	`

	_, err = as.db.Exec(query, userID, tempUsername, string(hashedPassword), now, now)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create temporary user: %w", err)
	}

	user := &User{
		ID:          userID,
		Username:    tempUsername,
		IsTemporary: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Generate JWT token
	token, err := as.generateToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("✅ Created temporary user: %s", tempUsername)
	return user, tempPassword, token, nil
}

// Login authenticates a user and returns a JWT token
func (as *AuthService) Login(username, password string) (*User, string, error) {
	// Get user from database
	query := `
		SELECT id, username, email, password_hash, email_verified, is_temporary, 
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE username = ?
	`

	var user User
	var passwordHash string
	var email sql.NullString
	var lastLoginAt sql.NullTime

	err := as.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &email, &passwordHash,
		&user.EmailVerified, &user.IsTemporary, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("invalid credentials")
		}
		return nil, "", fmt.Errorf("database error: %w", err)
	}

	if email.Valid {
		user.Email = &email.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	updateQuery := `UPDATE users SET last_login_at = ?, updated_at = ? WHERE id = ?`
	_, err = as.db.Exec(updateQuery, now, now, user.ID)
	if err != nil {
		log.Printf("⚠️ Failed to update last login time: %v", err)
	}
	user.LastLoginAt = &now
	user.UpdatedAt = now

	// Generate JWT token
	token, err := as.generateToken(&user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("✅ User logged in: %s", username)
	return &user, token, nil
}

// Register creates a new permanent user account
func (as *AuthService) Register(username, email, password string) (*User, string, error) {
	// Check if username already exists
	var exists bool
	err := as.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		return nil, "", fmt.Errorf("database error: %w", err)
	}
	if exists {
		return nil, "", fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if email != "" {
		err = as.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
		if err != nil {
			return nil, "", fmt.Errorf("database error: %w", err)
		}
		if exists {
			return nil, "", fmt.Errorf("email already exists")
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New().String()
	now := time.Now()

	// Insert user
	query := `
		INSERT INTO users (id, username, email, password_hash, email_verified, is_temporary, created_at, updated_at)
		VALUES (?, ?, ?, ?, FALSE, FALSE, ?, ?)
	`

	_, err = as.db.Exec(query, userID, username, email, string(hashedPassword), now, now)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	user := &User{
		ID:            userID,
		Username:      username,
		EmailVerified: false,
		IsTemporary:   false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Only set email if it's not empty
	if email != "" {
		user.Email = &email
	}

	// Generate JWT token
	token, err := as.generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("✅ User registered: %s", username)
	return user, token, nil
}

// SaveTemporaryAccount converts a temporary account to a permanent one
func (as *AuthService) SaveTemporaryAccount(userID, email, currentPassword string) (*User, error) {
	// Get current user
	user, err := as.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsTemporary {
		return nil, fmt.Errorf("user is not temporary")
	}

	// Verify current password
	var passwordHash string
	err = as.db.QueryRow("SELECT password_hash FROM users WHERE id = ?", userID).Scan(&passwordHash)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(currentPassword))
	if err != nil {
		return nil, fmt.Errorf("invalid current password")
	}

	// Check if email already exists
	var exists bool
	err = as.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ? AND id != ?)", email, userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already exists")
	}

	// Update user
	now := time.Now()
	query := `
		UPDATE users 
		SET email = ?, is_temporary = FALSE, updated_at = ?
		WHERE id = ?
	`

	_, err = as.db.Exec(query, email, now, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user
	user, err = as.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	log.Printf("✅ Temporary account saved: %s -> %s", user.Username, email)
	return user, nil
}

// VerifyEmail verifies a user's email address
func (as *AuthService) VerifyEmail(token string) (*User, error) {
	// Find user by verification token
	var userID string
	var expiresAt time.Time

	query := `
		SELECT id, email_verification_expires_at 
		FROM users 
		WHERE email_verification_token = ?
	`

	err := as.db.QueryRow(query, token).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid verification token")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("verification token expired")
	}

	// Update user as verified
	now := time.Now()
	updateQuery := `
		UPDATE users 
		SET email_verified = TRUE, email_verification_token = NULL, 
		    email_verification_expires_at = NULL, updated_at = ?
		WHERE id = ?
	`

	_, err = as.db.Exec(updateQuery, now, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// Get updated user
	user, err := as.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	log.Printf("✅ Email verified for user: %s", user.Username)
	return user, nil
}

// GetUserByID retrieves a user by ID
func (as *AuthService) GetUserByID(userID string) (*User, error) {
	query := `
		SELECT id, username, email, email_verified, is_temporary, 
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE id = ?
	`

	var user User
	var email sql.NullString
	var lastLoginAt sql.NullTime

	err := as.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &email, &user.EmailVerified,
		&user.IsTemporary, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if email.Valid {
		user.Email = &email.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// ValidateToken validates a JWT token and returns the user
func (as *AuthService) ValidateToken(tokenString string) (*User, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return as.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Get user from database
	user, err := as.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// generateToken generates a JWT token for a user
func (as *AuthService) generateToken(user *User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(as.tokenExpiry)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsTemp:   user.IsTemporary,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gogent",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(as.jwtSecret)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// Use crypto/rand for better randomness
		randBytes := make([]byte, 1)
		rand.Read(randBytes)
		b[i] = charset[int(randBytes[0])%len(charset)]
	}
	return string(b)
}

// generateRandomSecret generates a random JWT secret
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("no authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}
