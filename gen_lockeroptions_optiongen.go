// Code generated by optiongen. DO NOT EDIT.
// optiongen: github.com/timestee/optiongen

package redisson

import "time"

// LockerOptions should use newLockerOptions to initialize it
type LockerOptions struct {
	// annotation@KeyPrefix(KeyPrefix is the prefix of redis key for locks. Default value is defaultKeyPrefix)
	KeyPrefix string
	// annotation@KeyValidity(KeyValidity is the validity duration of locks and will be extended periodically by the ExtendInterval. Default value is defaultKeyValidity)
	KeyValidity time.Duration
	// annotation@TryNextAfter(TryNextAfter is the timeout duration before trying the next redis key for locks. Default value is defaultTryNextAfter)
	TryNextAfter time.Duration
	// annotation@KeyMajority(KeyMajority is at least how many redis keys in a total of KeyMajority*2-1 should be acquired to be a valid lock. Default value is defaultKeyMajority)
	KeyMajority int32
	// annotation@NoLoopTracking(NoLoopTracking will use NOLOOP in the CLIENT TRACKING command to avoid unnecessary notifications and thus have better performance. This can only be enabled if all your redis nodes >= 7.0.5)
	NoLoopTracking bool
	// annotation@FallbackSETPX(Use SET PX instead of SET PXAT when acquiring locks to be compatible with Redis < 6.2)
	FallbackSETPX bool
}

// newLockerOptions new LockerOptions
func newLockerOptions(opts ...LockerOption) *LockerOptions {
	cc := newDefaultLockerOptions()
	for _, opt := range opts {
		opt(cc)
	}
	if watchDogLockerOptions != nil {
		watchDogLockerOptions(cc)
	}
	return cc
}

// ApplyOption apply multiple new option and return the old ones
// sample:
// old := cc.ApplyOption(WithTimeout(time.Second))
// defer cc.ApplyOption(old...)
func (cc *LockerOptions) ApplyOption(opts ...LockerOption) []LockerOption {
	var previous []LockerOption
	for _, opt := range opts {
		previous = append(previous, opt(cc))
	}
	return previous
}

// LockerOption option func
type LockerOption func(cc *LockerOptions) LockerOption

// WithLockerOptionKeyPrefix option func for filed KeyPrefix
func WithLockerOptionKeyPrefix(v string) LockerOption {
	return func(cc *LockerOptions) LockerOption {
		previous := cc.KeyPrefix
		cc.KeyPrefix = v
		return WithLockerOptionKeyPrefix(previous)
	}
}

// WithLockerOptionKeyValidity option func for filed KeyValidity
func WithLockerOptionKeyValidity(v time.Duration) LockerOption {
	return func(cc *LockerOptions) LockerOption {
		previous := cc.KeyValidity
		cc.KeyValidity = v
		return WithLockerOptionKeyValidity(previous)
	}
}

// WithLockerOptionTryNextAfter option func for filed TryNextAfter
func WithLockerOptionTryNextAfter(v time.Duration) LockerOption {
	return func(cc *LockerOptions) LockerOption {
		previous := cc.TryNextAfter
		cc.TryNextAfter = v
		return WithLockerOptionTryNextAfter(previous)
	}
}

// WithLockerOptionKeyMajority option func for filed KeyMajority
func WithLockerOptionKeyMajority(v int32) LockerOption {
	return func(cc *LockerOptions) LockerOption {
		previous := cc.KeyMajority
		cc.KeyMajority = v
		return WithLockerOptionKeyMajority(previous)
	}
}

// WithLockerOptionNoLoopTracking option func for filed NoLoopTracking
func WithLockerOptionNoLoopTracking(v bool) LockerOption {
	return func(cc *LockerOptions) LockerOption {
		previous := cc.NoLoopTracking
		cc.NoLoopTracking = v
		return WithLockerOptionNoLoopTracking(previous)
	}
}

// WithLockerOptionFallbackSETPX option func for filed FallbackSETPX
func WithLockerOptionFallbackSETPX(v bool) LockerOption {
	return func(cc *LockerOptions) LockerOption {
		previous := cc.FallbackSETPX
		cc.FallbackSETPX = v
		return WithLockerOptionFallbackSETPX(previous)
	}
}

// InstallLockerOptionsWatchDog the installed func will called when newLockerOptions  called
func InstallLockerOptionsWatchDog(dog func(cc *LockerOptions)) { watchDogLockerOptions = dog }

// watchDogLockerOptions global watch dog
var watchDogLockerOptions func(cc *LockerOptions)

// setLockerOptionsDefaultValue default LockerOptions value
func setLockerOptionsDefaultValue(cc *LockerOptions) {
	for _, opt := range [...]LockerOption{
		WithLockerOptionKeyPrefix(defaultKeyPrefix),
		WithLockerOptionKeyValidity(defaultKeyValidity),
		WithLockerOptionTryNextAfter(defaultTryNextAfter),
		WithLockerOptionKeyMajority(defaultKeyMajority),
		WithLockerOptionNoLoopTracking(false),
		WithLockerOptionFallbackSETPX(false),
	} {
		opt(cc)
	}
}

// newDefaultLockerOptions new default LockerOptions
func newDefaultLockerOptions() *LockerOptions {
	cc := &LockerOptions{}
	setLockerOptionsDefaultValue(cc)
	return cc
}

// all getter func
func (cc *LockerOptions) GetKeyPrefix() string           { return cc.KeyPrefix }
func (cc *LockerOptions) GetKeyValidity() time.Duration  { return cc.KeyValidity }
func (cc *LockerOptions) GetTryNextAfter() time.Duration { return cc.TryNextAfter }
func (cc *LockerOptions) GetKeyMajority() int32          { return cc.KeyMajority }
func (cc *LockerOptions) GetNoLoopTracking() bool        { return cc.NoLoopTracking }
func (cc *LockerOptions) GetFallbackSETPX() bool         { return cc.FallbackSETPX }

// LockerOptionsVisitor visitor interface for LockerOptions
type LockerOptionsVisitor interface {
	GetKeyPrefix() string
	GetKeyValidity() time.Duration
	GetTryNextAfter() time.Duration
	GetKeyMajority() int32
	GetNoLoopTracking() bool
	GetFallbackSETPX() bool
}

// LockerOptionsInterface visitor + ApplyOption interface for LockerOptions
type LockerOptionsInterface interface {
	LockerOptionsVisitor
	ApplyOption(...LockerOption) []LockerOption
}
