package models

import "time"

type MerchantStatus string

const (
	MerchantStatusActive   MerchantStatus = "ACTIVE"
	MerchantStatusInactive MerchantStatus = "INACTIVE"
)

type Merchant struct {
	ID        int64          `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	Status    MerchantStatus `json:"status" db:"status"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}
