package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"menu-management/internal/service"
)

type MenuHandler struct {
	menuService *service.MenuService
}

func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

func (h *MenuHandler) GetActiveMenu(c *gin.Context) {
	merchantID, err := strconv.ParseInt(c.Query("merchant_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "merchant_id is required and must be a positive integer"})
		return
	}

	menu, err := h.menuService.GetActiveMenu(c.Request.Context(), merchantID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMerchantID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "merchant_id is required and must be a positive integer"})
		case errors.Is(err, service.ErrActiveMenuNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "active menu not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get menu"})
		}
		return
	}

	c.JSON(http.StatusOK, menu)
}
