package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"menu-management/internal/dto"
	"menu-management/internal/service"
)

type ItemHandler struct {
	itemService *service.ItemService
}

func NewItemHandler(itemService *service.ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) GetItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		return
	}

	item, err := h.itemService.GetItemByID(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		case errors.Is(err, service.ErrItemNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get item"})
		}
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *ItemHandler) UpdateItemAvailability(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		return
	}

	var req dto.UpdateItemAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "availability is required"})
		return
	}

	item, err := h.itemService.UpdateItemAvailability(c.Request.Context(), id, req.Availability)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		case errors.Is(err, service.ErrInvalidItemAvailability):
			c.JSON(http.StatusBadRequest, gin.H{"error": "availability must be AVAILABLE or OUT_OF_STOCK"})
		case errors.Is(err, service.ErrItemNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update item availability"})
		}
		return
	}

	c.JSON(http.StatusOK, item)
}
