package redisson

import (
	"context"
	"github.com/redis/rueidis/rueidislock"
)

// Locker is the interface of lock
type Locker interface {
	// WithContext acquires a distributed redis lock by name by waiting for it. It may return ErrLockerClosed.
	WithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error)
	// TryWithContext tries to acquire a distributed redis lock by name without waiting. It may return ErrNotLocked.
	TryWithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error)
	// Close closes the underlying Client
	Close()
}

// newLocker 新键一个 locker
func newLocker(v ConfVisitor, opts ...LockerOption) (Locker, error) {
	clientOption := confVisitor2ClientOption(v)
	cc := newLockerOptions(opts...)
	return rueidislock.NewLocker(rueidislock.LockerOption{
		KeyPrefix:      cc.GetKeyPrefix(),
		KeyValidity:    cc.GetKeyValidity(),
		TryNextAfter:   cc.GetTryNextAfter(),
		KeyMajority:    cc.GetKeyMajority(),
		NoLoopTracking: cc.GetNoLoopTracking(),
		FallbackSETPX:  cc.GetFallbackSETPX(),
		ClientOption:   clientOption,
	})
}

// NewLocker 新键一个 locker
func (c *client) NewLocker(opts ...LockerOption) (Locker, error) {
	return newLocker(c.v, opts...)
}
