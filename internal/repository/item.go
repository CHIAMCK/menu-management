package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"menu-management/internal/models"
)

type ItemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) FindItemByID(ctx context.Context, id int64) (models.Item, error) {
	var item models.Item
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if mapped := mapRecordNotFound(err); mapped != err {
			return models.Item{}, mapped
		}
		return models.Item{}, fmt.Errorf("find item: %w", err)
	}

	return item, nil
}

func (r *ItemRepository) FindItemsForOrder(ctx context.Context, ids []int64) ([]models.Item, error) {
	if len(ids) == 0 {
		return []models.Item{}, nil
	}

	var items []models.Item
	err := r.db.WithContext(ctx).
		Select("id", "merchant_id", "name", "price", "status", "availability").
		Where("id IN ?", ids).
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("find items for order: %w", err)
	}

	return items, nil
}

func (r *ItemRepository) UpdateItemAvailability(ctx context.Context, id int64, availability models.ItemAvailability) (models.Item, error) {
	var item models.Item
	result := r.db.WithContext(ctx).
		Model(&item).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Update("availability", availability)
	if result.Error != nil {
		return models.Item{}, fmt.Errorf("update item availability: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.Item{}, ErrNotFound
	}

	return item, nil
}
