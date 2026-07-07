package service

import "errors"

var (
	ErrActiveMenuNotFound = errors.New("active menu not found")
	ErrInvalidMerchantID  = errors.New("invalid merchant id")
	ErrInvalidItemID      = errors.New("invalid item id")
	ErrItemNotFound       = errors.New("item not found")
)
