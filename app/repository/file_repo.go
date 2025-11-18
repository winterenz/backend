package repository

import (
	"context"
	"time"

	"prak3/clean-architecture-fiber-mongo/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileRepository interface {
	Create(ctx context.Context, f *model.FileDoc) error
	FindAll(ctx context.Context) ([]model.FileDoc, error)
	FindByID(ctx context.Context, id string) (*model.FileDoc, error)
	Delete(ctx context.Context, id string) error
}

type fileRepo struct{ col *mongo.Collection }

func NewFileRepository(db *mongo.Database) FileRepository {
	return &fileRepo{col: db.Collection("files")}
}

func (r *fileRepo) Create(ctx context.Context, f *model.FileDoc) error {
	f.UploadedAt = time.Now()
	res, err := r.col.InsertOne(ctx, f)
	if err != nil { return err }
	f.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *fileRepo) FindAll(ctx context.Context) ([]model.FileDoc, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil { return nil, err }
	defer cur.Close(ctx)

	var out []model.FileDoc
	if err := cur.All(ctx, &out); err != nil { return nil, err }
	return out, nil
}

func (r *fileRepo) FindByID(ctx context.Context, id string) (*model.FileDoc, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil { return nil, err }
	var f model.FileDoc
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *fileRepo) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil { return err }
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
