package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"menu-management/internal/models"
)

func TestCreateOrder_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

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
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(input.UserID, input.MerchantID, models.OrderStatusReceived, input.TotalAmount).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec(`INSERT INTO order_items`).
		WithArgs(int64(10), input.Items[0].ItemID, input.Items[0].Quantity, input.Items[0].UnitPrice).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO order_items`).
		WithArgs(int64(10), input.Items[1].ItemID, input.Items[1].Quantity, input.Items[1].UnitPrice).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	repo := NewOrderRepository(db)
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	beginErr := errors.New("connection refused")
	mock.ExpectBegin().WillReturnError(beginErr)

	repo := NewOrderRepository(db)
	_, err = repo.CreateOrder(context.Background(), CreateOrderInput{
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	insertErr := errors.New("insert order failed")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(int64(1), int64(1), models.OrderStatusReceived, 12.99).
		WillReturnError(insertErr)
	mock.ExpectRollback()

	repo := NewOrderRepository(db)
	_, err = repo.CreateOrder(context.Background(), CreateOrderInput{
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	insertItemErr := errors.New("insert order item failed")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(int64(1), int64(1), models.OrderStatusReceived, 31.97).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec(`INSERT INTO order_items`).
		WithArgs(int64(10), int64(1), 2, 12.99).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO order_items`).
		WithArgs(int64(10), int64(4), 1, 5.99).
		WillReturnError(insertItemErr)
	mock.ExpectRollback()

	repo := NewOrderRepository(db)
	_, err = repo.CreateOrder(context.Background(), CreateOrderInput{
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	commitErr := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(int64(1), int64(1), models.OrderStatusReceived, 12.99).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec(`INSERT INTO order_items`).
		WithArgs(int64(10), int64(1), 1, 12.99).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(commitErr)

	repo := NewOrderRepository(db)
	_, err = repo.CreateOrder(context.Background(), CreateOrderInput{
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`UPDATE orders`).
		WithArgs(int64(2), models.OrderStatusPreparing).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "merchant_id", "status", "total_amount", "created_at", "updated_at",
		}).AddRow(2, 2, 1, models.OrderStatusPreparing, 14.99, createdAt, updatedAt))

	repo := NewOrderRepository(db)
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`UPDATE orders`).
		WithArgs(int64(99), models.OrderStatusPreparing).
		WillReturnError(sql.ErrNoRows)

	repo := NewOrderRepository(db)
	_, err = repo.UpdateOrderStatus(context.Background(), 99, models.OrderStatusPreparing)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateOrderStatus: want ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateOrderStatus_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queryErr := errors.New("connection refused")
	mock.ExpectQuery(`UPDATE orders`).
		WithArgs(int64(1), models.OrderStatusReady).
		WillReturnError(queryErr)

	repo := NewOrderRepository(db)
	_, err = repo.UpdateOrderStatus(context.Background(), 1, models.OrderStatusReady)
	if err == nil {
		t.Fatal("UpdateOrderStatus: expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
