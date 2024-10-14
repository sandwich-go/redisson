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

type wrapLocker struct {
	v ConfInterface
	rueidislock.Locker
}

func (w *wrapLocker) wrap(ctx context.Context, name string, f func(ctx context.Context, name string) (context.Context, context.CancelFunc, error)) (context.Context, context.CancelFunc, error) {
	if _, ok := ctx.Deadline(); ok {
		return f(ctx, name)
	}
	ctx0, cancel0 := context.WithTimeout(ctx, w.v.GetWriteTimeout())
	ctx1, cancel1, err := f(ctx0, name)
	return ctx1, func() {
		if cancel1 != nil {
			cancel1()
		}
		cancel0()
	}, err
}

func (w *wrapLocker) WithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error) {
	return w.wrap(ctx, name, w.Locker.WithContext)
}

// TryWithContext tries to acquire a distributed redis lock by name without waiting. It may return ErrNotLocked.
func (w *wrapLocker) TryWithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error) {
	return w.wrap(ctx, name, w.Locker.TryWithContext)
}

// ForceWithContext takes over a distributed redis lock by canceling the original holder. It may return ErrNotLocked.
func (w *wrapLocker) ForceWithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error) {
	return w.wrap(ctx, name, w.Locker.ForceWithContext)
}

// newLocker 新建一个 locker
func newLocker(c *client, opts ...LockerOption) (Locker, error) {
	// 校验版本
	if c.version.LessThan(mustNewSemVersion(fallbackSETPXVersion)) {
		opts = append(opts, WithLockerOptionFallbackSETPX(true))
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

// NewLocker 新建一个 locker
func (c *client) NewLocker(opts ...LockerOption) (Locker, error) {
	return newLocker(c, opts...)
}
