// Package main provides MongoDB-related functionalities for storing logs.
package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// connectToMongoDB connects to the MongoDB and returns the client and collection.
func connectToMongoDB(ctx context.Context) (*mongo.Client, *mongo.Collection, error) {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, nil, err
	}
	collection := mongoClient.Database("logsdb").Collection("logs")
	return mongoClient, collection, nil
}
