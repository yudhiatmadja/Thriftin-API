package model

import "time"

type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	PasswordHash string  `json:"-"`
	Username   string    `json:"username"`
	FullName   string    `json:"full_name"`
	Role       string    `json:"role"`
	IsVerified bool      `json:"is_verified"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}
