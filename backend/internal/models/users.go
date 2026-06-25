package models

import "time"

// User represents a system user
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // omitted from JSON serialization for security
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Credentials is used for login
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is used for registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}
