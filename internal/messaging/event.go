package messaging

import (
	"time"

	"menu-management/internal/dto"
)

type OrderPlacedItem struct {
	ItemID    int64   `json:"item_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type OrderPlacedEvent struct {
	OrderID     int64             `json:"order_id"`
	UserID      int64             `json:"user_id"`
	MerchantID  int64             `json:"merchant_id"`
	Status      string            `json:"status"`
	TotalAmount float64           `json:"total_amount"`
	CreatedAt   time.Time         `json:"created_at"`
	Items       []OrderPlacedItem `json:"items"`
}

func FromOrderDetail(order dto.OrderDetailResponse) OrderPlacedEvent {
	items := make([]OrderPlacedItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, OrderPlacedItem{
			ItemID:    item.ID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return OrderPlacedEvent{
		OrderID:     order.ID,
		UserID:      order.UserID,
		MerchantID:  order.MerchantID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt,
		Items:       items,
	}
}
