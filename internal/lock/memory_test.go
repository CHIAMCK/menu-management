package lock

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInMemoryUserLocker_FirstLockSucceeds(t *testing.T) {
	locker := NewInMemoryUserLocker(5 * time.Second)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("TryLock: want nil, got %v", err)
	}
}

func TestInMemoryUserLocker_SecondLockFailsWithinTTL(t *testing.T) {
	locker := NewInMemoryUserLocker(5 * time.Second)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("first TryLock: want nil, got %v", err)
	}
	if err := locker.TryLock(context.Background(), 1); !errors.Is(err, ErrUserLocked) {
		t.Fatalf("second TryLock: want ErrUserLocked, got %v", err)
	}
}

func TestInMemoryUserLocker_DifferentUsersIndependent(t *testing.T) {
	locker := NewInMemoryUserLocker(5 * time.Second)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("TryLock(1): want nil, got %v", err)
	}
	if err := locker.TryLock(context.Background(), 2); err != nil {
		t.Fatalf("TryLock(2): want nil, got %v", err)
	}
}

func TestInMemoryUserLocker_ExpiresAfterTTL(t *testing.T) {
	locker := NewInMemoryUserLocker(50 * time.Millisecond)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("first TryLock: want nil, got %v", err)
	}

	time.Sleep(60 * time.Millisecond)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("TryLock after expiry: want nil, got %v", err)
	}
}
