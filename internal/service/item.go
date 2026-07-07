package service

import (
	"context"
	"errors"
	"fmt"

	"menu-management/internal/dto"
	"menu-management/internal/models"
	"menu-management/internal/repository"
)

type ItemRepository interface {
	FindItemByID(ctx context.Context, id int64) (models.Item, error)
	UpdateItemAvailability(ctx context.Context, id int64, availability models.ItemAvailability) (models.Item, error)
}

type ItemService struct {
	itemRepo ItemRepository
}

func NewItemService(itemRepo ItemRepository) *ItemService {
	return &ItemService{itemRepo: itemRepo}
}

func (s *ItemService) GetItemByID(ctx context.Context, id int64) (dto.ItemDetailResponse, error) {
	if id <= 0 {
		return dto.ItemDetailResponse{}, ErrInvalidItemID
	}

	item, err := s.itemRepo.FindItemByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return dto.ItemDetailResponse{}, ErrItemNotFound
		default:
			return dto.ItemDetailResponse{}, fmt.Errorf("get item: %w", err)
		}
	}

	return toItemDetailResponse(item), nil
}

func (s *ItemService) UpdateItemAvailability(ctx context.Context, id int64, availability string) (dto.ItemDetailResponse, error) {
	if id <= 0 {
		return dto.ItemDetailResponse{}, ErrInvalidItemID
	}

	itemAvailability := models.ItemAvailability(availability)
	if itemAvailability != models.ItemAvailabilityAvailable && itemAvailability != models.ItemAvailabilityOutOfStock {
		return dto.ItemDetailResponse{}, ErrInvalidItemAvailability
	}

	item, err := s.itemRepo.UpdateItemAvailability(ctx, id, itemAvailability)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return dto.ItemDetailResponse{}, ErrItemNotFound
		default:
			return dto.ItemDetailResponse{}, fmt.Errorf("update item availability: %w", err)
		}
	}

	return toItemDetailResponse(item), nil
}

func toItemDetailResponse(item models.Item) dto.ItemDetailResponse {
	return dto.ItemDetailResponse{
		ID:           item.ID,
		MerchantID:   item.MerchantID,
		Name:         item.Name,
		Price:        item.Price,
		Status:       string(item.Status),
		Availability: string(item.Availability),
		CategoryID:   item.CategoryID,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}
