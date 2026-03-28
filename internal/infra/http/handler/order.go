package handler

import (
	"errors"
	"net/http"

	"go-order-api/internal/domain/logger"
	"go-order-api/internal/domain/order"
	usecase "go-order-api/internal/usecase/order"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	createOrderUC       usecase.ICreateOrderUseCase
	getOrderUC          usecase.IGetOrderUseCase
	updateOrderStatusUC usecase.IUpdateOrderStatusUseCase
	logger              logger.Logger
}

func NewOrderHandler(
	createUC usecase.ICreateOrderUseCase,
	getUC usecase.IGetOrderUseCase,
	updateUC usecase.IUpdateOrderStatusUseCase,
	l logger.Logger,
) *OrderHandler {
	return &OrderHandler{
		createOrderUC:       createUC,
		getOrderUC:          getUC,
		updateOrderStatusUC: updateUC,
		logger:              l,
	}
}

func (h *OrderHandler) SetupRoutes(router *gin.Engine) {
	router.POST("/orders", h.CreateOrder)
	router.GET("/orders/:id", h.GetOrder)
	router.PATCH("/orders/:id/status", h.UpdateStatus)
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Creates a new order with the given items and customer ID.
// @Tags orders
// @Accept json
// @Produce json
// @Param order body usecase.CreateOrderInput true "Order creation payload"
// @Success 201 {object} order.Order
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var input usecase.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	o, err := h.createOrderUC.Execute(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create order", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Order created successfully", logger.Any("orderId", o.ID))
	c.JSON(http.StatusCreated, o)
}

// GetOrder godoc
// @Summary Get an order by ID
// @Description Retrieves details of a specific order.
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} order.Order
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	o, err := h.getOrderUC.Execute(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, order.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, o)
}

type updateStatusRequest struct {
	Status order.OrderStatus `json:"status" binding:"required" enums:"created,processing,shipped,delivered"`
}

// UpdateStatus godoc
// @Summary Update order status
// @Description Updates the status of an existing order.
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param status body updateStatusRequest true "New status payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var input updateStatusRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ucInput := usecase.UpdateOrderStatusInput{
		ID:     id,
		Status: input.Status,
	}

	err := h.updateOrderStatusUC.Execute(c.Request.Context(), ucInput)
	if err != nil {
		if errors.Is(err, order.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		if errors.Is(err, order.ErrInvalidStatusTransition) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("Failed to update order status", err, logger.Any("orderId", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	h.logger.Info("Order status updated successfully", logger.Any("orderId", id), logger.Any("status", string(input.Status)))
	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}
