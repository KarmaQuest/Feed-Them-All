package auth

import "time"

// User represents a row in the users table.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	IsPremium    bool      `json:"is_premium"`
	XP           int       `json:"xp"`
	AvatarConfig []byte    `json:"avatar_config"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest is the payload for POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // feeder | giver | association
}

// LoginRequest is the payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse is returned after successful register or login.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}
