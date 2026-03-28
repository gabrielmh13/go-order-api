package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOrder(t *testing.T) {
	customerID := "customer-123"
	items := []OrderItem{
		{ProductID: "prod-1", Quantity: 2, Price: 100}, // 200
		{ProductID: "prod-2", Quantity: 1, Price: 50},  // 50
	}

	o, err := NewOrder(customerID, items)

	assert.NoError(t, err)
	assert.NotEmpty(t, o.ID)
	assert.Equal(t, customerID, o.CustomerID)
	assert.Equal(t, items, o.Items)
	assert.Equal(t, int64(250), o.TotalAmount)
	assert.Equal(t, StatusCreated, o.Status)
	assert.NotZero(t, o.CreatedAt)
	assert.NotZero(t, o.UpdatedAt)
	assert.True(t, o.CreatedAt.Equal(o.UpdatedAt))
}

func TestNewOrder_Validations(t *testing.T) {
	t.Run("Should fail if items are empty", func(t *testing.T) {
		o, err := NewOrder("cust-1", []OrderItem{})
		assert.ErrorIs(t, err, ErrEmptyOrder)
		assert.Nil(t, o)
	})

	t.Run("Should fail if any item price is negative", func(t *testing.T) {
		items := []OrderItem{
			{ProductID: "p1", Quantity: 1, Price: -10},
		}
		o, err := NewOrder("cust-1", items)
		assert.ErrorIs(t, err, ErrInvalidItemPrice)
		assert.Nil(t, o)
	})
}

func TestOrder_UpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus OrderStatus
		newStatus     OrderStatus
		expectError   error
	}{
		{
			name:          "Should transition from Created to Processing",
			initialStatus: StatusCreated,
			newStatus:     StatusProcessing,
			expectError:   nil,
		},
		{
			name:          "Should transition from Processing to Shipped",
			initialStatus: StatusProcessing,
			newStatus:     StatusShipped,
			expectError:   nil,
		},
		{
			name:          "Should transition from Shipped to Delivered",
			initialStatus: StatusShipped,
			newStatus:     StatusDelivered,
			expectError:   nil,
		},
		{
			name:          "Should fail transition from Created to Shipped",
			initialStatus: StatusCreated,
			newStatus:     StatusShipped,
			expectError:   ErrInvalidStatusTransition,
		},
		{
			name:          "Should fail transition from Delivered to Processing",
			initialStatus: StatusDelivered,
			newStatus:     StatusProcessing,
			expectError:   ErrInvalidStatusTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				Status:    tt.initialStatus,
				UpdatedAt: time.Now().Add(-time.Hour),
			}
			oldUpdatedAt := o.UpdatedAt

			err := o.UpdateStatus(tt.newStatus)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Equal(t, tt.initialStatus, o.Status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, o.Status)
				assert.True(t, o.UpdatedAt.After(oldUpdatedAt))
			}
		})
	}
}
