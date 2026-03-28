package order

import (
	"context"
	"testing"

	"go-order-api/internal/domain/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOrderUseCase_Execute(t *testing.T) {
	repo := new(MockOrderRepository)
	uc := NewGetOrderUseCase(repo)

	o := &order.Order{ID: "order-1", CustomerID: "user-1"}
	repo.On("GetByID", mock.Anything, "order-1").Return(o, nil)

	result, err := uc.Execute(context.Background(), "order-1")

	assert.NoError(t, err)
	assert.Equal(t, o, result)
	repo.AssertExpectations(t)
}
