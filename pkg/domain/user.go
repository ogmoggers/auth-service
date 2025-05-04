package domain

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthCode struct {
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}
