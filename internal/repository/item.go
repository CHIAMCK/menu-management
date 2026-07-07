package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
