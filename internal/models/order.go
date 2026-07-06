package models

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          int64       `json:"id" db:"id"`
	UserID      int64       `json:"user_id" db:"user_id"`
	MerchantID  int64       `json:"merchant_id" db:"merchant_id"`
	Status      OrderStatus `json:"status" db:"status"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}
