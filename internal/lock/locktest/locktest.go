package locktest

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"menu-management/internal/lock"
)

func NewUserLocker(t testing.TB) lock.UserLocker {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return lock.NewRedisUserLocker(client, 5*time.Second)
}
