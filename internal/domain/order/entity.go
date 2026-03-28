package order

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusCreated    OrderStatus = "created"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrOrderNotFound           = errors.New("order not found")
	ErrEmptyOrder              = errors.New("order must have at least one item")
	ErrInvalidItemPrice        = errors.New("item price cannot be negative")
)

type OrderItem struct {
	ProductID string `json:"productId" bson:"productId"`
	Quantity  int    `json:"quantity" bson:"quantity"`
	Price     int64  `json:"price" bson:"price"`
}

type Order struct {
	ID          string      `json:"id" bson:"_id"`
	CustomerID  string      `json:"customerId" bson:"customerId"`
	Items       []OrderItem `json:"items" bson:"items"`
	TotalAmount int64       `json:"totalAmount" bson:"totalAmount"`
	Status      OrderStatus `json:"status" bson:"status"`
	CreatedAt   time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt" bson:"updatedAt"`
}

func NewOrder(customerID string, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrder
	}

	var total int64
	for _, item := range items {
		if item.Price < 0 {
			return nil, ErrInvalidItemPrice
		}
		total += item.Price * int64(item.Quantity)
	}

	now := time.Now()
	return &Order{
		ID:          uuid.New().String(),
		CustomerID:  customerID,
		Items:       items,
		TotalAmount: total,
		Status:      StatusCreated,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (o *Order) UpdateStatus(newStatus OrderStatus) error {
	if !o.isValidTransition(newStatus) {
		return ErrInvalidStatusTransition
	}
	o.Status = newStatus
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) isValidTransition(newStatus OrderStatus) bool {
	switch o.Status {
	case StatusCreated:
		return newStatus == StatusProcessing
	case StatusProcessing:
		return newStatus == StatusShipped
	case StatusShipped:
		return newStatus == StatusDelivered
	default:
		return false
	}
}
