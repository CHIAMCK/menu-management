package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"menu-management/internal/models"
)

type OrderItemWithItem struct {
	ID        int64
	ItemID    int64
	Name      string
	Quantity  int
	UnitPrice float64
}

type CreateOrderInput struct {
	UserID      int64
	MerchantID  int64
	TotalAmount float64
	Items       []CreateOrderItemInput
}

type CreateOrderItemInput struct {
	ItemID    int64
	Quantity  int
	UnitPrice float64
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) FindOrderByID(ctx context.Context, id int64) (models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		if mapped := mapRecordNotFound(err); mapped != err {
			return models.Order{}, mapped
		}
		return models.Order{}, fmt.Errorf("find order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) FindOrderItemsByOrderID(ctx context.Context, orderID int64) ([]OrderItemWithItem, error) {
	var items []OrderItemWithItem
	err := r.db.WithContext(ctx).
		Table("order_items AS oi").
		Select("oi.id, oi.item_id, oi.quantity, oi.unit_price, i.name").
		Joins("JOIN items i ON oi.item_id = i.id").
		Where("oi.order_id = ?", orderID).
		Order("oi.id").
		Scan(&items).Error
	if err != nil {
		return nil, fmt.Errorf("find order items: %w", err)
	}

	return items, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, input CreateOrderInput) (int64, error) {
	var orderID int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order := models.Order{
			UserID:      input.UserID,
			MerchantID:  input.MerchantID,
			Status:      models.OrderStatusReceived,
			TotalAmount: input.TotalAmount,
		}
		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("insert order: %w", err)
		}

		if len(input.Items) > 0 {
			orderItems := make([]models.OrderItem, len(input.Items))
			for i, item := range input.Items {
				orderItems[i] = models.OrderItem{
					OrderID:   order.ID,
					ItemID:    item.ItemID,
					Quantity:  item.Quantity,
					UnitPrice: item.UnitPrice,
				}
			}

			if err := tx.Create(&orderItems).Error; err != nil {
				return fmt.Errorf("insert order items: %w", err)
			}
		}

		orderID = order.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id int64, status models.OrderStatus) (models.Order, error) {
	var order models.Order
	result := r.db.WithContext(ctx).
		Model(&order).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"updated_at": gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return models.Order{}, fmt.Errorf("update order status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.Order{}, ErrNotFound
	}

	return order, nil
}
