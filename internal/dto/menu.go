package dto

type MenuResponse struct {
	MenuID     int64              `json:"menu_id"`
	MenuName   string             `json:"menu_name"`
	MerchantID int64              `json:"merchant_id"`
	Categories []CategoryResponse `json:"categories"`
}

type CategoryResponse struct {
	ID    int64          `json:"id"`
	Name  string         `json:"name"`
	Items []ItemResponse `json:"items"`
}

type ItemResponse struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Availability string  `json:"availability"`
}
