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
	ID           int64            `json:"id" db:"id" gorm:"primaryKey"`
	MerchantID   int64            `json:"merchant_id" db:"merchant_id"`
	Name         string           `json:"name" db:"name"`
	Price        float64          `json:"price" db:"price" gorm:"type:numeric(12,2)"`
	Status       ItemStatus       `json:"status" db:"status" gorm:"type:item_status"`
	Availability ItemAvailability `json:"availability" db:"availability" gorm:"type:item_availability"`
	CategoryID   int64            `json:"category_id" db:"category_id"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}

func (Item) TableName() string {
	return "items"
}
