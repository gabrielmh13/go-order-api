package handler

import (
	"context"
	"net/http"
	"time"

	"go-order-api/internal/domain/health"
	"go-order-api/internal/domain/logger"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	checkers []health.HealthChecker
	logger   logger.Logger
}

func NewHealthHandler(l logger.Logger, checkers ...health.HealthChecker) *HealthHandler {
	return &HealthHandler{
		checkers: checkers,
		logger:   l,
	}
}

func (h *HealthHandler) SetupRoutes(router *gin.Engine) {
	router.GET("/health", h.HealthCheck)
}

// HealthCheck godoc
// @Summary API Healthcheck
// @Description Validates the health of the application and its dependencies (MongoDB, RabbitMQ).
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	overallStatus := "healthy"
	httpStatus := http.StatusOK
	services := make(map[string]string)

	for _, checker := range h.checkers {
		name := checker.Name()
		if err := checker.Check(ctx); err != nil {
			h.logger.Error("Healthcheck failed for service: "+name, err)
			services[name] = "down"
			overallStatus = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		} else {
			services[name] = "up"
		}
	}

	c.JSON(httpStatus, gin.H{
		"status":   overallStatus,
		"services": services,
	})
}
