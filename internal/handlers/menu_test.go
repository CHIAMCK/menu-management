package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"menu-management/internal/dto"
	"menu-management/internal/models"
	"menu-management/internal/repository"
	"menu-management/internal/service"
)

type mockMenuRepository struct {
	menu          models.Menu
	menuErr       error
	categories    []models.Category
	categoriesErr error
	items         []models.Item
	itemsErr      error
}

func (m *mockMenuRepository) FindActiveMenuByMerchantID(_ context.Context, _ int64) (models.Menu, error) {
	if m.menuErr != nil {
		return models.Menu{}, m.menuErr
	}
	return m.menu, nil
}

func (m *mockMenuRepository) FindCategoriesByMenuID(_ context.Context, _ int64) ([]models.Category, error) {
	if m.categoriesErr != nil {
		return nil, m.categoriesErr
	}
	return m.categories, nil
}

func (m *mockMenuRepository) FindItemsByCategoryIDs(_ context.Context, _ []int64) ([]models.Item, error) {
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	return m.items, nil
}

func setupMenuRouter(menuRepo *mockMenuRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := service.NewMenuService(menuRepo)
	handler := NewMenuHandler(svc)

	r := gin.New()
	r.GET("/menu", handler.GetActiveMenu)
	return r
}

func TestGetActiveMenu_InvalidMerchantID(t *testing.T) {
	r := setupMenuRouter(&mockMenuRepository{})

	for _, merchantID := range []string{"", "abc", "0", "-1"} {
		t.Run(merchantID, func(t *testing.T) {
			t.Setenv("MERCHANT_ID", merchantID)

			req := httptest.NewRequest(http.MethodGet, "/menu", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}

			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["error"] != "MERCHANT_ID environment variable is required and must be a positive integer" {
				t.Fatalf("error = %q, want MERCHANT_ID validation message", body["error"])
			}
		})
	}
}

func TestGetActiveMenu_NotFound(t *testing.T) {
	t.Setenv("MERCHANT_ID", "1")
	r := setupMenuRouter(&mockMenuRepository{
		menuErr: repository.ErrNotFound,
	})

	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] != "active menu not found" {
		t.Fatalf("error = %q, want active menu not found", body["error"])
	}
}

func TestGetActiveMenu_InternalError(t *testing.T) {
	t.Setenv("MERCHANT_ID", "1")
	r := setupMenuRouter(&mockMenuRepository{
		menuErr: errors.New("db down"),
	})

	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestGetActiveMenu_Success(t *testing.T) {
	t.Setenv("MERCHANT_ID", "1")
	r := setupMenuRouter(&mockMenuRepository{
		menu: models.Menu{ID: 1, MerchantID: 1, Name: "Lunch Menu"},
		categories: []models.Category{
			{ID: 10, Name: "Pizza"},
			{ID: 11, Name: "Sides"},
		},
		items: []models.Item{
			{ID: 1, CategoryID: 10, Name: "Margherita Pizza", Price: 12.99, Availability: models.ItemAvailabilityAvailable},
			{ID: 4, CategoryID: 11, Name: "Garlic Bread", Price: 5.99, Availability: models.ItemAvailabilityAvailable},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got dto.MenuResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.MenuID != 1 || got.MenuName != "Lunch Menu" || got.MerchantID != 1 {
		t.Fatalf("unexpected menu: %+v", got)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("len(Categories) = %d, want 2", len(got.Categories))
	}
	if len(got.Categories[0].Items) != 1 || got.Categories[0].Items[0].Name != "Margherita Pizza" {
		t.Fatalf("unexpected pizza category: %+v", got.Categories[0])
	}
	if len(got.Categories[1].Items) != 1 || got.Categories[1].Items[0].Name != "Garlic Bread" {
		t.Fatalf("unexpected sides category: %+v", got.Categories[1])
	}
}
