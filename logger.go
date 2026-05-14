package redisson

import (
	"fmt"
	"os"
	"sync/atomic"
)

// Logger 是 redisson 内部使用的最小日志接口。
// 用户可通过 SetLogger 替换为 zap/zerolog/slog 等实现。
//
// 实现需要保证并发安全。
type Logger interface {
	// Warnf 输出告警级别日志。
	Warnf(format string, args ...any)
	// Errorf 输出错误级别日志。
	Errorf(format string, args ...any)
}

// stderrLogger 是默认日志实现，写入 stderr。
// 这避免了把库内部日志污染到调用方 stdout（影响 CLI 工具的输出协议）。
type stderrLogger struct{}

func (stderrLogger) Warnf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "[redisson][WARN] "+format+"\n", args...)
}
func (stderrLogger) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "[redisson][ERROR] "+format+"\n", args...)
}

// NopLogger 丢弃所有日志，可在生产环境用 SetLogger(NopLogger{}) 关闭日志。
type NopLogger struct{}

func (NopLogger) Warnf(string, ...any)  {}
func (NopLogger) Errorf(string, ...any) {}

// loggerBox 给 atomic.Value 提供单一具体类型，规避不同实现存入同一 atomic.Value
// 时触发的 "store of inconsistently typed value" panic。
type loggerBox struct{ l Logger }

// loggerHolder 通过 atomic.Value 持有当前 Logger，保证 SetLogger 与读取并发安全。
var loggerHolder atomic.Value // loggerBox

func init() {
	loggerHolder.Store(loggerBox{l: stderrLogger{}})
}

// SetLogger 替换全局 logger。传 nil 等价于 NopLogger。
// 可在程序启动阶段调用一次；运行期调用也是安全的。
func SetLogger(l Logger) {
	if l == nil {
		l = NopLogger{}
	}
	loggerHolder.Store(loggerBox{l: l})
}

// GetLogger 返回当前 logger，便于库使用方查询或在自定义 logger 中委托。
func GetLogger() Logger {
	if v := loggerHolder.Load(); v != nil {
		if box, ok := v.(loggerBox); ok && box.l != nil {
			return box.l
		}
	}
	return stderrLogger{}
}

// warning 与 e 是库内部的日志快捷方式（保留旧名以减少改动面）。
// 它们都委托到当前 Logger，默认写 stderr 而非 stdout。
func warning(msg string) { GetLogger().Warnf("%s", msg) }
func e(msg string)       { GetLogger().Errorf("%s", msg) }
