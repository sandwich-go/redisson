package redisson

import (
	"fmt"
	"time"
)

const (
	// defaultDelayVisibilityTimeout 任务被 worker 拉走后默认的可见性超时；
	// 超过该时间仍未 ack 则被 reclaim 重新入 delay。
	// 1 分钟覆盖大多数业务 callback；超长任务请显式覆盖。
	defaultDelayVisibilityTimeout = time.Minute
	// defaultDelayRetryTimes 业务失败时默认的最大重试次数；超过则触发死信 hook。
	defaultDelayRetryTimes = 3
	// defaultDelayPollInterval 默认 ticker 间隔（poll 与 reclaim 共用）。
	defaultDelayPollInterval = time.Second
	// defaultDelayPollBatch 单次 poll/reclaim 处理的最大 item 数；防止单次 Lua 脚本耗时过长。
	defaultDelayPollBatch = 64
	// defaultDelayRedisOpTimeout 单次 Redis 操作（包括 Lua 脚本）的兜底超时；
	// 与业务 callback 超时（VisibilityTimeout）解耦，避免 visibility 设很大时 Redis 操作也跟着挂很久。
	defaultDelayRedisOpTimeout = 5 * time.Second
	// defaultDelayRetryBackoff 业务失败后重新可见的间隔；与 visibility timeout 解耦。
	defaultDelayRetryBackoff = time.Second
)

//go:generate optiongen --option_with_struct_name=true --new_func=newDelayOptions --empty_composite_nil=true --usage_tag_name=usage
func DelayOptionsOptionDeclareWithDefault() any {
	return map[string]any{
		// annotation@Prefix(延迟队列前缀)
		"Prefix": "",
		// annotation@VisibilityTimeout(任务被 worker 拉走后的可见性超时；worker 在该时间内未 ack 则任务被 reclaim 重新入 delay。建议覆盖业务 callback 实际耗时上限。)
		"VisibilityTimeout": defaultDelayVisibilityTimeout,
		// annotation@RetryTimes(comment="重试次数，当业务处理超时，或业务处理返回错误，则重试")
		"RetryTimes": defaultDelayRetryTimes,
		// annotation@HandleDeadLetter(comment="处理死信，当达到最大重试次数，则为死信")
		"HandleDeadLetter": func(bs []byte) { warning(fmt.Sprintf("got dead letter, %q", bs)) },
		// annotation@PollInterval(poll/reclaim ticker 触发间隔；过小会增加 Redis 压力，过大会让到期任务的派发延迟变高。0 表示使用 defaultDelayPollInterval。)
		"PollInterval": time.Duration(0),
		// annotation@PollBatch(单次 poll/reclaim 处理的最大 item 数；0 表示使用 defaultDelayPollBatch。)
		"PollBatch": int(0),
		// annotation@RedisOpTimeout(单次 Redis 操作的兜底超时（与 VisibilityTimeout 解耦）；0 表示使用 defaultDelayRedisOpTimeout。)
		"RedisOpTimeout": time.Duration(0),
		// annotation@RetryBackoff(业务失败后重新可见的间隔；0 表示使用 defaultDelayRetryBackoff。)
		"RetryBackoff": time.Duration(0),
	}
}
