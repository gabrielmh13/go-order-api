package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-order-api/internal/domain/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHealthChecker struct {
	mock.Mock
}

func (m *MockHealthChecker) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockHealthChecker) Check(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...logger.Field)             {}
func (m *MockLogger) Error(msg string, err error, fields ...logger.Field) {}
func (m *MockLogger) Debug(msg string, fields ...logger.Field)            {}
func (m *MockLogger) Warning(msg string, fields ...logger.Field)          {}
func (m *MockLogger) Fatal(msg string, err error, fields ...logger.Field) {}

func TestHealthHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 200 when all services are healthy", func(t *testing.T) {
		checker := new(MockHealthChecker)
		checker.On("Name").Return("mongodb")
		checker.On("Check", mock.Anything).Return(nil)

		l := new(MockLogger)
		h := NewHealthHandler(l, checker)

		router := gin.New()
		h.SetupRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"healthy"`)
		assert.Contains(t, w.Body.String(), `"mongodb":"up"`)
		checker.AssertExpectations(t)
	})

	t.Run("Should return 503 when a service is unhealthy", func(t *testing.T) {
		checker := new(MockHealthChecker)
		checker.On("Name").Return("rabbitmq")
		checker.On("Check", mock.Anything).Return(errors.New("connection failed"))

		l := new(MockLogger)
		h := NewHealthHandler(l, checker)

		router := gin.New()
		h.SetupRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"unhealthy"`)
		assert.Contains(t, w.Body.String(), `"rabbitmq":"down"`)
		checker.AssertExpectations(t)
	})
}
