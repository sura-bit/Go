package models

import (
	"time"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

func IsValidStatus(s TaskStatus) bool {
	switch s {
	case StatusPending, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

type CreateTaskDTO struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     string     `json:"due_date" binding:"required"` 
	Status      TaskStatus `json:"status" binding:"required"`
}

type UpdateTaskDTO struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	DueDate     *string     `json:"due_date"`
	Status      *TaskStatus `json:"status"`
}


type TaskDB struct {
	ID          interface{} `bson:"_id,omitempty"` 
	Title       string      `bson:"title"`
	Description string      `bson:"description"`
	DueDate     time.Time   `bson:"due_date"`
	Status      TaskStatus  `bson:"status"`
	CreatedAt   time.Time   `bson:"created_at"`
	UpdatedAt   time.Time   `bson:"updated_at"`
}


type TaskOut struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     time.Time  `json:"due_date"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
