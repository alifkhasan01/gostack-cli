package entity

import "time"

// User represents a user in the system.
type User struct {
	ID        int       `json:"id"         db:"id"`
	Name      string    `json:"name"       db:"name"`
	Email     string    `json:"email"      db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest is the payload for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name"  binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// UpdateUserRequest is the payload for updating a user.
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"omitempty,email"`
}
