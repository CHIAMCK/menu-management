package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"menu-management/internal/dto"
	"menu-management/internal/lock"
	"menu-management/internal/models"
	"menu-management/internal/repository"
	"menu-management/internal/service"
)

type mockOrderRepository struct {
	order       models.Order
	orderErr    error
	items       []repository.OrderItemWithItem
	itemsErr    error
	createInput repository.CreateOrderInput
	createID    int64
	createErr   error
	updateInput struct {
		id     int64
		status models.OrderStatus
	}
	updatedOrder models.Order
	updateErr    error
}

func (m *mockOrderRepository) FindOrderByID(_ context.Context, id int64) (models.Order, error) {
	if m.orderErr != nil {
		return models.Order{}, m.orderErr
	}
	if m.order.ID == 0 && m.createID != 0 && id == m.createID {
		return models.Order{
			ID:          m.createID,
			UserID:      m.createInput.UserID,
			MerchantID:  m.createInput.MerchantID,
			Status:      models.OrderStatusReceived,
			TotalAmount: m.createInput.TotalAmount,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	}
	return m.order, nil
}

func (m *mockOrderRepository) FindOrderItemsByOrderID(_ context.Context, orderID int64) ([]repository.OrderItemWithItem, error) {
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	if len(m.items) > 0 {
		return m.items, nil
	}
	if m.createID != 0 && orderID == m.createID {
		items := make([]repository.OrderItemWithItem, 0, len(m.createInput.Items))
		for i, item := range m.createInput.Items {
			items = append(items, repository.OrderItemWithItem{
				ID:        int64(i + 1),
				ItemID:    item.ItemID,
				Name:      "Margherita Pizza",
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
			})
		}
		return items, nil
	}
	return m.items, nil
}

func (m *mockOrderRepository) CreateOrder(_ context.Context, input repository.CreateOrderInput) (int64, error) {
	m.createInput = input
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 99
	}
	return m.createID, nil
}

func (m *mockOrderRepository) UpdateOrderStatus(_ context.Context, id int64, status models.OrderStatus) (models.Order, error) {
	m.updateInput.id = id
	m.updateInput.status = status
	if m.updateErr != nil {
		return models.Order{}, m.updateErr
	}
	if m.updatedOrder.ID != 0 {
		return m.updatedOrder, nil
	}
	order := m.order
	order.Status = status
	return order, nil
}

type mockItemRepository struct {
	items    []models.Item
	itemsErr error
}

func (m *mockItemRepository) FindItemByID(_ context.Context, id int64) (models.Item, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return models.Item{}, repository.ErrNotFound
}

func (m *mockItemRepository) FindItemsByIDs(_ context.Context, ids []int64) ([]models.Item, error) {
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}

	itemsByID := make(map[int64]models.Item, len(m.items))
	for _, item := range m.items {
		itemsByID[item.ID] = item
	}

	found := make([]models.Item, 0, len(ids))
	for _, id := range ids {
		if item, ok := itemsByID[id]; ok {
			found = append(found, item)
		}
	}
	return found, nil
}

func (m *mockItemRepository) UpdateItemAvailability(_ context.Context, _ int64, _ models.ItemAvailability) (models.Item, error) {
	return models.Item{}, errors.New("not implemented")
}

func setupOrderRouter(orderRepo *mockOrderRepository, itemRepo *mockItemRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	userLocker := lock.NewInMemoryUserLocker(5 * time.Second)
	svc := service.NewOrderService(orderRepo, itemRepo, nil, userLocker)
	handler := NewOrderHandler(svc)

	r := gin.New()
	r.GET("/orders/:id", handler.GetOrder)
	r.POST("/orders", handler.CreateOrder)
	r.PATCH("/orders/:id/status", handler.UpdateOrderStatus)
	return r
}

func TestGetOrder_InvalidID(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{}, &mockItemRepository{})

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
	}, &mockItemRepository{})

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
	}, &mockItemRepository{})

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
	}, &mockItemRepository{})

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

	if got.OrderID != 1 || got.Status != "COMPLETED" || got.TotalAmount != 31.97 {
		t.Fatalf("unexpected order: %+v", got)
	}
	if len(got.Items) != 1 || got.Items[0].Name != "Margherita Pizza" {
		t.Fatalf("unexpected items: %+v", got.Items)
	}
}

func TestCreateOrder_Success(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{createID: 99}, &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
			{ID: 4, MerchantID: 1, Price: 5.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	})

	body := `{"user_id":1,"merchant_id":1,"items":[{"item_id":1,"quantity":2},{"item_id":4,"quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var got dto.OrderDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	wantTotal := 12.99*2 + 5.99
	if got.OrderID != 99 || got.Status != "RECEIVED" || got.TotalAmount != wantTotal {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestCreateOrder_InvalidBody(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{}, &mockItemRepository{})

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`{"user_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateOrder_ItemNotFound(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{}, &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	})

	body := `{"user_id":1,"merchant_id":1,"items":[{"item_id":1,"quantity":1},{"item_id":99,"quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCreateOrder_ItemUnavailable(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{}, &mockItemRepository{
		items: []models.Item{
			{ID: 5, MerchantID: 1, Price: 2.50, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityOutOfStock},
		},
	})

	body := `{"user_id":1,"merchant_id":1,"items":[{"item_id":5,"quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateOrder_UserLocked(t *testing.T) {
	userLocker := lock.NewInMemoryUserLocker(5 * time.Second)
	if err := userLocker.TryLock(1); err != nil {
		t.Fatalf("TryLock: unexpected error: %v", err)
	}

	gin.SetMode(gin.TestMode)
	svc := service.NewOrderService(&mockOrderRepository{createID: 99}, &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	}, nil, userLocker)
	handler := NewOrderHandler(svc)

	r := gin.New()
	r.POST("/orders", handler.CreateOrder)

	body := `{"user_id":1,"merchant_id":1,"items":[{"item_id":1,"quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] != "order already in progress, please wait" {
		t.Fatalf("error = %q, want order already in progress, please wait", resp["error"])
	}
}

func TestUpdateOrderStatus_InvalidID(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{}, &mockItemRepository{})

	body := `{"status":"PREPARING"}`
	req := httptest.NewRequest(http.MethodPatch, "/orders/abc/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateOrderStatus_InvalidBody(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{}, &mockItemRepository{})

	req := httptest.NewRequest(http.MethodPatch, "/orders/1/status", bytes.NewBufferString(`{"status":"SHIPPED"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateOrderStatus_InvalidTransition(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{
		order: models.Order{ID: 1, Status: models.OrderStatusReceived},
	}, &mockItemRepository{})

	body := `{"status":"READY"}`
	req := httptest.NewRequest(http.MethodPatch, "/orders/1/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestUpdateOrderStatus_NotFound(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{
		orderErr: repository.ErrNotFound,
	}, &mockItemRepository{})

	body := `{"status":"PREPARING"}`
	req := httptest.NewRequest(http.MethodPatch, "/orders/99/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdateOrderStatus_Success(t *testing.T) {
	r := setupOrderRouter(&mockOrderRepository{
		order: models.Order{
			ID:          2,
			UserID:      2,
			MerchantID:  1,
			Status:      models.OrderStatusReceived,
			TotalAmount: 14.99,
		},
		updatedOrder: models.Order{
			ID:          2,
			UserID:      2,
			MerchantID:  1,
			Status:      models.OrderStatusPreparing,
			TotalAmount: 14.99,
		},
		items: []repository.OrderItemWithItem{
			{ID: 3, ItemID: 2, Name: "Pepperoni Pizza", Quantity: 1, UnitPrice: 14.99},
		},
	}, &mockItemRepository{})

	body := `{"status":"PREPARING"}`
	req := httptest.NewRequest(http.MethodPatch, "/orders/2/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got dto.OrderDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.OrderID != 2 || got.Status != "PREPARING" {
		t.Fatalf("unexpected order: %+v", got)
	}
}
