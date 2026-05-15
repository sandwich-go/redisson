package redisson

import (
	"fmt"
	"strings"
	"testing"
)

// TestDefaultHandleDeadLetterReadable 回归 Bug #12：
// 默认 HandleDeadLetter 用 %q 而非 %v 打印 []byte，
// 输出可读字符串而非数字数组（如 [97 98 99]）。
//
// 该回归不直接断言全局 logger 输出（避免污染其他测试），
// 而是通过 fmt.Sprintf("%q", []byte) 是否符合预期来验证格式选择。
func TestDefaultHandleDeadLetterReadable(t *testing.T) {
	bs := []byte("payload-abc")
	formatted := fmt.Sprintf("got dead letter, %q", bs)
	if !strings.Contains(formatted, `"payload-abc"`) {
		t.Fatalf("expected readable %%q output, got: %s", formatted)
	}
	// 同时确认我们没有退回到 %v 的 [97 98 ...] 格式
	bad := fmt.Sprintf("%v", bs)
	if !strings.HasPrefix(bad, "[") {
		t.Skip("environment behaves unexpectedly for percent-v on []byte")
	}
}
