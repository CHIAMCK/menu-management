package service

import "errors"

var (
	ErrActiveMenuNotFound = errors.New("active menu not found")
	ErrInvalidMerchantID  = errors.New("invalid merchant id")
)
