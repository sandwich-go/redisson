package redisson

import "time"

const (
	defaultKeyPrefix    = "redislock"
	defaultKeyValidity  = 5 * time.Second
	defaultTryNextAfter = 20 * time.Millisecond
	// defaultKeyMajority Redlock 多 key 仲裁默认值；KeyMajority=2 意味着锁会分散到
	// N=KeyMajority*2-1=3 个 slot 上，需要至少 2 个 key 获取成功。
	// 单 Redis 实例足够保证可用性时可显式覆盖为 1，避免无谓的多 key 写入。
	defaultKeyMajority = int32(2)
)

//go:generate optiongen --option_with_struct_name=true --new_func=newLockerOptions --empty_composite_nil=true --usage_tag_name=usage
func LockerOptionsOptionDeclareWithDefault() any {
	return map[string]any{
		// annotation@KeyPrefix(KeyPrefix is the prefix of redis key for locks. Default value is defaultKeyPrefix)
		"KeyPrefix": string(defaultKeyPrefix),
		// annotation@KeyValidity(KeyValidity is the validity duration of locks. The lock holder will renew it periodically inside rueidislock to keep the lock alive. Default value is defaultKeyValidity)
		"KeyValidity": defaultKeyValidity,
		// annotation@TryNextAfter(TryNextAfter is the timeout duration before trying the next redis key for locks. Default value is defaultTryNextAfter)
		"TryNextAfter": defaultTryNextAfter,
		// annotation@KeyMajority(KeyMajority follows Redlock semantics: a lock spans N=KeyMajority*2-1 redis keys (each on its own slot for cluster), and needs at least KeyMajority keys to be acquired to be valid. Set to 1 if a single redis instance is enough. Default value is defaultKeyMajority)
		"KeyMajority": defaultKeyMajority,
		// annotation@NoLoopTracking(NoLoopTracking will use NOLOOP in the CLIENT TRACKING command to avoid unnecessary notifications and thus have better performance. This can only be enabled if all your redis nodes >= 7.0.5)
		"NoLoopTracking": false,
		// annotation@FallbackSETPX(Use SET PX instead of SET PXAT when acquiring locks to be compatible with Redis < 6.2)
		"FallbackSETPX": false,
	}
}
