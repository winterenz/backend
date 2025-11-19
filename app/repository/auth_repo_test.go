package repository

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestUserRepository_GetByUsernameOrEmail(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success - find by username", func(mt *mtest.T) {
		repo := NewUserRepository(mt.DB)
		userID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: userID},
				{Key: "username", Value: "testuser"},
				{Key: "email", Value: "test@example.com"},
				{Key: "role", Value: "user"},
			}))

		user, err := repo.GetByUsernameOrEmail(context.Background(), "testuser")
		if err != nil {
			t.Errorf("GetByUsernameOrEmail() error = %v", err)
		}
		if user == nil {
			t.Error("GetByUsernameOrEmail() returned nil user")
		}
		if user != nil {
			if user.Username != "testuser" {
				t.Errorf("GetByUsernameOrEmail() Username = %v, want testuser", user.Username)
			}
			if user.Email != "test@example.com" {
				t.Errorf("GetByUsernameOrEmail() Email = %v, want test@example.com", user.Email)
			}
		}
	})

	mt.Run("success - find by email", func(mt *mtest.T) {
		repo := NewUserRepository(mt.DB)
		userID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: userID},
				{Key: "username", Value: "testuser"},
				{Key: "email", Value: "test@example.com"},
				{Key: "role", Value: "user"},
			}))

		user, err := repo.GetByUsernameOrEmail(context.Background(), "test@example.com")
		if err != nil {
			t.Errorf("GetByUsernameOrEmail() error = %v", err)
		}
		if user == nil {
			t.Error("GetByUsernameOrEmail() returned nil user")
		}
		if user != nil {
			if user.Email != "test@example.com" {
				t.Errorf("GetByUsernameOrEmail() Email = %v, want test@example.com", user.Email)
			}
		}
	})

	mt.Run("not found - returns nil, nil", func(mt *mtest.T) {
		repo := NewUserRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch))

		user, err := repo.GetByUsernameOrEmail(context.Background(), "nonexistent")
		if err != nil {
			t.Errorf("GetByUsernameOrEmail() error = %v, want nil", err)
		}
		if user != nil {
			t.Errorf("GetByUsernameOrEmail() user = %v, want nil", user)
		}
	})
}
