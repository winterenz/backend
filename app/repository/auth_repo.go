package repository

import (
	"context"

	"prak/clean-architecture-fiber-mongo/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	GetByUsernameOrEmail(ctx context.Context, x string) (*model.User, error)
}

type userRepo struct{ c *mongo.Collection }

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepo{c: db.Collection("users")}
}

func (r *userRepo) GetByUsernameOrEmail(ctx context.Context, x string) (*model.User, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"username": x},
			{"email": x},
		},
	}

	var u model.User
	err := r.c.FindOne(ctx, filter).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, nil 
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
