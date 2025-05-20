package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collectionName = "users"

type UserRepository struct {
	collection *mongo.Collection
}

var _ domain.UserRepository = (*UserRepository)(nil)

func NewUserRepository(ctx context.Context, db *mongo.Database) (*UserRepository, error) {
	// Set email as unique key
	if err := ensureUserIndexes(ctx, db.Collection(collectionName)); err != nil {
		return nil, err
	}
	return &UserRepository{collection: db.Collection(collectionName)}, nil
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	u.ID = uuid.New().String()
	u.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, bson.M{
		"_id":        u.ID,
		"email":      u.Email,
		"password":   u.Password,
		"role":       string(u.Role), // store as string
		"created_at": u.CreatedAt,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("repo.Create: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("findByEmail: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("finByID: %w", err)
	}

	return nil, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func ensureUserIndexes(ctx context.Context, col *mongo.Collection) error {
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := col.Indexes().CreateOne(ctx, index)
	return err
}
