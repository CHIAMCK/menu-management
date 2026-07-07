package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"menu-management/internal/models"
)

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) FindActiveMenuByMerchantID(ctx context.Context, merchantID int64) (models.Menu, error) {
	const query = `
		SELECT id, merchant_id, name, status
		FROM menus
		WHERE merchant_id = $1 AND status = 'ACTIVE'
		LIMIT 1`

	var menu models.Menu
	err := r.db.QueryRowContext(ctx, query, merchantID).Scan(
		&menu.ID,
		&menu.MerchantID,
		&menu.Name,
		&menu.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Menu{}, ErrNotFound
	}
	if err != nil {
		return models.Menu{}, fmt.Errorf("find active menu: %w", err)
	}

	return menu, nil
}

func (r *MenuRepository) FindCategoriesByMenuID(ctx context.Context, menuID int64) ([]models.Category, error) {
	const query = `
		SELECT id, name, status
		FROM categories
		WHERE menu_id = $1 AND status = 'ACTIVE'
		ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, menuID)
	if err != nil {
		return nil, fmt.Errorf("find categories: %w", err)
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Status,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

func (r *MenuRepository) FindItemsByCategoryIDs(ctx context.Context, categoryIDs []int64) ([]models.Item, error) {
	if len(categoryIDs) == 0 {
		return []models.Item{}, nil
	}

	const query = `
		SELECT id, name, price, status, availability, category_id
		FROM items
		WHERE category_id = ANY($1) AND status = 'ACTIVE'
		ORDER BY category_id, created_at`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(categoryIDs))
	if err != nil {
		return nil, fmt.Errorf("find items: %w", err)
	}
	defer rows.Close()

	items := make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.Status,
			&item.Availability,
			&item.CategoryID,
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
