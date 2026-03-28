package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-order-api/internal/domain/order"
	usecase "go-order-api/internal/usecase/order"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCreateOrderUseCase struct {
	mock.Mock
}

func (m *MockCreateOrderUseCase) Execute(ctx context.Context, input usecase.CreateOrderInput) (*order.Order, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

type MockGetOrderUseCase struct {
	mock.Mock
}

func (m *MockGetOrderUseCase) Execute(ctx context.Context, id string) (*order.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

type MockUpdateOrderStatusUseCase struct {
	mock.Mock
}

func (m *MockUpdateOrderStatusUseCase) Execute(ctx context.Context, input usecase.UpdateOrderStatusInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func TestOrderHandler_CreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 201 when order is created", func(t *testing.T) {
		createUC := new(MockCreateOrderUseCase)
		getUC := new(MockGetOrderUseCase)
		updateUC := new(MockUpdateOrderStatusUseCase)
		l := new(MockLogger)

		h := NewOrderHandler(createUC, getUC, updateUC, l)
		router := gin.New()
		h.SetupRoutes(router)

		input := usecase.CreateOrderInput{
			CustomerID: "cust-1",
			Items: []usecase.CreateOrderItem{
				{ProductID: "p1", Quantity: 1, Price: 100},
			},
		}
		expectedOrder, _ := order.NewOrder("cust-1", []order.OrderItem{{ProductID: "p1", Quantity: 1, Price: 100}})
		createUC.On("Execute", mock.Anything, input).Return(expectedOrder, nil)

		body, _ := json.Marshal(input)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		createUC.AssertExpectations(t)
	})

	t.Run("Should return 400 when input is invalid", func(t *testing.T) {
		h := NewOrderHandler(nil, nil, nil, nil)
		router := gin.New()
		h.SetupRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBufferString(`{invalid-json}`))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOrderHandler_GetOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 200 when order exists", func(t *testing.T) {
		getUC := new(MockGetOrderUseCase)
		h := NewOrderHandler(nil, getUC, nil, nil)
		router := gin.New()
		h.SetupRoutes(router)

		expectedOrder, _ := order.NewOrder("cust-1", []order.OrderItem{{ProductID: "p1", Quantity: 1, Price: 100}})
		getUC.On("Execute", mock.Anything, "test-id").Return(expectedOrder, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/orders/test-id", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		getUC.AssertExpectations(t)
	})

	t.Run("Should return 404 when order not found", func(t *testing.T) {
		getUC := new(MockGetOrderUseCase)
		h := NewOrderHandler(nil, getUC, nil, nil)
		router := gin.New()
		h.SetupRoutes(router)

		getUC.On("Execute", mock.Anything, "not-found").Return(nil, order.ErrOrderNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/orders/not-found", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestOrderHandler_UpdateStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 200 when status is updated", func(t *testing.T) {
		updateUC := new(MockUpdateOrderStatusUseCase)
		l := new(MockLogger)
		h := NewOrderHandler(nil, nil, updateUC, l)
		router := gin.New()
		h.SetupRoutes(router)

		input := usecase.UpdateOrderStatusInput{ID: "id-1", Status: order.StatusProcessing}
		updateUC.On("Execute", mock.Anything, input).Return(nil)

		body, _ := json.Marshal(map[string]string{"status": "processing"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/orders/id-1/status", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		updateUC.AssertExpectations(t)
	})

	t.Run("Should return 422 for invalid transition", func(t *testing.T) {
		updateUC := new(MockUpdateOrderStatusUseCase)
		h := NewOrderHandler(nil, nil, updateUC, nil)
		router := gin.New()
		h.SetupRoutes(router)

		updateUC.On("Execute", mock.Anything, mock.Anything).Return(order.ErrInvalidStatusTransition)

		body, _ := json.Marshal(map[string]string{"status": "shipped"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/orders/id-1/status", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
