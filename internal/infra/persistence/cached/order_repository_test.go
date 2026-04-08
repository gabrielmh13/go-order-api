package cached

import (
	"context"
	"errors"
	"testing"
	"time"

	domaincache "go-order-api/internal/domain/cache"
	"go-order-api/internal/domain/logger"
	"go-order-api/internal/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type mockOrderRepository struct {
	mock.Mock
}

func (m *mockOrderRepository) Save(ctx context.Context, o *order.Order) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}

func (m *mockOrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *mockOrderRepository) UpdateStatus(ctx context.Context, id string, status order.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type noopLogger struct{}

func (n *noopLogger) Info(msg string, fields ...logger.Field)             {}
func (n *noopLogger) Error(msg string, err error, fields ...logger.Field) {}
func (n *noopLogger) Debug(msg string, fields ...logger.Field)            {}
func (n *noopLogger) Warning(msg string, fields ...logger.Field)          {}
func (n *noopLogger) Fatal(msg string, err error, fields ...logger.Field) {}

var _ domaincache.ICache = (*mockCache)(nil)

func TestOrderRepository_GetByID_ReturnsCachedOrder(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockOrderRepository)
	cachedRepo := NewOrderRepository(repo, cache, 5*time.Minute, &noopLogger{})

	cache.On("Get", mock.Anything, "order:order-1").Return(`{"id":"order-1","customerId":"cust-1","items":[{"productId":"p1","quantity":1,"price":100}],"totalAmount":100,"status":"created","createdAt":"2026-04-07T12:00:00Z","updatedAt":"2026-04-07T12:00:00Z"}`, nil).Once()

	result, err := cachedRepo.GetByID(context.Background(), "order-1")

	assert.NoError(t, err)
	assert.Equal(t, "order-1", result.ID)
	assert.Equal(t, "cust-1", result.CustomerID)
	repo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	cache.AssertExpectations(t)
}

func TestOrderRepository_GetByID_LoadsFromRepositoryAndCaches(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockOrderRepository)
	cachedRepo := NewOrderRepository(repo, cache, 5*time.Minute, &noopLogger{})

	o := &order.Order{ID: "order-1", CustomerID: "cust-1", Status: order.StatusCreated}

	cache.On("Get", mock.Anything, "order:order-1").Return("", nil).Once()
	repo.On("GetByID", mock.Anything, "order-1").Return(o, nil).Once()
	cache.On("Set", mock.Anything, "order:order-1", mock.AnythingOfType("string"), 5*time.Minute).Return(nil).Once()

	result, err := cachedRepo.GetByID(context.Background(), "order-1")

	assert.NoError(t, err)
	assert.Equal(t, o, result)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestOrderRepository_GetByID_FallsBackWhenCacheFails(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockOrderRepository)
	cachedRepo := NewOrderRepository(repo, cache, 5*time.Minute, &noopLogger{})

	o := &order.Order{ID: "order-1", CustomerID: "cust-1", Status: order.StatusCreated}

	cache.On("Get", mock.Anything, "order:order-1").Return("", errors.New("redis down")).Once()
	repo.On("GetByID", mock.Anything, "order-1").Return(o, nil).Once()
	cache.On("Set", mock.Anything, "order:order-1", mock.AnythingOfType("string"), 5*time.Minute).Return(nil).Once()

	result, err := cachedRepo.GetByID(context.Background(), "order-1")

	assert.NoError(t, err)
	assert.Equal(t, o, result)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestOrderRepository_UpdateStatus_InvalidatesCache(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockOrderRepository)
	cachedRepo := NewOrderRepository(repo, cache, 5*time.Minute, &noopLogger{})

	repo.On("UpdateStatus", mock.Anything, "order-1", order.StatusProcessing).Return(nil).Once()
	cache.On("Delete", mock.Anything, "order:order-1").Return(nil).Once()

	err := cachedRepo.UpdateStatus(context.Background(), "order-1", order.StatusProcessing)

	assert.NoError(t, err)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestOrderRepository_GetByID_WithoutCache_UsesRepository(t *testing.T) {
	repo := new(mockOrderRepository)
	cachedRepo := NewOrderRepository(repo, nil, 5*time.Minute, &noopLogger{})

	o := &order.Order{ID: "order-1", CustomerID: "cust-1", Status: order.StatusCreated}

	repo.On("GetByID", mock.Anything, "order-1").Return(o, nil).Once()

	result, err := cachedRepo.GetByID(context.Background(), "order-1")

	assert.NoError(t, err)
	assert.Equal(t, o, result)
	repo.AssertExpectations(t)
}

func TestOrderRepository_UpdateStatus_WithoutCache_UsesRepository(t *testing.T) {
	repo := new(mockOrderRepository)
	cachedRepo := NewOrderRepository(repo, nil, 5*time.Minute, &noopLogger{})

	repo.On("UpdateStatus", mock.Anything, "order-1", order.StatusProcessing).Return(nil).Once()

	err := cachedRepo.UpdateStatus(context.Background(), "order-1", order.StatusProcessing)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
