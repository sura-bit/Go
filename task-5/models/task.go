package models

import "time"

type TaskStatus string

const (
	StatusPending     TaskStatus = "pending"
	StatusInProgress  TaskStatus = "in_progress"
	StatusDone        TaskStatus = "done"
)

func IsValidStatus(s TaskStatus) bool {
	switch s {
	case StatusPending, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     time.Time  `json:"due_date"` // RFC3339 in/out
	Status      TaskStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Payloads used for create/update (to validate input cleanly)
type CreateTaskDTO struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     string     `json:"due_date" binding:"required"` // RFC3339 string
	Status      TaskStatus `json:"status" binding:"required"`
}

type UpdateTaskDTO struct {
	Title       *string     `json:"title"`       // optional
	Description *string     `json:"description"` // optional
	DueDate     *string     `json:"due_date"`    // optional, RFC3339
	Status      *TaskStatus `json:"status"`      // optional
}
