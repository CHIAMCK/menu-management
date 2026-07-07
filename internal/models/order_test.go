package models

import "testing"

func TestOrderStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from   OrderStatus
		to     OrderStatus
		allowed bool
	}{
		{OrderStatusReceived, OrderStatusPreparing, true},
		{OrderStatusPreparing, OrderStatusReady, true},
		{OrderStatusReady, OrderStatusCompleted, true},
		{OrderStatusReceived, OrderStatusReady, false},
		{OrderStatusReceived, OrderStatusCompleted, false},
		{OrderStatusPreparing, OrderStatusCompleted, false},
		{OrderStatusPreparing, OrderStatusReceived, false},
		{OrderStatusReady, OrderStatusPreparing, false},
		{OrderStatusCompleted, OrderStatusPreparing, false},
		{OrderStatusReceived, OrderStatusReceived, false},
		{OrderStatusCompleted, OrderStatusCompleted, false},
	}

	for _, tt := range tests {
		got := tt.from.CanTransitionTo(tt.to)
		if got != tt.allowed {
			t.Fatalf("%s -> %s: got %v, want %v", tt.from, tt.to, got, tt.allowed)
		}
	}
}
