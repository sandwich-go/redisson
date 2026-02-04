package redisson

import "time"

const (
	defaultRateLimiterKeyPrefix = "redisratelimiter"
)

//go:generate optiongen --option_with_struct_name=true --new_func=newRateLimiterOptions --empty_composite_nil=true --usage_tag_name=usage
func RateLimiterOptionsOptionDeclareWithDefault() any {
	return map[string]any{
		// annotation@KeyPrefix(Prefix for Redis keys used by this limiter)
		"KeyPrefix": string(defaultRateLimiterKeyPrefix),
		// annotation@Limit(Maximum number of allowed requests per window.)
		"Limit": int(0),
		// annotation@Window(Time window duration for rate limiting. Must be greater than 1 millisecond.)
		"Window": time.Duration(0),
	}
}
