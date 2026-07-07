package dto

import "time"

type UpdateItemAvailabilityRequest struct {
	Availability string `json:"availability" binding:"required"`
}

type ItemDetailResponse struct {
	ID           int64     `json:"id"`
	MerchantID   int64     `json:"merchant_id"`
	Name         string    `json:"name"`
	Price        float64   `json:"price"`
	Status       string    `json:"status"`
	Availability string    `json:"availability"`
	CategoryID   int64     `json:"category_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
