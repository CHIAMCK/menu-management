package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
