package service

import (
	"context"
	"errors"
	"testing"

	"menu-management/internal/models"
	"menu-management/internal/repository"
)

type mockMenuRepository struct {
	menu          models.Menu
	menuErr       error
	categories    []models.Category
	categoriesErr error
	items         []models.Item
	itemsErr      error
}

func (m *mockMenuRepository) FindActiveMenuByMerchantID(_ context.Context, _ int64) (models.Menu, error) {
	if m.menuErr != nil {
		return models.Menu{}, m.menuErr
	}
	return m.menu, nil
}

func (m *mockMenuRepository) FindCategoriesByMenuID(_ context.Context, _ int64) ([]models.Category, error) {
	if m.categoriesErr != nil {
		return nil, m.categoriesErr
	}
	return m.categories, nil
}

func (m *mockMenuRepository) FindItemsByCategoryIDs(_ context.Context, _ []int64) ([]models.Item, error) {
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	return m.items, nil
}

func TestGetActiveMenu_InvalidMerchantID(t *testing.T) {
	svc := NewMenuService(&mockMenuRepository{})

	for _, merchantID := range []int64{0, -1} {
		_, err := svc.GetActiveMenu(context.Background(), merchantID)
		if !errors.Is(err, ErrInvalidMerchantID) {
			t.Fatalf("GetActiveMenu(%d): want ErrInvalidMerchantID, got %v", merchantID, err)
		}
	}
}

func TestGetActiveMenu_NotFound(t *testing.T) {
	svc := NewMenuService(&mockMenuRepository{
		menuErr: repository.ErrNotFound,
	})

	_, err := svc.GetActiveMenu(context.Background(), 1)
	if !errors.Is(err, ErrActiveMenuNotFound) {
		t.Fatalf("GetActiveMenu: want ErrActiveMenuNotFound, got %v", err)
	}
}

func TestGetActiveMenu_FindMenuError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewMenuService(&mockMenuRepository{
		menuErr: repoErr,
	})

	_, err := svc.GetActiveMenu(context.Background(), 1)
	if err == nil || errors.Is(err, ErrActiveMenuNotFound) {
		t.Fatalf("GetActiveMenu: want wrapped repo error, got %v", err)
	}
}

func TestGetActiveMenu_FindCategoriesError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewMenuService(&mockMenuRepository{
		menu:          models.Menu{ID: 1, MerchantID: 1, Name: "Lunch Menu"},
		categoriesErr: repoErr,
	})

	_, err := svc.GetActiveMenu(context.Background(), 1)
	if err == nil {
		t.Fatal("GetActiveMenu: expected error, got nil")
	}
}

func TestGetActiveMenu_FindItemsError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewMenuService(&mockMenuRepository{
		menu: models.Menu{ID: 1, MerchantID: 1, Name: "Lunch Menu"},
		categories: []models.Category{
			{ID: 10, Name: "Pizza"},
		},
		itemsErr: repoErr,
	})

	_, err := svc.GetActiveMenu(context.Background(), 1)
	if err == nil {
		t.Fatal("GetActiveMenu: expected error, got nil")
	}
}

func TestGetActiveMenu_Success(t *testing.T) {
	svc := NewMenuService(&mockMenuRepository{
		menu: models.Menu{ID: 1, MerchantID: 1, Name: "Lunch Menu"},
		categories: []models.Category{
			{ID: 10, Name: "Pizza"},
			{ID: 11, Name: "Sides"},
		},
		items: []models.Item{
			{ID: 1, CategoryID: 10, Name: "Margherita Pizza", Price: 12.99, Availability: models.ItemAvailabilityAvailable},
			{ID: 4, CategoryID: 11, Name: "Garlic Bread", Price: 5.99, Availability: models.ItemAvailabilityOutOfStock},
		},
	})

	got, err := svc.GetActiveMenu(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetActiveMenu: unexpected error: %v", err)
	}

	if got.ID != 1 || got.Name != "Lunch Menu" || got.MerchantID != 1 {
		t.Fatalf("unexpected menu fields: %+v", got)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("len(Categories) = %d, want 2", len(got.Categories))
	}
	if len(got.Categories[0].Items) != 1 || got.Categories[0].Items[0].Availability != "AVAILABLE" {
		t.Fatalf("unexpected pizza category: %+v", got.Categories[0])
	}
	if len(got.Categories[1].Items) != 1 || got.Categories[1].Items[0].Availability != "OUT_OF_STOCK" {
		t.Fatalf("unexpected sides category: %+v", got.Categories[1])
	}
}

func TestGetActiveMenu_SuccessEmptyCategories(t *testing.T) {
	svc := NewMenuService(&mockMenuRepository{
		menu:       models.Menu{ID: 1, MerchantID: 1, Name: "Lunch Menu"},
		categories: []models.Category{},
		items:      []models.Item{},
	})

	got, err := svc.GetActiveMenu(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetActiveMenu: unexpected error: %v", err)
	}
	if got.Categories == nil || len(got.Categories) != 0 {
		t.Fatalf("Categories = %v, want empty slice", got.Categories)
	}
}
