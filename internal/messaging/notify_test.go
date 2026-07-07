package messaging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

func TestLogKitchenDisplayNotification(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	event := OrderPlacedEvent{
		OrderID:     10,
		UserID:      2,
		MerchantID:  1,
		Status:      "RECEIVED",
		TotalAmount: 25.98,
		CreatedAt:   time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		Items: []OrderPlacedItem{
			{ItemID: 1, Name: "Margherita Pizza", Quantity: 2, UnitPrice: 12.99},
		},
	}

	LogKitchenDisplayNotification(logger, event)

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Unmarshal log output: %v", err)
	}

	if logEntry["msg"] != "kitchen display notification" {
		t.Fatalf("msg = %v, want kitchen display notification", logEntry["msg"])
	}
	if logEntry["event"] != "order.placed" {
		t.Fatalf("event = %v, want order.placed", logEntry["event"])
	}
	if logEntry["order_id"] != float64(10) {
		t.Fatalf("order_id = %v, want 10", logEntry["order_id"])
	}
	if logEntry["merchant_id"] != float64(1) {
		t.Fatalf("merchant_id = %v, want 1", logEntry["merchant_id"])
	}
	if logEntry["user_id"] != float64(2) {
		t.Fatalf("user_id = %v, want 2", logEntry["user_id"])
	}
	if logEntry["status"] != "RECEIVED" {
		t.Fatalf("status = %v, want RECEIVED", logEntry["status"])
	}
	if logEntry["total_amount"] != 25.98 {
		t.Fatalf("total_amount = %v, want 25.98", logEntry["total_amount"])
	}
}
