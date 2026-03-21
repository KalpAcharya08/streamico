package user

import "time"

// User represents the internal database record.
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	AvatarURL    *string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserProfile is the safe JSON representation of a user.
type UserProfile struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
