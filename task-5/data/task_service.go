package data

import (
	"errors"
	"sync"
	"time"

	"task_manager/models"
)

var (
	ErrNotFound      = errors.New("task not found")
	ErrInvalidStatus = errors.New("invalid task status")
	ErrInvalidDate   = errors.New("invalid due date (use RFC3339)")
)

// InMemoryTaskService provides a thread-safe in-memory store.
type InMemoryTaskService struct {
	mu    sync.RWMutex
	seq   int64
	tasks map[int64]models.Task
}

func NewInMemoryTaskService() *InMemoryTaskService {
	return &InMemoryTaskService{
		tasks: make(map[int64]models.Task),
	}
}

func (s *InMemoryTaskService) List() []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out
}

func (s *InMemoryTaskService) Get(id int64) (models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	return t, nil
}

func (s *InMemoryTaskService) Create(dto models.CreateTaskDTO) (models.Task, error) {
	if !models.IsValidStatus(dto.Status) {
		return models.Task{}, ErrInvalidStatus
	}
	due, err := time.Parse(time.RFC3339, dto.DueDate)
	if err != nil {
		return models.Task{}, ErrInvalidDate
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	task := models.Task{
		ID:          s.seq,
		Title:       dto.Title,
		Description: dto.Description,
		DueDate:     due,
		Status:      dto.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.tasks[task.ID] = task
	return task, nil
}

func (s *InMemoryTaskService) Update(id int64, dto models.UpdateTaskDTO) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}

	if dto.Title != nil {
		task.Title = *dto.Title
	}
	if dto.Description != nil {
		task.Description = *dto.Description
	}
	if dto.Status != nil {
		if !models.IsValidStatus(*dto.Status) {
			return models.Task{}, ErrInvalidStatus
		}
		task.Status = *dto.Status
	}
	if dto.DueDate != nil {
		d, err := time.Parse(time.RFC3339, *dto.DueDate)
		if err != nil {
			return models.Task{}, ErrInvalidDate
		}
		task.DueDate = d
	}
	task.UpdatedAt = time.Now()
	s.tasks[id] = task
	return task, nil
}

func (s *InMemoryTaskService) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}
