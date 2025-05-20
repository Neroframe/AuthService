package mongo

import "go.mongodb.org/mongo-driver/mongo"

var collectionName = "users"

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{collection: db.Collection(collectionName)}
}
