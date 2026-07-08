package dto

import "time"

type CreateOrderRequest struct {
	UserID     int64                    `json:"user_id" binding:"required,gt=0"`
	MerchantID int64                    `json:"merchant_id" binding:"required,gt=0"`
	Items      []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type CreateOrderItemRequest struct {
	ItemID   int64 `json:"item_id" binding:"required,gt=0"`
	Quantity int   `json:"quantity" binding:"required,gt=0"`
}

type OrderDetailResponse struct {
	ID          int64               `json:"id"`
	UserID      int64               `json:"user_id"`
	MerchantID  int64               `json:"merchant_id"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Items       []OrderItemResponse `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=RECEIVED PREPARING READY COMPLETED"`
}

type OrderItemResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}
