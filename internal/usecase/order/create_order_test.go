package order

import (
	"context"
	"testing"

	"go-order-api/internal/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrderUseCase_Execute(t *testing.T) {
	repo := new(MockOrderRepository)
	uc := NewCreateOrderUseCase(repo)

	input := CreateOrderInput{
		CustomerID: "user-123",
		Items: []CreateOrderItem{
			{ProductID: "prod-1", Quantity: 2, Price: 50},
		},
	}

	repo.On("Save", mock.Anything, mock.AnythingOfType("*order.Order")).Return(nil)

	o, err := uc.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, o)
	assert.Equal(t, "user-123", o.CustomerID)
	assert.Equal(t, int64(100), o.TotalAmount)
	assert.Equal(t, order.StatusCreated, o.Status)
	repo.AssertExpectations(t)
}
