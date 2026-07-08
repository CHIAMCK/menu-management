package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"menu-management/internal/models"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) FindActiveMenuByMerchantID(ctx context.Context, merchantID int64) (models.Menu, error) {
	var menu models.Menu
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND status = ?", merchantID, models.MenuStatusActive).
		First(&menu).Error
	if err != nil {
		if mapped := mapRecordNotFound(err); mapped != err {
			return models.Menu{}, mapped
		}
		return models.Menu{}, fmt.Errorf("find active menu: %w", err)
	}

	return menu, nil
}

func (r *MenuRepository) FindCategoriesByMenuID(ctx context.Context, menuID int64) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).
		Select("id", "name", "status").
		Where("menu_id = ? AND status = ?", menuID, models.CategoryStatusActive).
		Order("created_at").
		Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("find categories: %w", err)
	}

	return categories, nil
}

func (r *MenuRepository) FindItemsByCategoryIDs(ctx context.Context, categoryIDs []int64) ([]models.Item, error) {
	if len(categoryIDs) == 0 {
		return []models.Item{}, nil
	}

	var items []models.Item
	err := r.db.WithContext(ctx).
		Select("id", "name", "price", "status", "availability", "category_id").
		Where("category_id IN ? AND status = ?", categoryIDs, models.ItemStatusActive).
		Order("category_id, created_at").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("find items: %w", err)
	}

	return items, nil
}
