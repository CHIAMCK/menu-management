package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"menu-management/internal/models"
)

func newMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return gormDB, mock
}

func TestCreateOrder_Success(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	input := CreateOrderInput{
		UserID:      1,
		MerchantID:  1,
		TotalAmount: 31.97,
		Items: []CreateOrderItemInput{
			{ItemID: 1, Quantity: 2, UnitPrice: 12.99},
			{ItemID: 4, Quantity: 1, UnitPrice: 5.99},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "orders"`).
		WithArgs(input.UserID, input.MerchantID, models.OrderStatusReceived, input.TotalAmount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectQuery(`INSERT INTO "order_items"`).
		WithArgs(int64(10), int64(1), 2, 12.99, sqlmock.AnyArg(), sqlmock.AnyArg(), int64(10), int64(4), 1, 5.99, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	mock.ExpectCommit()

	repo := NewOrderRepository(gormDB)
	orderID, err := repo.CreateOrder(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateOrder: unexpected error: %v", err)
	}
	if orderID != 10 {
		t.Fatalf("orderID = %d, want 10", orderID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateOrder_BeginTxFails(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	beginErr := errors.New("connection refused")
	mock.ExpectBegin().WillReturnError(beginErr)

	repo := NewOrderRepository(gormDB)
	_, err := repo.CreateOrder(context.Background(), CreateOrderInput{
		UserID:      1,
		MerchantID:  1,
		TotalAmount: 12.99,
		Items:       []CreateOrderItemInput{{ItemID: 1, Quantity: 1, UnitPrice: 12.99}},
	})
	if err == nil {
		t.Fatal("CreateOrder: expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateOrder_InsertOrderFails(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	insertErr := errors.New("insert order failed")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "orders"`).
		WithArgs(int64(1), int64(1), models.OrderStatusReceived, 12.99, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(insertErr)
	mock.ExpectRollback()

	repo := NewOrderRepository(gormDB)
	_, err := repo.CreateOrder(context.Background(), CreateOrderInput{
		UserID:      1,
		MerchantID:  1,
		TotalAmount: 12.99,
		Items:       []CreateOrderItemInput{{ItemID: 1, Quantity: 1, UnitPrice: 12.99}},
	})
	if err == nil {
		t.Fatal("CreateOrder: expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateOrder_InsertOrderItemFails(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	insertItemErr := errors.New("insert order item failed")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "orders"`).
		WithArgs(int64(1), int64(1), models.OrderStatusReceived, 31.97, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectQuery(`INSERT INTO "order_items"`).
		WillReturnError(insertItemErr)
	mock.ExpectRollback()

	repo := NewOrderRepository(gormDB)
	_, err := repo.CreateOrder(context.Background(), CreateOrderInput{
		UserID:      1,
		MerchantID:  1,
		TotalAmount: 31.97,
		Items: []CreateOrderItemInput{
			{ItemID: 1, Quantity: 2, UnitPrice: 12.99},
			{ItemID: 4, Quantity: 1, UnitPrice: 5.99},
		},
	})
	if err == nil {
		t.Fatal("CreateOrder: expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateOrder_CommitFails(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	commitErr := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "orders"`).
		WithArgs(int64(1), int64(1), models.OrderStatusReceived, 12.99, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectQuery(`INSERT INTO "order_items"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(commitErr)

	repo := NewOrderRepository(gormDB)
	_, err := repo.CreateOrder(context.Background(), CreateOrderInput{
		UserID:      1,
		MerchantID:  1,
		TotalAmount: 12.99,
		Items:       []CreateOrderItemInput{{ItemID: 1, Quantity: 1, UnitPrice: 12.99}},
	})
	if err == nil {
		t.Fatal("CreateOrder: expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateOrderStatus_Success(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "orders" SET "status"=$1,"updated_at"=NOW() WHERE id = $2 RETURNING *`)).
		WithArgs(models.OrderStatusPreparing, int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "merchant_id", "status", "total_amount", "created_at", "updated_at",
		}).AddRow(2, 2, 1, models.OrderStatusPreparing, 14.99, createdAt, updatedAt))

	repo := NewOrderRepository(gormDB)
	order, err := repo.UpdateOrderStatus(context.Background(), 2, models.OrderStatusPreparing)
	if err != nil {
		t.Fatalf("UpdateOrderStatus: unexpected error: %v", err)
	}

	if order.ID != 2 || order.UserID != 2 || order.MerchantID != 1 {
		t.Fatalf("unexpected order ids: %+v", order)
	}
	if order.Status != models.OrderStatusPreparing {
		t.Fatalf("Status = %q, want PREPARING", order.Status)
	}
	if order.TotalAmount != 14.99 {
		t.Fatalf("TotalAmount = %v, want 14.99", order.TotalAmount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateOrderStatus_NotFound(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "orders" SET "status"=$1,"updated_at"=NOW() WHERE id = $2 RETURNING *`)).
		WithArgs(models.OrderStatusPreparing, int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "merchant_id", "status", "total_amount", "created_at", "updated_at",
		}))

	repo := NewOrderRepository(gormDB)
	_, err := repo.UpdateOrderStatus(context.Background(), 99, models.OrderStatusPreparing)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateOrderStatus: want ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateOrderStatus_QueryError(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	queryErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "orders" SET "status"=$1,"updated_at"=NOW() WHERE id = $2 RETURNING *`)).
		WithArgs(models.OrderStatusReady, int64(1)).
		WillReturnError(queryErr)

	repo := NewOrderRepository(gormDB)
	_, err := repo.UpdateOrderStatus(context.Background(), 1, models.OrderStatusReady)
	if err == nil {
		t.Fatal("UpdateOrderStatus: expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
