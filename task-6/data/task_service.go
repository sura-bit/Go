package data

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"task_manager/models"
)

var (
	ErrNotFound      = errors.New("task not found")
	ErrInvalidStatus = errors.New("invalid task status")
	ErrInvalidDate   = errors.New("invalid due date (use RFC3339)")
)

type TaskService struct {
	col *mongo.Collection
}

func NewTaskService(col *mongo.Collection) *TaskService {
	return &TaskService{col: col}
}

const defaultTimeout = 5 * time.Second

func parseObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

func toOut(doc models.TaskDB) models.TaskOut {
	oid, _ := doc.ID.(primitive.ObjectID)
	return models.TaskOut{
		ID:          oid.Hex(),
		Title:       doc.Title,
		Description: doc.Description,
		DueDate:     doc.DueDate,
		Status:      doc.Status,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}
}

func (s *TaskService) List(ctx context.Context) ([]models.TaskOut, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cur, err := s.col.Find(ctx, bson.D{}, nil)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []models.TaskOut
	for cur.Next(ctx) {
		var t models.TaskDB
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		results = append(results, toOut(t))
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *TaskService) Get(ctx context.Context, id string) (models.TaskOut, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	oid, err := parseObjectID(id)
	if err != nil {
		return models.TaskOut{}, ErrNotFound
	}
	var t models.TaskDB
	if err := s.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&t); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.TaskOut{}, ErrNotFound
		}
		return models.TaskOut{}, err
	}
	return toOut(t), nil
}

func (s *TaskService) Create(ctx context.Context, dto models.CreateTaskDTO) (models.TaskOut, error) {
	if !models.IsValidStatus(dto.Status) {
		return models.TaskOut{}, ErrInvalidStatus
	}
	due, err := time.Parse(time.RFC3339, dto.DueDate)
	if err != nil {
		return models.TaskOut{}, ErrInvalidDate
	}

	now := time.Now()
	doc := models.TaskDB{
		Title:       dto.Title,
		Description: dto.Description,
		DueDate:     due,
		Status:      dto.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := s.col.InsertOne(ctx, doc)
	if err != nil {
		return models.TaskOut{}, err
	}
	doc.ID = res.InsertedID
	return toOut(doc), nil
}

func (s *TaskService) Update(ctx context.Context, id string, dto models.UpdateTaskDTO) (models.TaskOut, error) {
	oid, err := parseObjectID(id)
	if err != nil {
		return models.TaskOut{}, ErrNotFound
	}

	set := bson.M{"updated_at": time.Now()}
	if dto.Title != nil {
		set["title"] = *dto.Title
	}
	if dto.Description != nil {
		set["description"] = *dto.Description
	}
	if dto.Status != nil {
		if !models.IsValidStatus(*dto.Status) {
			return models.TaskOut{}, ErrInvalidStatus
		}
		set["status"] = *dto.Status
	}
	if dto.DueDate != nil {
		d, err := time.Parse(time.RFC3339, *dto.DueDate)
		if err != nil {
			return models.TaskOut{}, ErrInvalidDate
		}
		set["due_date"] = d
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	after := options.After
	var updated models.TaskDB
	err = s.col.FindOneAndUpdate(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(after),
	).Decode(&updated)
	

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.TaskOut{}, ErrNotFound
		}
		return models.TaskOut{}, err
	}
	return toOut(updated), nil
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	oid, err := parseObjectID(id)
	if err != nil {
		return ErrNotFound
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := s.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
