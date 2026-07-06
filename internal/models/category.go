package models

import "time"

type CategoryStatus string

const (
	CategoryStatusActive   CategoryStatus = "ACTIVE"
	CategoryStatusInactive CategoryStatus = "INACTIVE"
)

type Category struct {
	ID        int64          `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	MenuID    int64          `json:"menu_id" db:"menu_id"`
	Status    CategoryStatus `json:"status" db:"status"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}
