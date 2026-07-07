package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisUserLocker struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisUserLocker(client *redis.Client, ttl time.Duration) *RedisUserLocker {
	return &RedisUserLocker{
		client: client,
		ttl:    ttl,
	}
}

func (l *RedisUserLocker) TryLock(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("order:user-lock:%d", userID)
	ok, err := l.client.SetNX(ctx, key, "1", l.ttl).Result()
	if err != nil {
		return fmt.Errorf("redis setnx: %w", err)
	}

	if !ok {
		return ErrUserLocked
	}

	return nil
}
