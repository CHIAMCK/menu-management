package locktest

import (
	"testing"
	"time"

	"menu-management/internal/lock"
)

func NewUserLocker(t testing.TB) lock.UserLocker {
	t.Helper()

	return lock.NewInMemoryUserLocker(5 * time.Second)
}
