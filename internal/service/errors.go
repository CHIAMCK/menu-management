package service

import "errors"

var (
	ErrActiveMenuNotFound = errors.New("active menu not found")
	ErrInvalidMerchantID  = errors.New("invalid merchant id")
	ErrInvalidItemID            = errors.New("invalid item id")
	ErrItemNotFound             = errors.New("item not found")
	ErrInvalidItemAvailability  = errors.New("invalid item availability")
	ErrInvalidOrderID           = errors.New("invalid order id")
	ErrOrderNotFound            = errors.New("order not found")
	ErrInvalidOrderRequest      = errors.New("invalid order request")
	ErrDuplicateOrderItem       = errors.New("duplicate order item")
	ErrItemUnavailable          = errors.New("item unavailable")
	ErrItemMerchantMismatch     = errors.New("item merchant mismatch")
)
