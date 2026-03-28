package mongodb

import (
	"context"
	"time"

	"go-order-api/internal/domain/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewConnection(ctx context.Context, uri string, l logger.Logger) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		l.Error("Failed to connect to MongoDB", err, logger.Any("uri", uri))
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		l.Error("Failed to ping MongoDB", err)
		return nil, err
	}

	l.Info("Connected to MongoDB successfully")
	return client, nil
}
