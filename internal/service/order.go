package service

import (
	"context"
	"errors"
	"fmt"

	"menu-management/internal/dto"
	"menu-management/internal/models"
	"menu-management/internal/repository"
)

type OrderRepository interface {
	FindOrderByID(ctx context.Context, id int64) (models.Order, error)
	FindOrderItemsByOrderID(ctx context.Context, orderID int64) ([]repository.OrderItemWithItem, error)
}

type OrderService struct {
	orderRepo OrderRepository
}

func NewOrderService(orderRepo OrderRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int64) (dto.OrderDetailResponse, error) {
	if id <= 0 {
		return dto.OrderDetailResponse{}, ErrInvalidOrderID
	}

	order, err := s.orderRepo.FindOrderByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return dto.OrderDetailResponse{}, ErrOrderNotFound
		default:
			return dto.OrderDetailResponse{}, fmt.Errorf("get order: %w", err)
		}
	}

	items, err := s.orderRepo.FindOrderItemsByOrderID(ctx, id)
	if err != nil {
		return dto.OrderDetailResponse{}, fmt.Errorf("get order items: %w", err)
	}

	return toOrderDetailResponse(order, items), nil
}

func toOrderDetailResponse(order models.Order, items []repository.OrderItemWithItem) dto.OrderDetailResponse {
	itemResponses := make([]dto.OrderItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, dto.OrderItemResponse{
			ID:        item.ID,
			ItemID:    item.ItemID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return dto.OrderDetailResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		MerchantID:  order.MerchantID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
		Items:       itemResponses,
	}
}
