package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collectionName = "users"

type UserRepository struct {
	collection *mongo.Collection
}

var _ repository.UserRepository = (*UserRepository)(nil)

func NewUserRepository(ctx context.Context, db *mongo.Database) (*UserRepository, error) {
	// Set email as unique key
	if err := ensureUserIndexes(ctx, db.Collection(collectionName)); err != nil {
		return nil, fmt.Errorf("repo error in defining unique email index: %w", err)
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
			return repository.ErrEmailAlreadyUsed
		}
		return fmt.Errorf("repo Create: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("repo FinbyEmail: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("repo FindOne: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User, fields ...string) (*domain.User, error) {
	filter := bson.M{"_id": u.ID}
	update := buildUpdateFields(u, fields...)

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated domain.User

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("repo Update: %w", err)
	}
	return &updated, nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("repo Delete: %w", err)
	}

	if res.DeletedCount == 0 {
		return repository.ErrNotFound
	}

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

func buildUpdateFields(u *domain.User, allowedFields ...string) bson.M {
	set := bson.M{}
	for _, field := range allowedFields {
		switch field {
		case "username":
			set["username"] = u.Username
		case "phone":
			set["phone"] = u.Phone
		case "password":
			set["password"] = u.Password
		case "email":
			set["email"] = u.Email
		case "updated_at":
			set["updated_at"] = u.UpdatedAt
		}
	}
	return bson.M{"$set": set}
}
