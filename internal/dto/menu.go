package dto

type MenuResponse struct {
	MenuID     int64              `json:"menu_id"`
	MenuName   string             `json:"menu_name"`
	MerchantID int64              `json:"merchant_id"`
	Status     string             `json:"status"`
	Categories []CategoryResponse `json:"categories"`
}

type CategoryResponse struct {
	ID     int64          `json:"id"`
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Items  []ItemResponse `json:"items"`
}

type ItemResponse struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Status       string  `json:"status"`
	Availability string  `json:"availability"`
}
