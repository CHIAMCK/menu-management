package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"menu-management/internal/dto"
	"menu-management/internal/service"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		case errors.Is(err, service.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderRequest):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order request"})
		case errors.Is(err, service.ErrDuplicateOrderItem):
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate item in order"})
		case errors.Is(err, service.ErrItemNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		case errors.Is(err, service.ErrItemUnavailable):
			c.JSON(http.StatusBadRequest, gin.H{"error": "item unavailable"})
		case errors.Is(err, service.ErrItemMerchantMismatch):
			c.JSON(http.StatusBadRequest, gin.H{"error": "item does not belong to merchant"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		}
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be RECEIVED, PREPARING, READY, or COMPLETED"})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		case errors.Is(err, service.ErrInvalidOrderStatus):
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be RECEIVED, PREPARING, READY, or COMPLETED"})
		case errors.Is(err, service.ErrInvalidOrderStatusTransition):
			c.JSON(http.StatusConflict, gin.H{"error": "invalid order status transition"})
		case errors.Is(err, service.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}
