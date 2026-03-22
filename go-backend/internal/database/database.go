package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func Connect(mongoURI, dbName string) (*mongo.Client, error) {
	// Configure BSON options to decode ObjectIDs as hex strings
	bsonOpts := &options.BSONOptions{
		ObjectIDAsHexString: true,
	}

	// Connect to MongoDB with BSON options
	clientOpts := options.Client().
		ApplyURI(mongoURI).
		SetBSONOptions(bsonOpts)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	// Ping the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("error pinging MongoDB: %w", err)
	}

	log.Println("MongoDB connected successfully")

	// Set global client and database
	Client = client
	DB = client.Database(dbName)

	return client, nil
}

func Disconnect() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return Client.Disconnect(ctx)
	}
	return nil
}
