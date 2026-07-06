package service

import (
	"context"
	"errors"
	"fmt"

	"menu-management/internal/dto"
	"menu-management/internal/models"
	"menu-management/internal/repository"
)

type MenuRepository interface {
	FindActiveMenuByMerchantID(ctx context.Context, merchantID int64) (models.Menu, error)
	FindCategoriesByMenuID(ctx context.Context, menuID int64) ([]models.Category, error)
	FindItemsByCategoryIDs(ctx context.Context, categoryIDs []int64) ([]models.Item, error)
}

type MenuService struct {
	menuRepo MenuRepository
}

func NewMenuService(menuRepo MenuRepository) *MenuService {
	return &MenuService{menuRepo: menuRepo}
}

func (s *MenuService) GetActiveMenu(ctx context.Context, merchantID int64) (dto.MenuResponse, error) {
	if merchantID <= 0 {
		return dto.MenuResponse{}, ErrInvalidMerchantID
	}

	menu, err := s.menuRepo.FindActiveMenuByMerchantID(ctx, merchantID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return dto.MenuResponse{}, ErrActiveMenuNotFound
		default:
			return dto.MenuResponse{}, fmt.Errorf("get active menu: %w", err)
		}
	}

	categories, err := s.menuRepo.FindCategoriesByMenuID(ctx, menu.ID)
	if err != nil {
		return dto.MenuResponse{}, fmt.Errorf("get categories: %w", err)
	}

	categoryIDs := make([]int64, 0, len(categories))
	for _, category := range categories {
		categoryIDs = append(categoryIDs, category.ID)
	}

	items, err := s.menuRepo.FindItemsByCategoryIDs(ctx, categoryIDs)
	if err != nil {
		return dto.MenuResponse{}, fmt.Errorf("get items: %w", err)
	}

	return toMenuResponse(menu, categories, items), nil
}

func toMenuResponse(menu models.Menu, categories []models.Category, items []models.Item) dto.MenuResponse {
	itemsByCategory := make(map[int64][]dto.ItemResponse, len(categories))
	for _, category := range categories {
		itemsByCategory[category.ID] = []dto.ItemResponse{}
	}

	for _, item := range items {
		itemsByCategory[item.CategoryID] = append(itemsByCategory[item.CategoryID], dto.ItemResponse{
			ID:           item.ID,
			Name:         item.Name,
			Price:        item.Price,
			Status:       string(item.Status),
			Availability: string(item.Availability),
		})
	}

	categoryResponses := make([]dto.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		categoryResponses = append(categoryResponses, dto.CategoryResponse{
			ID:     category.ID,
			Name:   category.Name,
			Status: string(category.Status),
			Items:  itemsByCategory[category.ID],
		})
	}

	return dto.MenuResponse{
		MenuID:     menu.ID,
		MenuName:   menu.Name,
		MerchantID: menu.MerchantID,
		Status:     string(menu.Status),
		Categories: categoryResponses,
	}
}
