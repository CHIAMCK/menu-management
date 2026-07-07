package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
