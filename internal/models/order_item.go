package models

import "time"

type OrderItem struct {
	ID        int64     `json:"id" db:"id" gorm:"primaryKey"`
	OrderID   int64     `json:"order_id" db:"order_id"`
	ItemID    int64     `json:"item_id" db:"item_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	UnitPrice float64   `json:"unit_price" db:"unit_price" gorm:"type:numeric(12,2)"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
