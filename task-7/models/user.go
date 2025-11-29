package models

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type UserDB struct {
	ID           interface{} `bson:"_id,omitempty"`
	Username     string      `bson:"username"`
	PasswordHash string      `bson:"password_hash"`
	Role         Role        `bson:"role"`
	CreatedAt    time.Time   `bson:"created_at"`
	UpdatedAt    time.Time   `bson:"updated_at"`
}

type RegisterDTO struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserOut struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

type PromoteDTO struct {
	UserID string `json:"user_id" binding:"required"` 
}
