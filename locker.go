package redisson

import (
	"context"
	"github.com/coreos/go-semver/semver"
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

const fallbackSETPXVersion = "6.2.0"

// newLocker 新键一个 locker
func newLocker(v ConfVisitor, version *semver.Version, opts ...LockerOption) (Locker, error) {
	clientOption := confVisitor2ClientOption(v)
	// 校验版本
	if version != nil && version.LessThan(mustNewSemVersion(fallbackSETPXVersion)) {
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
		ClientOption:   clientOption,
	})
}

// NewLocker 新键一个 locker
func (r *resp3) NewLocker(opts ...LockerOption) (Locker, error) {
	return newLocker(r.v, r.handler.getVersion(), opts...)
}
func (r *resp2) NewLocker(...LockerOption) (Locker, error) { panic("not implemented") }
