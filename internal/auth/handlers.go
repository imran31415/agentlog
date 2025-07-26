package auth

import (
	"encoding/json"
	"net/http"
	"time"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string    `json:"token"`
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// CreateTemporaryUserRequest represents the temporary user creation request
type CreateTemporaryUserRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

// CreateTemporaryUserResponse represents the temporary user creation response
type CreateTemporaryUserResponse struct {
	User              *User  `json:"user"`
	TemporaryPassword string `json:"temporary_password"`
	Token             string `json:"token"`
}

// SaveTemporaryAccountRequest represents the save temporary account request
type SaveTemporaryAccountRequest struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
}

// SaveTemporaryAccountResponse represents the save temporary account response
type SaveTemporaryAccountResponse struct {
	User      *User `json:"user"`
	EmailSent bool  `json:"email_sent"`
}

// ConnectTemporaryAccountRequest represents the connect temporary account request
type ConnectTemporaryAccountRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"newPassword"`
}

// ConnectTemporaryAccountResponse represents the connect temporary account response
type ConnectTemporaryAccountResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// VerifyEmailRequest represents the email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmailResponse represents the email verification response
type VerifyEmailResponse struct {
	User     *User `json:"user"`
	Verified bool  `json:"verified"`
}

// GetCurrentUserResponse represents the current user response
type GetCurrentUserResponse struct {
	User *User `json:"user"`
}

// AuthHandlers provides HTTP handlers for authentication
type AuthHandlers struct {
	authService *AuthService
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService *AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// LoginHandler handles user login
func (ah *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, token, err := ah.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	response := LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterHandler handles user registration
func (ah *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, token, err := ah.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := RegisterResponse{
		User:  user,
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateTemporaryUserHandler handles temporary user creation
func (ah *AuthHandlers) CreateTemporaryUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTemporaryUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, tempPassword, token, err := ah.authService.CreateTemporaryUser(req.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CreateTemporaryUserResponse{
		User:              user,
		TemporaryPassword: tempPassword,
		Token:             token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SaveTemporaryAccountHandler handles saving temporary accounts
func (ah *AuthHandlers) SaveTemporaryAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SaveTemporaryAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user from context (must be authenticated)
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	updatedUser, err := ah.authService.SaveTemporaryAccount(user.ID, req.Email, req.CurrentPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Send verification email
	emailSent := false // Placeholder for email sending logic

	response := SaveTemporaryAccountResponse{
		User:      updatedUser,
		EmailSent: emailSent,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConnectTemporaryAccountHandler handles connecting temporary accounts to email with new password
func (ah *AuthHandlers) ConnectTemporaryAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConnectTemporaryAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user from context (must be authenticated)
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	updatedUser, newToken, err := ah.authService.ConnectTemporaryAccount(user.ID, req.Email, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := ConnectTemporaryAccountResponse{
		Token: newToken,
		User:  updatedUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyEmailHandler handles email verification
func (ah *AuthHandlers) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := ah.authService.VerifyEmail(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := VerifyEmailResponse{
		User:     user,
		Verified: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentUserHandler handles getting current user information
func (ah *AuthHandlers) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	response := GetCurrentUserResponse{
		User: user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
