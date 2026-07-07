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
	CreateOrder(ctx context.Context, input repository.CreateOrderInput) (int64, error)
}

type OrderService struct {
	orderRepo OrderRepository
	itemRepo  ItemRepository
}

func NewOrderService(orderRepo OrderRepository, itemRepo ItemRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		itemRepo:  itemRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (dto.OrderDetailResponse, error) {
	if err := validateCreateOrderRequest(req); err != nil {
		return dto.OrderDetailResponse{}, err
	}

	itemIDs := make([]int64, 0, len(req.Items))
	for _, item := range req.Items {
		itemIDs = append(itemIDs, item.ItemID)
	}

	items, err := s.itemRepo.FindItemsByIDs(ctx, itemIDs)
	if err != nil {
		return dto.OrderDetailResponse{}, fmt.Errorf("find order items: %w", err)
	}

	itemsByID := make(map[int64]models.Item, len(items))
	for _, item := range items {
		itemsByID[item.ID] = item
	}

	orderItems := make([]repository.CreateOrderItemInput, 0, len(req.Items))
	var totalAmount float64

	for _, reqItem := range req.Items {
		item, ok := itemsByID[reqItem.ItemID]
		if !ok {
			return dto.OrderDetailResponse{}, ErrItemNotFound
		}
		if item.MerchantID != req.MerchantID {
			return dto.OrderDetailResponse{}, ErrItemMerchantMismatch
		}
		if item.Status != models.ItemStatusActive || item.Availability != models.ItemAvailabilityAvailable {
			return dto.OrderDetailResponse{}, ErrItemUnavailable
		}

		lineTotal := item.Price * float64(reqItem.Quantity)
		totalAmount += lineTotal
		orderItems = append(orderItems, repository.CreateOrderItemInput{
			ItemID:    reqItem.ItemID,
			Quantity:  reqItem.Quantity,
			UnitPrice: item.Price,
		})
	}

	orderID, err := s.orderRepo.CreateOrder(ctx, repository.CreateOrderInput{
		UserID:      req.UserID,
		MerchantID:  req.MerchantID,
		TotalAmount: totalAmount,
		Items:       orderItems,
	})
	if err != nil {
		return dto.OrderDetailResponse{}, fmt.Errorf("create order: %w", err)
	}

	return s.GetOrderByID(ctx, orderID)
}

func validateCreateOrderRequest(req dto.CreateOrderRequest) error {
	if req.UserID <= 0 || req.MerchantID <= 0 || len(req.Items) == 0 {
		return ErrInvalidOrderRequest
	}

	seenItemIDs := make(map[int64]struct{}, len(req.Items))
	for _, item := range req.Items {
		if item.ItemID <= 0 || item.Quantity <= 0 {
			return ErrInvalidOrderRequest
		}
		if _, exists := seenItemIDs[item.ItemID]; exists {
			return ErrDuplicateOrderItem
		}
		seenItemIDs[item.ItemID] = struct{}{}
	}

	return nil
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
