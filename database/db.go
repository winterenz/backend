package database

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() (*mongo.Client, *mongo.Database, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		if alt := os.Getenv("MONGO_DB_NAME"); alt != "" {
			dbName = alt
		} else {
			dbName = "alumnidb"
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}
	// optional: ping
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, err
	}
	return client, client.Database(dbName), nil
}
