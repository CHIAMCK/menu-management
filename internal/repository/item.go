package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"menu-management/internal/models"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) FindItemByID(ctx context.Context, id int64) (models.Item, error) {
	const query = `
		SELECT id, merchant_id, name, price, status, availability, category_id, created_at, updated_at
		FROM items
		WHERE id = $1`

	var item models.Item
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.MerchantID,
		&item.Name,
		&item.Price,
		&item.Status,
		&item.Availability,
		&item.CategoryID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Item{}, ErrNotFound
	}
	if err != nil {
		return models.Item{}, fmt.Errorf("find item: %w", err)
	}

	return item, nil
}

func (r *ItemRepository) FindItemsByIDs(ctx context.Context, ids []int64) ([]models.Item, error) {
	if len(ids) == 0 {
		return []models.Item{}, nil
	}

	const query = `
		SELECT id, merchant_id, name, price, status, availability, category_id, created_at, updated_at
		FROM items
		WHERE id = ANY($1)`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("find items by ids: %w", err)
	}
	defer rows.Close()

	items := make([]models.Item, 0, len(ids))
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ID,
			&item.MerchantID,
			&item.Name,
			&item.Price,
			&item.Status,
			&item.Availability,
			&item.CategoryID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (r *ItemRepository) UpdateItemAvailability(ctx context.Context, id int64, availability models.ItemAvailability) (models.Item, error) {
	const query = `
		UPDATE items
		SET availability = $1
		WHERE id = $2
		RETURNING id, merchant_id, name, price, status, availability, category_id, created_at, updated_at`

	var item models.Item
	err := r.db.QueryRowContext(ctx, query, availability, id).Scan(
		&item.ID,
		&item.MerchantID,
		&item.Name,
		&item.Price,
		&item.Status,
		&item.Availability,
		&item.CategoryID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Item{}, ErrNotFound
	}
	if err != nil {
		return models.Item{}, fmt.Errorf("update item availability: %w", err)
	}

	return item, nil
}
