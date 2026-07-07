package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"menu-management/internal/models"
	"menu-management/internal/repository"
)

type mockOrderRepository struct {
	order      models.Order
	orderErr   error
	items      []repository.OrderItemWithItem
	itemsErr   error
	orderID    int64
	itemsForID int64
}

func (m *mockOrderRepository) FindOrderByID(_ context.Context, id int64) (models.Order, error) {
	m.orderID = id
	if m.orderErr != nil {
		return models.Order{}, m.orderErr
	}
	return m.order, nil
}

func (m *mockOrderRepository) FindOrderItemsByOrderID(_ context.Context, orderID int64) ([]repository.OrderItemWithItem, error) {
	m.itemsForID = orderID
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	return m.items, nil
}

func TestGetOrderByID_InvalidID(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{})

	for _, id := range []int64{0, -1} {
		_, err := svc.GetOrderByID(context.Background(), id)
		if !errors.Is(err, ErrInvalidOrderID) {
			t.Fatalf("GetOrderByID(%d): want ErrInvalidOrderID, got %v", id, err)
		}
	}
}

func TestGetOrderByID_NotFound(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{
		orderErr: repository.ErrNotFound,
	})

	_, err := svc.GetOrderByID(context.Background(), 99)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("GetOrderByID: want ErrOrderNotFound, got %v", err)
	}
}

func TestGetOrderByID_FindOrderError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewOrderService(&mockOrderRepository{
		orderErr: repoErr,
	})

	_, err := svc.GetOrderByID(context.Background(), 1)
	if err == nil || errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("GetOrderByID: want wrapped repo error, got %v", err)
	}
}

func TestGetOrderByID_FindItemsError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewOrderService(&mockOrderRepository{
		order: models.Order{ID: 1},
		itemsErr: repoErr,
	})

	_, err := svc.GetOrderByID(context.Background(), 1)
	if err == nil {
		t.Fatal("GetOrderByID: expected error, got nil")
	}
}

func TestGetOrderByID_Success(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC)

	mockRepo := &mockOrderRepository{
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
			{ID: 2, ItemID: 4, Name: "Garlic Bread", Quantity: 1, UnitPrice: 5.99},
		},
	}
	svc := NewOrderService(mockRepo)

	got, err := svc.GetOrderByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetOrderByID: unexpected error: %v", err)
	}

	if mockRepo.orderID != 1 {
		t.Fatalf("FindOrderByID called with id %d, want 1", mockRepo.orderID)
	}
	if mockRepo.itemsForID != 1 {
		t.Fatalf("FindOrderItemsByOrderID called with orderID %d, want 1", mockRepo.itemsForID)
	}

	if got.ID != 1 || got.UserID != 1 || got.MerchantID != 1 {
		t.Fatalf("unexpected order fields: %+v", got)
	}
	if got.Status != "COMPLETED" {
		t.Fatalf("Status = %q, want COMPLETED", got.Status)
	}
	if got.TotalAmount != 31.97 {
		t.Fatalf("TotalAmount = %v, want 31.97", got.TotalAmount)
	}
	if len(got.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(got.Items))
	}
	if got.Items[0].Name != "Margherita Pizza" || got.Items[0].Quantity != 2 {
		t.Fatalf("unexpected first item: %+v", got.Items[0])
	}
}

func TestGetOrderByID_SuccessEmptyItems(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{
		order: models.Order{ID: 2, Status: models.OrderStatusPending},
		items: []repository.OrderItemWithItem{},
	})

	got, err := svc.GetOrderByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetOrderByID: unexpected error: %v", err)
	}
	if got.Items == nil || len(got.Items) != 0 {
		t.Fatalf("Items = %v, want empty slice", got.Items)
	}
}
