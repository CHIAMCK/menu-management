package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

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
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) FindOrderByID(ctx context.Context, id int64) (models.Order, error) {
	const query = `
		SELECT id, user_id, merchant_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1`

	var order models.Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.MerchantID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Order{}, ErrNotFound
	}
	if err != nil {
		return models.Order{}, fmt.Errorf("find order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) FindOrderItemsByOrderID(ctx context.Context, orderID int64) ([]OrderItemWithItem, error) {
	const query = `
		SELECT oi.id, oi.item_id, oi.quantity, oi.unit_price, i.name
		FROM order_items oi
		JOIN items i ON oi.item_id = i.id
		WHERE oi.order_id = $1
		ORDER BY oi.id`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("find order items: %w", err)
	}
	defer rows.Close()

	items := make([]OrderItemWithItem, 0)
	for rows.Next() {
		var item OrderItemWithItem
		if err := rows.Scan(
			&item.ID,
			&item.ItemID,
			&item.Quantity,
			&item.UnitPrice,
			&item.Name,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}

	return items, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, input CreateOrderInput) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin create order tx: %w", err)
	}
	defer tx.Rollback()

	const insertOrderQuery = `
		INSERT INTO orders (user_id, merchant_id, status, total_amount)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var orderID int64
	err = tx.QueryRowContext(
		ctx,
		insertOrderQuery,
		input.UserID,
		input.MerchantID,
		models.OrderStatusReceived,
		input.TotalAmount,
	).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("insert order: %w", err)
	}

	if err := insertOrderItems(ctx, tx, orderID, input.Items); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit create order tx: %w", err)
	}

	return orderID, nil
}

func insertOrderItems(ctx context.Context, tx *sql.Tx, orderID int64, items []CreateOrderItemInput) error {
	if len(items) == 0 {
		return nil
	}

	itemIDs := make([]int64, len(items))
	quantities := make([]int32, len(items))
	unitPrices := make([]float64, len(items))
	for i, item := range items {
		itemIDs[i] = item.ItemID
		quantities[i] = int32(item.Quantity)
		unitPrices[i] = item.UnitPrice
	}

	const query = `
		INSERT INTO order_items (order_id, item_id, quantity, unit_price)
		SELECT $1, unnest($2::bigint[]), unnest($3::int[]), unnest($4::numeric[])`

	if _, err := tx.ExecContext(ctx, query, orderID, pq.Array(itemIDs), pq.Array(quantities), pq.Array(unitPrices)); err != nil {
		return fmt.Errorf("insert order items: %w", err)
	}

	return nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id int64, status models.OrderStatus) (models.Order, error) {
	const query = `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, merchant_id, status, total_amount, created_at, updated_at`

	var order models.Order
	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&order.ID,
		&order.UserID,
		&order.MerchantID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Order{}, ErrNotFound
	}
	if err != nil {
		return models.Order{}, fmt.Errorf("update order status: %w", err)
	}

	return order, nil
}
