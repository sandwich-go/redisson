package redisson

import (
	"context"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
)

// Locker is the interface of lock
type Locker interface {
	// WithContext acquires a distributed redis lock by name by waiting for it. It may return ErrLockerClosed.
	WithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error)
	// TryWithContext tries to acquire a distributed redis lock by name without waiting. It may return ErrNotLocked.
	TryWithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error)
	// ForceWithContext takes over a distributed redis lock by canceling the original holder. It may return ErrNotLocked.
	ForceWithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error)
}

const fallbackSETPXVersion = "6.2.0"

// newLocker 新键一个 locker
func newLocker(c *client, opts ...LockerOption) (Locker, error) {
	// 校验版本
	if c.version.LessThan(mustNewSemVersion(fallbackSETPXVersion)) {
		opts = append(opts, WithFallbackSETPX(true))
	}
	cc := newLockerOptions(opts...)
	return rueidislock.NewLocker(rueidislock.LockerOption{
		KeyPrefix:      cc.GetKeyPrefix(),
		KeyValidity:    cc.GetKeyValidity(),
		TryNextAfter:   cc.GetTryNextAfter(),
		KeyMajority:    cc.GetKeyMajority(),
		NoLoopTracking: cc.GetNoLoopTracking(),
		FallbackSETPX:  cc.GetFallbackSETPX(),
		ClientOption:   confVisitor2ClientOption(c.v),
		ClientBuilder: func(option rueidis.ClientOption) (rueidis.Client, error) {
			return c.cmd, nil
		},
	})
}

// NewLocker 新键一个 locker
func (c *client) NewLocker(opts ...LockerOption) (Locker, error) {
	return newLocker(c, opts...)
}
