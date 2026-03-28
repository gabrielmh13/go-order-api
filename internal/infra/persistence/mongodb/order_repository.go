package mongodb

import (
	"context"
	"errors"

	"go-order-api/internal/domain/logger"
	"go-order-api/internal/domain/order"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
	logger     logger.Logger
}

func NewOrderRepository(db *mongo.Database, l logger.Logger) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
		logger:     l,
	}
}

func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
	_, err := r.collection.InsertOne(ctx, o)
	if err != nil {
		r.logger.Error("Failed to save order in MongoDB", err, logger.Any("orderId", o.ID))
		return err
	}
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	var o order.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&o)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, order.ErrOrderNotFound
		}
		r.logger.Error("Failed to find order in MongoDB", err, logger.Any("orderId", id))
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status order.OrderStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
		"$currentDate": bson.M{
			"updatedAt": true,
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update status in MongoDB", err, logger.Any("orderId", id))
		return err
	}
	if result.MatchedCount == 0 {
		return order.ErrOrderNotFound
	}
	return nil
}
