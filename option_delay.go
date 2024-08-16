package redisson

import (
	"fmt"
	"time"
)

//go:generate optiongen --option_with_struct_name=true --new_func=newDelayOptions --empty_composite_nil=true --usage_tag_name=usage
func DelayOptionsOptionDeclareWithDefault() any {
	return map[string]any{
		// annotation@Prefix(延迟队列前缀)
		"Prefix": "",
		// annotation@Timeout(业务处理超时时间，如果超过该时间未处理，则重试)
		"Timeout": time.Duration(1 * time.Minute),
		// annotation@RetryTimes(comment="重试次数，当业务处理超时，或业务处理返回错误，则重试")
		"RetryTimes": 3,
		// annotation@HandleDeadLetter(comment="处理死信，当达到最大重试次数，则为死信")
		"HandleDeadLetter": func(bs []byte) { warning(fmt.Sprintf("got dead letter, %v", bs)) },
	}
}
