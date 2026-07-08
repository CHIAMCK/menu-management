package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"menu-management/internal/dto"
	"menu-management/internal/lock/locktest"
	"menu-management/internal/messaging"
	"menu-management/internal/models"
	"menu-management/internal/repository"
)

type noopUserLocker struct{}

func (noopUserLocker) TryLock(context.Context, int64) error { return nil }

type mockOrderRepository struct {
	order       models.Order
	orderErr    error
	items       []repository.OrderItemWithItem
	itemsErr    error
	orderID     int64
	itemsForID  int64
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
	m.orderID = id
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
	m.itemsForID = orderID
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
				Name:      "Item",
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

func (m *mockItemRepository) FindItemsForOrder(_ context.Context, ids []int64) ([]models.Item, error) {
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

func TestGetOrderByID_InvalidID(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{}, nil, noopUserLocker{})

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
	}, &mockItemRepository{}, nil, noopUserLocker{})

	_, err := svc.GetOrderByID(context.Background(), 99)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("GetOrderByID: want ErrOrderNotFound, got %v", err)
	}
}

func TestGetOrderByID_FindOrderError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewOrderService(&mockOrderRepository{
		orderErr: repoErr,
	}, &mockItemRepository{}, nil, noopUserLocker{})

	_, err := svc.GetOrderByID(context.Background(), 1)
	if err == nil || errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("GetOrderByID: want wrapped repo error, got %v", err)
	}
}

func TestGetOrderByID_FindItemsError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewOrderService(&mockOrderRepository{
		order:    models.Order{ID: 1},
		itemsErr: repoErr,
	}, &mockItemRepository{}, nil, noopUserLocker{})

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
	svc := NewOrderService(mockRepo, &mockItemRepository{}, nil, noopUserLocker{})

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
		order: models.Order{ID: 2, Status: models.OrderStatusReceived},
		items: []repository.OrderItemWithItem{},
	}, &mockItemRepository{}, nil, noopUserLocker{})

	got, err := svc.GetOrderByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetOrderByID: unexpected error: %v", err)
	}
	if got.Items == nil || len(got.Items) != 0 {
		t.Fatalf("Items = %v, want empty slice", got.Items)
	}
}

type mockOrderEventPublisher struct {
	event   messaging.OrderPlacedEvent
	publishErr error
}

func (m *mockOrderEventPublisher) PublishOrderPlaced(_ context.Context, event messaging.OrderPlacedEvent) error {
	m.event = event
	return m.publishErr
}

func (m *mockOrderEventPublisher) Close() error {
	return nil
}

func TestCreateOrder_PublishesOrderPlacedEvent(t *testing.T) {
	orderRepo := &mockOrderRepository{createID: 99}
	itemRepo := &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Name: "Margherita Pizza", Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	}
	publisher := &mockOrderEventPublisher{}
	svc := NewOrderService(orderRepo, itemRepo, publisher, noopUserLocker{})

	got, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items:      []dto.CreateOrderItemRequest{{ItemID: 1, Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("CreateOrder: unexpected error: %v", err)
	}

	if publisher.event.OrderID != got.ID {
		t.Fatalf("published OrderID = %d, want %d", publisher.event.OrderID, got.ID)
	}
	if publisher.event.UserID != 1 || publisher.event.MerchantID != 1 {
		t.Fatalf("unexpected published order fields: %+v", publisher.event)
	}
	if publisher.event.Status != "RECEIVED" {
		t.Fatalf("published Status = %q, want RECEIVED", publisher.event.Status)
	}
	if len(publisher.event.Items) != 1 || publisher.event.Items[0].Quantity != 2 {
		t.Fatalf("unexpected published items: %+v", publisher.event.Items)
	}
}

func TestCreateOrder_Success(t *testing.T) {
	orderRepo := &mockOrderRepository{createID: 99}
	itemRepo := &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Name: "Margherita Pizza", Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
			{ID: 4, MerchantID: 1, Name: "Garlic Bread", Price: 5.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	}
	svc := NewOrderService(orderRepo, itemRepo, nil, noopUserLocker{})

	got, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items: []dto.CreateOrderItemRequest{
			{ItemID: 1, Quantity: 2},
			{ItemID: 4, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: unexpected error: %v", err)
	}

	wantTotal := 12.99*2 + 5.99
	if orderRepo.createInput.TotalAmount != wantTotal {
		t.Fatalf("TotalAmount = %v, want %v", orderRepo.createInput.TotalAmount, wantTotal)
	}
	if orderRepo.orderID != 0 {
		t.Fatalf("FindOrderByID should not be called after create, got orderID %d", orderRepo.orderID)
	}
	if orderRepo.itemsForID != 0 {
		t.Fatalf("FindOrderItemsByOrderID should not be called after create, got itemsForID %d", orderRepo.itemsForID)
	}
	if got.ID != 99 || got.Status != "RECEIVED" || got.TotalAmount != wantTotal {
		t.Fatalf("unexpected order: %+v", got)
	}
	if len(got.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(got.Items))
	}
	if got.Items[0].Name == "" {
		t.Fatalf("expected item name from loaded data, got empty name")
	}
}

func TestCreateOrder_ItemNotFound(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Name: "Margherita Pizza", Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	}, nil, noopUserLocker{})

	_, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items: []dto.CreateOrderItemRequest{
			{ItemID: 1, Quantity: 1},
			{ItemID: 99, Quantity: 1},
		},
	})
	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("CreateOrder: want ErrItemNotFound, got %v", err)
	}
}

func TestCreateOrder_ItemUnavailable(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{
		items: []models.Item{
			{ID: 5, MerchantID: 1, Name: "Sparkling Water", Price: 2.50, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityOutOfStock},
		},
	}, nil, noopUserLocker{})

	_, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items:      []dto.CreateOrderItemRequest{{ItemID: 5, Quantity: 1}},
	})
	if !errors.Is(err, ErrItemUnavailable) {
		t.Fatalf("CreateOrder: want ErrItemUnavailable, got %v", err)
	}
}

func TestCreateOrder_DuplicateItem(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{}, nil, noopUserLocker{})

	_, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items: []dto.CreateOrderItemRequest{
			{ItemID: 1, Quantity: 1},
			{ItemID: 1, Quantity: 2},
		},
	})
	if !errors.Is(err, ErrDuplicateOrderItem) {
		t.Fatalf("CreateOrder: want ErrDuplicateOrderItem, got %v", err)
	}
}

func TestCreateOrder_InvalidRequest(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{}, nil, noopUserLocker{})

	tests := []struct {
		name string
		req  dto.CreateOrderRequest
	}{
		{
			name: "missing user id",
			req: dto.CreateOrderRequest{
				MerchantID: 1,
				Items:      []dto.CreateOrderItemRequest{{ItemID: 1, Quantity: 1}},
			},
		},
		{
			name: "missing merchant id",
			req: dto.CreateOrderRequest{
				UserID: 1,
				Items:  []dto.CreateOrderItemRequest{{ItemID: 1, Quantity: 1}},
			},
		},
		{
			name: "empty items",
			req: dto.CreateOrderRequest{
				UserID:     1,
				MerchantID: 1,
				Items:      []dto.CreateOrderItemRequest{},
			},
		},
		{
			name: "zero quantity",
			req: dto.CreateOrderRequest{
				UserID:     1,
				MerchantID: 1,
				Items:      []dto.CreateOrderItemRequest{{ItemID: 1, Quantity: 0}},
			},
		},
		{
			name: "missing item id",
			req: dto.CreateOrderRequest{
				UserID:     1,
				MerchantID: 1,
				Items:      []dto.CreateOrderItemRequest{{Quantity: 1}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateOrder(context.Background(), tt.req)
			if !errors.Is(err, ErrInvalidOrderRequest) {
				t.Fatalf("CreateOrder: want ErrInvalidOrderRequest, got %v", err)
			}
		})
	}
}

func TestCreateOrder_EmptyItemName(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Name: "   ", Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	}, nil, noopUserLocker{})

	_, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items:      []dto.CreateOrderItemRequest{{ItemID: 1, Quantity: 1}},
	})
	if !errors.Is(err, ErrInvalidOrderRequest) {
		t.Fatalf("CreateOrder: want ErrInvalidOrderRequest, got %v", err)
	}
}

func TestCreateOrder_UserLocked(t *testing.T) {
	userLocker := locktest.NewUserLocker(t)
	if err := userLocker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("TryLock: unexpected error: %v", err)
	}

	orderRepo := &mockOrderRepository{createID: 99}
	svc := NewOrderService(orderRepo, &mockItemRepository{
		items: []models.Item{
			{ID: 1, MerchantID: 1, Name: "Margherita Pizza", Price: 12.99, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	}, nil, userLocker)

	_, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items:      []dto.CreateOrderItemRequest{{ItemID: 1, Quantity: 1}},
	})
	if !errors.Is(err, ErrUserOrderLocked) {
		t.Fatalf("CreateOrder: want ErrUserOrderLocked, got %v", err)
	}
	if orderRepo.createInput.UserID != 0 {
		t.Fatalf("CreateOrder should not reach repository when user is locked")
	}
}

func TestCreateOrder_ItemMerchantMismatch(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{
		items: []models.Item{
			{ID: 7, MerchantID: 2, Name: "Caesar Salad", Price: 4.50, Status: models.ItemStatusActive, Availability: models.ItemAvailabilityAvailable},
		},
	}, nil, noopUserLocker{})

	_, err := svc.CreateOrder(context.Background(), dto.CreateOrderRequest{
		UserID:     1,
		MerchantID: 1,
		Items:      []dto.CreateOrderItemRequest{{ItemID: 7, Quantity: 1}},
	})
	if !errors.Is(err, ErrItemMerchantMismatch) {
		t.Fatalf("CreateOrder: want ErrItemMerchantMismatch, got %v", err)
	}
}

func TestUpdateOrderStatus_InvalidID(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{}, nil, noopUserLocker{})

	for _, id := range []int64{0, -1} {
		_, err := svc.UpdateOrderStatus(context.Background(), id, "PREPARING")
		if !errors.Is(err, ErrInvalidOrderID) {
			t.Fatalf("UpdateOrderStatus(%d): want ErrInvalidOrderID, got %v", id, err)
		}
	}
}

func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{}, &mockItemRepository{}, nil, noopUserLocker{})

	_, err := svc.UpdateOrderStatus(context.Background(), 1, "SHIPPED")
	if !errors.Is(err, ErrInvalidOrderStatus) {
		t.Fatalf("UpdateOrderStatus: want ErrInvalidOrderStatus, got %v", err)
	}
}

func TestUpdateOrderStatus_NotFound(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{
		orderErr: repository.ErrNotFound,
	}, &mockItemRepository{}, nil, noopUserLocker{})

	_, err := svc.UpdateOrderStatus(context.Background(), 99, "PREPARING")
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("UpdateOrderStatus: want ErrOrderNotFound, got %v", err)
	}
}

func TestUpdateOrderStatus_InvalidTransition(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{
		order: models.Order{ID: 1, Status: models.OrderStatusReceived},
	}, &mockItemRepository{}, nil, noopUserLocker{})

	_, err := svc.UpdateOrderStatus(context.Background(), 1, "READY")
	if !errors.Is(err, ErrInvalidOrderStatusTransition) {
		t.Fatalf("UpdateOrderStatus: want ErrInvalidOrderStatusTransition, got %v", err)
	}
}

func TestUpdateOrderStatus_CompletedOrder(t *testing.T) {
	svc := NewOrderService(&mockOrderRepository{
		order: models.Order{ID: 1, Status: models.OrderStatusCompleted},
	}, &mockItemRepository{}, nil, noopUserLocker{})

	_, err := svc.UpdateOrderStatus(context.Background(), 1, "PREPARING")
	if !errors.Is(err, ErrInvalidOrderStatusTransition) {
		t.Fatalf("UpdateOrderStatus: want ErrInvalidOrderStatusTransition, got %v", err)
	}
}

func TestUpdateOrderStatus_Success(t *testing.T) {
	updatedAt := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	orderRepo := &mockOrderRepository{
		order: models.Order{
			ID:          1,
			UserID:      1,
			MerchantID:  1,
			Status:      models.OrderStatusReceived,
			TotalAmount: 14.99,
		},
		updatedOrder: models.Order{
			ID:          1,
			UserID:      1,
			MerchantID:  1,
			Status:      models.OrderStatusPreparing,
			TotalAmount: 14.99,
			UpdatedAt:   updatedAt,
		},
		items: []repository.OrderItemWithItem{
			{ID: 1, ItemID: 2, Name: "Pepperoni Pizza", Quantity: 1, UnitPrice: 14.99},
		},
	}
	svc := NewOrderService(orderRepo, &mockItemRepository{}, nil, noopUserLocker{})

	got, err := svc.UpdateOrderStatus(context.Background(), 1, "PREPARING")
	if err != nil {
		t.Fatalf("UpdateOrderStatus: unexpected error: %v", err)
	}

	if orderRepo.updateInput.id != 1 || orderRepo.updateInput.status != models.OrderStatusPreparing {
		t.Fatalf("UpdateOrderStatus repo call = %+v, want id=1 status=PREPARING", orderRepo.updateInput)
	}
	if got.Status != "PREPARING" {
		t.Fatalf("Status = %q, want PREPARING", got.Status)
	}
	if len(got.Items) != 1 || got.Items[0].Name != "Pepperoni Pizza" {
		t.Fatalf("unexpected items: %+v", got.Items)
	}
}
