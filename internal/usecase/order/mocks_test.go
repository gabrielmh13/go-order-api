package order

import (
	"context"

	"go-order-api/internal/domain/order"

	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Save(ctx context.Context, o *order.Order) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id string, status order.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockMessageBroker struct {
	mock.Mock
}

func (m *MockMessageBroker) Publish(ctx context.Context, queue string, body any) error {
	args := m.Called(ctx, queue, body)
	return args.Error(0)
}
