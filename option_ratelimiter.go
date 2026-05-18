package redisson

import (
	"errors"
	"time"
)

const (
	defaultRateLimiterKeyPrefix = "redisratelimiter"
	// minRateLimiterWindow Window 最小精度：底层 rueidislimiter 要求 > 1ms。
	minRateLimiterWindow = time.Millisecond
)

var (
	// ErrRateLimiterInvalidLimit RateLimiterOptions.Limit 必须 > 0。
	ErrRateLimiterInvalidLimit = errors.New("rate limiter limit must be greater than 0")
	// ErrRateLimiterInvalidWindow RateLimiterOptions.Window 必须 > 1ms。
	ErrRateLimiterInvalidWindow = errors.New("rate limiter window must be greater than 1 millisecond")
)

//go:generate optiongen --option_with_struct_name=true --new_func=newRateLimiterOptions --empty_composite_nil=true --usage_tag_name=usage
func RateLimiterOptionsOptionDeclareWithDefault() any {
	return map[string]any{
		// annotation@KeyPrefix(Prefix for Redis keys used by this limiter)
		"KeyPrefix": string(defaultRateLimiterKeyPrefix),
		// annotation@Limit(Maximum number of allowed requests per window. Must be greater than 0; no default — caller must specify it before NewRateLimiter.)
		"Limit": int(0),
		// annotation@Window(Time window duration for rate limiting. Must be greater than 1 millisecond; no default — caller must specify it before NewRateLimiter.)
		"Window": time.Duration(0),
	}
}
