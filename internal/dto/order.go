package dto

import "time"

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

type OrderItemResponse struct {
	ID        int64   `json:"id"`
	ItemID    int64   `json:"item_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}
