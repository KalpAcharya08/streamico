package auth

import "time"

// =============================================================================
// HTTP Request / Response Models
// =============================================================================

// RegisterRequest defines the expected JSON payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest defines the expected JSON payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse defines the JSON response sent back on successful login/register.
// Notice that we explicitly DO NOT include the password_hash here.
type AuthResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

// UserInfo is a safe representation of the user data to send to the client.
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// =============================================================================
// Internal Data Models (Database)
// =============================================================================

// User represents the full database row from the `users` table.
// This is used internally by the Service and Repository, but never sent directly
// to the client (to avoid leaking PasswordHash).
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	AvatarURL    *string // Use pointer for nullable strings in database
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
