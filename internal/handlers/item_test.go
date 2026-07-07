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
	"menu-management/internal/models"
	"menu-management/internal/repository"
	"menu-management/internal/service"
)

type mockItemHandlerRepository struct {
	items       []models.Item
	findErr     error
	updateInput struct {
		id           int64
		availability models.ItemAvailability
	}
	updatedItem models.Item
	updateErr   error
}

func (m *mockItemHandlerRepository) FindItemByID(_ context.Context, id int64) (models.Item, error) {
	if m.findErr != nil {
		return models.Item{}, m.findErr
	}
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return models.Item{}, repository.ErrNotFound
}

func (m *mockItemHandlerRepository) FindItemsByIDs(_ context.Context, _ []int64) ([]models.Item, error) {
	return nil, errors.New("not implemented")
}

func (m *mockItemHandlerRepository) UpdateItemAvailability(_ context.Context, id int64, availability models.ItemAvailability) (models.Item, error) {
	m.updateInput.id = id
	m.updateInput.availability = availability
	if m.updateErr != nil {
		return models.Item{}, m.updateErr
	}
	if m.updatedItem.ID != 0 {
		return m.updatedItem, nil
	}
	for i, item := range m.items {
		if item.ID == id {
			item.Availability = availability
			m.items[i] = item
			return item, nil
		}
	}
	return models.Item{}, repository.ErrNotFound
}

func setupItemRouter(itemRepo *mockItemHandlerRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := service.NewItemService(itemRepo)
	handler := NewItemHandler(svc)

	r := gin.New()
	r.GET("/menu/items/:id", handler.GetItem)
	r.PATCH("/menu/items/:id", handler.UpdateItemAvailability)
	return r
}

func TestGetItem_InvalidID(t *testing.T) {
	r := setupItemRouter(&mockItemHandlerRepository{})

	for _, path := range []string{"/menu/items/abc", "/menu/items/0", "/menu/items/-1"} {
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

func TestGetItem_NotFound(t *testing.T) {
	r := setupItemRouter(&mockItemHandlerRepository{})

	req := httptest.NewRequest(http.MethodGet, "/menu/items/99", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] != "item not found" {
		t.Fatalf("error = %q, want item not found", body["error"])
	}
}

func TestGetItem_InternalError(t *testing.T) {
	r := setupItemRouter(&mockItemHandlerRepository{
		findErr: errors.New("db down"),
	})

	req := httptest.NewRequest(http.MethodGet, "/menu/items/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestGetItem_Success(t *testing.T) {
	createdAt := time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC)

	r := setupItemRouter(&mockItemHandlerRepository{
		items: []models.Item{
			{
				ID:           1,
				MerchantID:   1,
				Name:         "Margherita Pizza",
				Price:        12.99,
				Status:       models.ItemStatusActive,
				Availability: models.ItemAvailabilityAvailable,
				CategoryID:   10,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/menu/items/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got dto.ItemDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != 1 || got.Name != "Margherita Pizza" || got.Price != 12.99 {
		t.Fatalf("unexpected item: %+v", got)
	}
	if got.Availability != "AVAILABLE" || got.Status != "ACTIVE" {
		t.Fatalf("unexpected status/availability: %+v", got)
	}
}

func TestUpdateItemAvailability_InvalidID(t *testing.T) {
	r := setupItemRouter(&mockItemHandlerRepository{})

	body := `{"availability":"OUT_OF_STOCK"}`
	req := httptest.NewRequest(http.MethodPatch, "/menu/items/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateItemAvailability_InvalidBody(t *testing.T) {
	r := setupItemRouter(&mockItemHandlerRepository{})

	for _, body := range []string{`{}`, `{"availability":"SOLD_OUT"}`} {
		req := httptest.NewRequest(http.MethodPatch, "/menu/items/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %s: status = %d, want %d", body, rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp["error"] != "availability must be AVAILABLE or OUT_OF_STOCK" {
			t.Fatalf("error = %q, want availability validation message", resp["error"])
		}
	}
}

func TestUpdateItemAvailability_NotFound(t *testing.T) {
	r := setupItemRouter(&mockItemHandlerRepository{})

	body := `{"availability":"OUT_OF_STOCK"}`
	req := httptest.NewRequest(http.MethodPatch, "/menu/items/99", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdateItemAvailability_InternalError(t *testing.T) {
	r := setupItemRouter(&mockItemHandlerRepository{
		items: []models.Item{
			{ID: 1, Availability: models.ItemAvailabilityAvailable},
		},
		updateErr: errors.New("db down"),
	})

	body := `{"availability":"OUT_OF_STOCK"}`
	req := httptest.NewRequest(http.MethodPatch, "/menu/items/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestUpdateItemAvailability_Success(t *testing.T) {
	updatedAt := time.Date(2026, 1, 10, 11, 0, 0, 0, time.UTC)

	r := setupItemRouter(&mockItemHandlerRepository{
		items: []models.Item{
			{
				ID:           1,
				MerchantID:   1,
				Name:         "Margherita Pizza",
				Price:        12.99,
				Status:       models.ItemStatusActive,
				Availability: models.ItemAvailabilityAvailable,
				CategoryID:   10,
			},
		},
		updatedItem: models.Item{
			ID:           1,
			MerchantID:   1,
			Name:         "Margherita Pizza",
			Price:        12.99,
			Status:       models.ItemStatusActive,
			Availability: models.ItemAvailabilityOutOfStock,
			CategoryID:   10,
			UpdatedAt:    updatedAt,
		},
	})

	body := `{"availability":"OUT_OF_STOCK"}`
	req := httptest.NewRequest(http.MethodPatch, "/menu/items/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got dto.ItemDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != 1 || got.Availability != "OUT_OF_STOCK" {
		t.Fatalf("unexpected item: %+v", got)
	}
}
