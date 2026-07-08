package lock

import (
	"context"
	"sync"
	"time"
)

type InMemoryUserLocker struct {
	mu    sync.Mutex
	locks map[int64]time.Time
	ttl   time.Duration
}

func NewInMemoryUserLocker(ttl time.Duration) *InMemoryUserLocker {
	return &InMemoryUserLocker{
		locks: make(map[int64]time.Time),
		ttl:   ttl,
	}
}

func (l *InMemoryUserLocker) TryLock(_ context.Context, userID int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if expiry, ok := l.locks[userID]; ok && now.Before(expiry) {
		return ErrUserLocked
	}

	l.locks[userID] = now.Add(l.ttl)
	return nil
}
