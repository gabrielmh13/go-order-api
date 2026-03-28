package order

import (
	"context"

	"go-order-api/internal/domain/order"
)

type GetOrderUseCase struct {
	orderRepo order.IOrderRepository
}

func NewGetOrderUseCase(repo order.IOrderRepository) *GetOrderUseCase {
	return &GetOrderUseCase{
		orderRepo: repo,
	}
}

func (uc *GetOrderUseCase) Execute(ctx context.Context, id string) (*order.Order, error) {
	return uc.orderRepo.GetByID(ctx, id)
}
