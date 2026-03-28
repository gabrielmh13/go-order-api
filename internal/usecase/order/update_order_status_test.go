package order

import (
	"context"
	"testing"

	"go-order-api/internal/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateOrderStatusUseCase_Execute(t *testing.T) {
	repo := new(MockOrderRepository)
	broker := new(MockMessageBroker)
	usecase := NewUpdateOrderStatusUseCase(repo, broker)

	ctx := context.Background()
	id := "test-id"
	o := order.NewOrder(id, []order.OrderItem{{ProductID: "p1", Quantity: 1, Price: 100}})
	o.Status = order.StatusCreated

	input := UpdateOrderStatusInput{
		ID:     id,
		Status: order.StatusProcessing,
	}

	repo.On("GetByID", ctx, id).Return(o, nil)
	repo.On("UpdateStatus", ctx, id, order.StatusProcessing).Return(nil)
	broker.On("Publish", ctx, "order.status.changed", mock.Anything).Return(nil)

	err := usecase.Execute(ctx, input)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	broker.AssertExpectations(t)
}

func TestUpdateOrderStatusUseCase_InvalidTransition(t *testing.T) {
	repo := new(MockOrderRepository)
	broker := new(MockMessageBroker)
	usecase := NewUpdateOrderStatusUseCase(repo, broker)

	ctx := context.Background()
	id := "test-id"
	o := order.NewOrder(id, []order.OrderItem{{ProductID: "p1", Quantity: 1, Price: 100}})
	o.Status = order.StatusCreated

	input := UpdateOrderStatusInput{
		ID:     id,
		Status: order.StatusDelivered,
	}

	repo.On("GetByID", ctx, id).Return(o, nil)

	err := usecase.Execute(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, order.ErrInvalidStatusTransition, err)
	repo.AssertExpectations(t)
	broker.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}
