package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"menu-management/internal/models"
	"menu-management/internal/repository"
)

type mockItemServiceRepository struct {
	items       []models.Item
	findErr     error
	updateInput struct {
		id           int64
		availability models.ItemAvailability
	}
	updatedItem models.Item
	updateErr   error
}

func (m *mockItemServiceRepository) FindItemByID(_ context.Context, id int64) (models.Item, error) {
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

func (m *mockItemServiceRepository) FindItemsByIDs(_ context.Context, _ []int64) ([]models.Item, error) {
	return nil, errors.New("not implemented")
}

func (m *mockItemServiceRepository) UpdateItemAvailability(_ context.Context, id int64, availability models.ItemAvailability) (models.Item, error) {
	m.updateInput.id = id
	m.updateInput.availability = availability
	if m.updateErr != nil {
		return models.Item{}, m.updateErr
	}
	if m.updatedItem.ID != 0 {
		return m.updatedItem, nil
	}
	return models.Item{}, repository.ErrNotFound
}

func TestGetItemByID_InvalidID(t *testing.T) {
	svc := NewItemService(&mockItemServiceRepository{})

	for _, id := range []int64{0, -1} {
		_, err := svc.GetItemByID(context.Background(), id)
		if !errors.Is(err, ErrInvalidItemID) {
			t.Fatalf("GetItemByID(%d): want ErrInvalidItemID, got %v", id, err)
		}
	}
}

func TestGetItemByID_NotFound(t *testing.T) {
	svc := NewItemService(&mockItemServiceRepository{})

	_, err := svc.GetItemByID(context.Background(), 99)
	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("GetItemByID: want ErrItemNotFound, got %v", err)
	}
}

func TestGetItemByID_FindError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewItemService(&mockItemServiceRepository{
		findErr: repoErr,
	})

	_, err := svc.GetItemByID(context.Background(), 1)
	if err == nil || errors.Is(err, ErrItemNotFound) {
		t.Fatalf("GetItemByID: want wrapped repo error, got %v", err)
	}
}

func TestGetItemByID_Success(t *testing.T) {
	createdAt := time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC)

	svc := NewItemService(&mockItemServiceRepository{
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

	got, err := svc.GetItemByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetItemByID: unexpected error: %v", err)
	}

	if got.ID != 1 || got.Name != "Margherita Pizza" || got.Price != 12.99 {
		t.Fatalf("unexpected item: %+v", got)
	}
	if got.Availability != "AVAILABLE" || got.CategoryID != 10 {
		t.Fatalf("unexpected fields: %+v", got)
	}
}

func TestUpdateItemAvailability_InvalidID(t *testing.T) {
	svc := NewItemService(&mockItemServiceRepository{})

	for _, id := range []int64{0, -1} {
		_, err := svc.UpdateItemAvailability(context.Background(), id, "AVAILABLE")
		if !errors.Is(err, ErrInvalidItemID) {
			t.Fatalf("UpdateItemAvailability(%d): want ErrInvalidItemID, got %v", id, err)
		}
	}
}

func TestUpdateItemAvailability_InvalidAvailability(t *testing.T) {
	svc := NewItemService(&mockItemServiceRepository{})

	for _, availability := range []string{"", "SOLD_OUT", "available"} {
		_, err := svc.UpdateItemAvailability(context.Background(), 1, availability)
		if !errors.Is(err, ErrInvalidItemAvailability) {
			t.Fatalf("UpdateItemAvailability(%q): want ErrInvalidItemAvailability, got %v", availability, err)
		}
	}
}

func TestUpdateItemAvailability_NotFound(t *testing.T) {
	svc := NewItemService(&mockItemServiceRepository{})

	_, err := svc.UpdateItemAvailability(context.Background(), 99, "OUT_OF_STOCK")
	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("UpdateItemAvailability: want ErrItemNotFound, got %v", err)
	}
}

func TestUpdateItemAvailability_UpdateError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewItemService(&mockItemServiceRepository{
		updateErr: repoErr,
	})

	_, err := svc.UpdateItemAvailability(context.Background(), 1, "OUT_OF_STOCK")
	if err == nil || errors.Is(err, ErrItemNotFound) {
		t.Fatalf("UpdateItemAvailability: want wrapped repo error, got %v", err)
	}
}

func TestUpdateItemAvailability_Success(t *testing.T) {
	updatedAt := time.Date(2026, 1, 10, 11, 0, 0, 0, time.UTC)
	itemRepo := &mockItemServiceRepository{
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
	}
	svc := NewItemService(itemRepo)

	got, err := svc.UpdateItemAvailability(context.Background(), 1, "OUT_OF_STOCK")
	if err != nil {
		t.Fatalf("UpdateItemAvailability: unexpected error: %v", err)
	}

	if itemRepo.updateInput.id != 1 || itemRepo.updateInput.availability != models.ItemAvailabilityOutOfStock {
		t.Fatalf("UpdateItemAvailability repo call = %+v, want id=1 availability=OUT_OF_STOCK", itemRepo.updateInput)
	}
	if got.ID != 1 || got.Availability != "OUT_OF_STOCK" {
		t.Fatalf("unexpected item: %+v", got)
	}
}
