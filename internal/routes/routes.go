package routes

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"menu-management/internal/handlers"
	"menu-management/internal/repository"
	"menu-management/internal/service"
)

func Setup(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	menuRepo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(menuRepo)
	menuHandler := handlers.NewMenuHandler(menuService)

	itemRepo := repository.NewItemRepository(db)
	itemService := service.NewItemService(itemRepo)
	itemHandler := handlers.NewItemHandler(itemService)

	v1 := r.Group("/v1")
	{
		v1.GET("/menu", menuHandler.GetActiveMenu)
		v1.GET("/menu/items/:id", itemHandler.GetItem)
	}

	return r
}
