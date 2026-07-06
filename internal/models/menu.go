package models

import "time"

type MenuStatus string

const (
	MenuStatusActive   MenuStatus = "ACTIVE"
	MenuStatusInactive MenuStatus = "INACTIVE"
)

type Menu struct {
	ID         int64      `json:"id" db:"id"`
	MerchantID int64      `json:"merchant_id" db:"merchant_id"`
	Name       string     `json:"name" db:"name"`
	Status     MenuStatus `json:"status" db:"status"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}
