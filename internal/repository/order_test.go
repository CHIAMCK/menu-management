package repository

import (
	"context"
	"errors"
	"testing"

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
		WithArgs(input.UserID, input.MerchantID, models.OrderStatusPending, input.TotalAmount).
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
		WithArgs(int64(1), int64(1), models.OrderStatusPending, 12.99).
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
		WithArgs(int64(1), int64(1), models.OrderStatusPending, 31.97).
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
		WithArgs(int64(1), int64(1), models.OrderStatusPending, 12.99).
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
