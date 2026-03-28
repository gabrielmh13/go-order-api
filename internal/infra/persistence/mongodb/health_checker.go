package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoHealthChecker struct {
	client *mongo.Client
}

func NewMongoHealthChecker(client *mongo.Client) *MongoHealthChecker {
	return &MongoHealthChecker{
		client: client,
	}
}

func (h *MongoHealthChecker) Name() string {
	return "mongodb"
}

func (h *MongoHealthChecker) Check(ctx context.Context) error {
	return h.client.Ping(ctx, nil)
}
