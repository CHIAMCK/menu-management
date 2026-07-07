package lock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisUserLocker_FirstLockSucceeds(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := NewRedisUserLocker(client, 5*time.Second)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("TryLock: want nil, got %v", err)
	}
}

func TestRedisUserLocker_SecondLockFailsWithinTTL(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := NewRedisUserLocker(client, 5*time.Second)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("first TryLock: want nil, got %v", err)
	}
	if err := locker.TryLock(context.Background(), 1); !errors.Is(err, ErrUserLocked) {
		t.Fatalf("second TryLock: want ErrUserLocked, got %v", err)
	}
}

func TestRedisUserLocker_DifferentUsersIndependent(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := NewRedisUserLocker(client, 5*time.Second)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("TryLock(1): want nil, got %v", err)
	}
	if err := locker.TryLock(context.Background(), 2); err != nil {
		t.Fatalf("TryLock(2): want nil, got %v", err)
	}
}

func TestRedisUserLocker_ExpiresAfterTTL(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := NewRedisUserLocker(client, 50*time.Millisecond)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("first TryLock: want nil, got %v", err)
	}

	mr.FastForward(60 * time.Millisecond)

	if err := locker.TryLock(context.Background(), 1); err != nil {
		t.Fatalf("TryLock after expiry: want nil, got %v", err)
	}
}
