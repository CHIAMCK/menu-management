package models

import "time"

type ItemStatus string

const (
	ItemStatusActive   ItemStatus = "ACTIVE"
	ItemStatusInactive ItemStatus = "INACTIVE"
	ItemStatusArchived ItemStatus = "ARCHIVED"
)

type ItemAvailability string

const (
	ItemAvailabilityAvailable  ItemAvailability = "AVAILABLE"
	ItemAvailabilityOutOfStock ItemAvailability = "OUT_OF_STOCK"
)

type Item struct {
	ID           int64            `json:"id" db:"id"`
	MerchantID   int64            `json:"merchant_id" db:"merchant_id"`
	Name         string           `json:"name" db:"name"`
	Price        float64          `json:"price" db:"price"`
	Status       ItemStatus       `json:"status" db:"status"`
	Availability ItemAvailability `json:"availability" db:"availability"`
	CategoryID   int64            `json:"category_id" db:"category_id"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}
