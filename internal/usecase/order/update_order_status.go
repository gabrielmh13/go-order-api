package order

import (
	"context"
	"time"

	"go-order-api/internal/domain/broker"
	"go-order-api/internal/domain/order"
)

type UpdateOrderStatusInput struct {
	ID     string            `json:"id"`
	Status order.OrderStatus `json:"status"`
}

type UpdateOrderStatusUseCase struct {
	repo   order.IOrderRepository
	broker broker.IMessageBroker
}

type IUpdateOrderStatusUseCase interface {
	Execute(ctx context.Context, input UpdateOrderStatusInput) error
}

func NewUpdateOrderStatusUseCase(r order.IOrderRepository, b broker.IMessageBroker) *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{
		repo:   r,
		broker: b,
	}
}

func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, input UpdateOrderStatusInput) error {
	o, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}

	oldStatus := o.Status
	if err := o.UpdateStatus(input.Status); err != nil {
		return err
	}

	if err := uc.repo.UpdateStatus(ctx, input.ID, input.Status); err != nil {
		return err
	}

	event := order.OrderStatusChangedEvent{
		OrderID:   o.ID,
		OldStatus: oldStatus,
		NewStatus: o.Status,
		Timestamp: time.Now(),
	}

	return uc.broker.Publish(ctx, "order.status.changed", event)
}
