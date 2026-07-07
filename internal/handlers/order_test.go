package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"menu-management/internal/dto"
	"menu-management/internal/models"
	"menu-management/internal/repository"
	"menu-management/internal/service"
)

type mockOrderRepository struct {
	order    models.Order
	orderErr error
	items    []repository.OrderItemWithItem
	itemsErr error
}

func (m *mockOrderRepository) FindOrderByID(_ context.Context, _ int64) (models.Order, error) {
	if m.orderErr != nil {
		return models.Order{}, m.orderErr
	}
	return m.order, nil
}

func (m *mockOrderRepository) FindOrderItemsByOrderID(_ context.Context, _ int64) ([]repository.OrderItemWithItem, error) {
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	return m.items, nil
}

func setupOrderRouter(repo *mockOrderRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := service.NewOrderService(repo)
	handler := NewOrderHandler(svc)

	r := gin.New()
	r.GET("/orders/:id", handler.GetOrder)
	return r
}

func TestGetOrder_InvalidID(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{})

	for _, path := range []string{"/orders/abc", "/orders/0", "/orders/-1"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("GET %s: status = %d, want %d", path, rec.Code, http.StatusBadRequest)
		}

		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("GET %s: decode response: %v", path, err)
		}
		if body["error"] != "id must be a positive integer" {
			t.Fatalf("GET %s: error = %q, want id must be a positive integer", path, body["error"])
		}
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{
		orderErr: repository.ErrNotFound,
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/99", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] != "order not found" {
		t.Fatalf("error = %q, want order not found", body["error"])
	}
}

func TestGetOrder_InternalError(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{
		orderErr: errors.New("db down"),
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestGetOrder_Success(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC)

	r := setupOrderRouter(&mockOrderRepository{
		order: models.Order{
			ID:          1,
			UserID:      1,
			MerchantID:  1,
			Status:      models.OrderStatusCompleted,
			TotalAmount: 31.97,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
		items: []repository.OrderItemWithItem{
			{ID: 1, ItemID: 1, Name: "Margherita Pizza", Quantity: 2, UnitPrice: 12.99},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got dto.OrderDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != 1 || got.Status != "COMPLETED" || got.TotalAmount != 31.97 {
		t.Fatalf("unexpected order: %+v", got)
	}
	if len(got.Items) != 1 || got.Items[0].Name != "Margherita Pizza" {
		t.Fatalf("unexpected items: %+v", got.Items)
	}
}
