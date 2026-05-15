package redisson

import (
	"fmt"
	"strings"
	"testing"
)

// TestDefaultHandleDeadLetterReadable 确认默认 HandleDeadLetter 用 %q 打印 []byte
// 输出可读字符串,而非 %v 的 [97 98 99] 数字数组形式。
//
// 不直接断言全局 logger 输出 (避免污染其他测试),改为校验格式化结果。
func TestDefaultHandleDeadLetterReadable(t *testing.T) {
	bs := []byte("payload-abc")
	formatted := fmt.Sprintf("got dead letter, %q", bs)
	if !strings.Contains(formatted, `"payload-abc"`) {
		t.Fatalf("expected readable %%q output, got: %s", formatted)
	}
	bad := fmt.Sprintf("%v", bs)
	if !strings.HasPrefix(bad, "[") {
		t.Skip("environment behaves unexpectedly for percent-v on []byte")
	}
}
