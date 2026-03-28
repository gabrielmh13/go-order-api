package order

import (
	"time"
)

type OrderStatusChangedEvent struct {
	OrderID   string      `json:"orderId"`
	OldStatus OrderStatus `json:"oldStatus"`
	NewStatus OrderStatus `json:"newStatus"`
	Timestamp time.Time   `json:"timestamp"`
}
