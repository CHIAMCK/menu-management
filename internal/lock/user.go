package lock

import (
	"context"
	"errors"
)

var ErrUserLocked = errors.New("user locked")

// UserLocker prevents duplicate order submissions per user.
type UserLocker interface {
	TryLock(ctx context.Context, userID int64) error
}
