package messaging

import (
	"encoding/json"
	"testing"
	"time"

	"menu-management/internal/dto"
)

func TestFromOrderDetail(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	order := dto.OrderDetailResponse{
		ID:          42,
		UserID:      1,
		MerchantID:  2,
		Status:      "RECEIVED",
		TotalAmount: 31.97,
		CreatedAt:   createdAt,
		Items: []dto.OrderItemResponse{
			{ID: 1, Name: "Margherita Pizza", Quantity: 2, UnitPrice: 12.99},
			{ID: 4, Name: "Garlic Bread", Quantity: 1, UnitPrice: 5.99},
		},
	}

	event := FromOrderDetail(order)

	if event.OrderID != 42 || event.UserID != 1 || event.MerchantID != 2 {
		t.Fatalf("unexpected event header: %+v", event)
	}
	if event.Status != "RECEIVED" || event.TotalAmount != 31.97 {
		t.Fatalf("unexpected event status/total: %+v", event)
	}
	if !event.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %v, want %v", event.CreatedAt, createdAt)
	}
	if len(event.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(event.Items))
	}
	if event.Items[0].Name != "Margherita Pizza" || event.Items[1].UnitPrice != 5.99 {
		t.Fatalf("unexpected items: %+v", event.Items)
	}
}

func TestOrderPlacedEvent_JSONRoundTrip(t *testing.T) {
	event := OrderPlacedEvent{
		OrderID:     7,
		UserID:      3,
		MerchantID:  1,
		Status:      "RECEIVED",
		TotalAmount: 14.99,
		CreatedAt:   time.Date(2026, 2, 1, 8, 30, 0, 0, time.UTC),
		Items: []OrderPlacedItem{
			{ItemID: 2, Name: "Pepperoni Pizza", Quantity: 1, UnitPrice: 14.99},
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded OrderPlacedEvent
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.OrderID != event.OrderID || decoded.UserID != event.UserID {
		t.Fatalf("decoded header mismatch: %+v", decoded)
	}
	if len(decoded.Items) != 1 || decoded.Items[0].Name != "Pepperoni Pizza" {
		t.Fatalf("decoded items mismatch: %+v", decoded.Items)
	}
}
