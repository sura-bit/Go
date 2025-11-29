package data

import (
	"context"
	"errors"
	"time"

	"task_manager/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists     = errors.New("username already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrBadCredentials = errors.New("invalid username or password")
)

type UserService struct {
	col *mongo.Collection
}

func NewUserService(col *mongo.Collection) *UserService {
	return &UserService{col: col}
}

const userTimeout = 5 * time.Second

func (s *UserService) Count(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()
	return s.col.CountDocuments(ctx, bson.D{})
}

func (s *UserService) FindByUsername(ctx context.Context, username string) (models.UserDB, error) {
	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()
	var u models.UserDB
	err := s.col.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.UserDB{}, ErrUserNotFound
		}
		return models.UserDB{}, err
	}
	return u, nil
}

// Register creates a new user. If no users exist, first user becomes admin.
func (s *UserService) Register(ctx context.Context, username, password string) (models.UserDB, error) {
	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	// uniqueness
	_, err := s.FindByUsername(ctx, username)
	if err == nil {
		return models.UserDB{}, ErrUserExists
	}
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		// an actual DB error
		return models.UserDB{}, err
	}

	role := models.RoleUser
	cnt, err := s.Count(ctx)
	if err != nil {
		return models.UserDB{}, err
	}
	if cnt == 0 {
		role = models.RoleAdmin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserDB{}, err
	}

	now := time.Now()
	doc := models.UserDB{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	res, err := s.col.InsertOne(ctx, doc)
	if err != nil {
		// Duplicate key (if unique index enforced)
		var we *mongo.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					return models.UserDB{}, ErrUserExists
				}
			}
		}
		return models.UserDB{}, err
	}
	doc.ID = res.InsertedID
	return doc, nil
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (models.UserDB, error) {
	u, err := s.FindByUsername(ctx, username)
	if err != nil {
		return models.UserDB{}, ErrBadCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return models.UserDB{}, ErrBadCredentials
	}
	return u, nil
}

func (s *UserService) PromoteToAdmin(ctx context.Context, userID string) (models.UserDB, error) {
	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return models.UserDB{}, ErrUserNotFound
	}
	var updated models.UserDB
err = s.col.FindOneAndUpdate(
	ctx,
	bson.M{"_id": oid},
	bson.M{"$set": bson.M{"role": models.RoleAdmin, "updated_at": time.Now()}},
	options.FindOneAndUpdate().SetReturnDocument(options.After),
).Decode(&updated)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.UserDB{}, ErrUserNotFound
		}
		return models.UserDB{}, err
	}
	return updated, nil
}
