package models

import "time"

type OrderStatus string

const (
	OrderStatusReceived  OrderStatus = "RECEIVED"
	OrderStatusPreparing OrderStatus = "PREPARING"
	OrderStatusReady     OrderStatus = "READY"
	OrderStatusCompleted OrderStatus = "COMPLETED"
)

var orderStatusTransitions = map[OrderStatus]OrderStatus{
	OrderStatusReceived:  OrderStatusPreparing,
	OrderStatusPreparing: OrderStatusReady,
	OrderStatusReady:     OrderStatusCompleted,
}

func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	next, ok := orderStatusTransitions[s]
	return ok && next == target
}

type Order struct {
	ID          int64       `json:"id" db:"id" gorm:"primaryKey"`
	UserID      int64       `json:"user_id" db:"user_id"`
	MerchantID  int64       `json:"merchant_id" db:"merchant_id"`
	Status      OrderStatus `json:"status" db:"status" gorm:"type:order_status;default:RECEIVED"`
	TotalAmount float64     `json:"total_amount" db:"total_amount" gorm:"type:numeric(12,2)"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}
