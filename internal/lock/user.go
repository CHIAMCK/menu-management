package lock

import (
	"errors"
	"sync"
	"time"
)

var ErrUserLocked = errors.New("user locked")

// UserLocker prevents duplicate order submissions per user.
// InMemoryUserLocker is per-process only; not shared across app replicas.
type UserLocker interface {
	TryLock(userID int64) error
}

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

func (l *InMemoryUserLocker) TryLock(userID int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for id, expiresAt := range l.locks {
		if now.After(expiresAt) {
			delete(l.locks, id)
		}
	}

	if expiresAt, ok := l.locks[userID]; ok && now.Before(expiresAt) {
		return ErrUserLocked
	}

	l.locks[userID] = now.Add(l.ttl)
	return nil
}
