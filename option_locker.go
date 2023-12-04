package redisson

import "time"

const (
	defaultKeyPrefix      = "redislock"
	defaultKeyValidity    = 5 * time.Second
	defaultExtendInterval = 1 * time.Second
	defaultTryNextAfter   = 20 * time.Millisecond
	defaultKeyMajority    = int32(2)
)

//go:generate optiongen --option_with_struct_name=false --new_func=newLockerOptions --empty_composite_nil=true --usage_tag_name=usage
func LockerOptionsOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		// annotation@KeyPrefix(KeyPrefix is the prefix of redis key for locks. Default value is defaultKeyPrefix)
		"KeyPrefix": string(defaultKeyPrefix),
		// annotation@KeyValidity(KeyValidity is the validity duration of locks and will be extended periodically by the ExtendInterval. Default value is defaultKeyValidity)
		"KeyValidity": time.Duration(defaultKeyValidity),
		// annotation@TryNextAfter(TryNextAfter is the timeout duration before trying the next redis key for locks. Default value is defaultTryNextAfter)
		"TryNextAfter": time.Duration(defaultTryNextAfter),
		// annotation@KeyMajority(KeyMajority is at least how many redis keys in a total of KeyMajority*2-1 should be acquired to be a valid lock. Default value is defaultKeyMajority)
		"KeyMajority": int32(defaultKeyMajority),
		// annotation@NoLoopTracking(NoLoopTracking will use NOLOOP in the CLIENT TRACKING command to avoid unnecessary notifications and thus have better performance. This can only be enabled if all your redis nodes >= 7.0.5)
		"NoLoopTracking": false,
		// annotation@FallbackSETPX(Use SET PX instead of SET PXAT when acquiring locks to be compatible with Redis < 6.2)
		"FallbackSETPX": false,
	}
}
