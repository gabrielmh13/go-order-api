package order

import (
	"context"

	"go-order-api/internal/domain/order"
)

type CreateOrderInput struct {
	CustomerID string            `json:"customerId" binding:"required"`
	Items      []CreateOrderItem `json:"items" binding:"required,gt=0,dive"`
}

type CreateOrderItem struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
	Price     int64  `json:"price" binding:"required,gt=0"`
}

type CreateOrderUseCase struct {
	orderRepo order.IOrderRepository
}

type ICreateOrderUseCase interface {
	Execute(ctx context.Context, input CreateOrderInput) (*order.Order, error)
}

func NewCreateOrderUseCase(repo order.IOrderRepository) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo: repo,
	}
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (*order.Order, error) {
	items := make([]order.OrderItem, len(input.Items))
	for i, item := range input.Items {
		items[i] = order.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	newOrder, err := order.NewOrder(input.CustomerID, items)
	if err != nil {
		return nil, err
	}

	if err := uc.orderRepo.Save(ctx, newOrder); err != nil {
		return nil, err
	}

	return newOrder, nil
}
