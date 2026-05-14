package redisson

import (
	"errors"
	"fmt"
	"strings"

	"github.com/redis/rueidis"
)

// 包级 sentinel error，用于在调用方使用 errors.Is 进行可靠判别。
//
// 这些 sentinel 同时具备稳定的 Error() 字符串：当 rueidis 返回的原始错误
// 文案与之匹配时，可被对应的 IsXxx 函数识别。
var (
	// ErrNoScript 对应 Redis 服务端 "NOSCRIPT" 响应。
	// 当通过 EVALSHA 调用而脚本未在缓存中时会返回。
	ErrNoScript = errors.New("NOSCRIPT No matching script. Please use EVAL")

	// ErrNoCache 对应 rueidis.ErrNoCache，当客户端缓存被禁用但执行 cache 路径时返回。
	// 提供出来方便用户用 errors.Is(err, redisson.ErrNoCache) 判别。
	ErrNoCache = rueidis.ErrNoCache
)

// ParameterError 表示 builder 参数校验失败，调用方传入了非法的 enum / 范围 / 互斥参数等。
//
// 在 builder 阶段以 panic 抛出（属于编程错误，应在测试中立即发现），
// 但通过自定义类型而非 string，便于 recover() 时识别和处理：
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        if pe, ok := r.(*ParameterError); ok {
//	            // 处理参数错误
//	            _ = pe
//	        }
//	    }
//	}()
type ParameterError struct {
	Reason string
}

// Error 实现 error 接口。
func (e *ParameterError) Error() string { return "redisson: " + e.Reason }

// NewParameterError 构造 ParameterError。
func NewParameterError(format string, args ...any) *ParameterError {
	return &ParameterError{Reason: fmt.Sprintf(format, args...)}
}

// IsParameterError 判别 panic 抛出的 r 是否为 ParameterError。
func IsParameterError(r any) bool {
	_, ok := r.(*ParameterError)
	return ok
}

// IsNoScriptError 判别是否为 NOSCRIPT 错误。
// 优先使用 errors.Is，回退到字符串前缀匹配以兼容 rueidis 直接返回的协议错误。
func IsNoScriptError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNoScript) {
		return true
	}
	return strings.HasPrefix(err.Error(), "NOSCRIPT ")
}

// ErrorFormatFunc 格式化 error 数组
// 调用 Error 时，会将  error 数组进行格式化，默认 ListFormatFunc
// 可以通过 SetFormatFunc 函数进行设置
type ErrorFormatFunc func([]error) string

// Errors 错误数组，可以将多个 error 进行组装，并当成 error 进行函数传递或返回
type Errors struct {
	errors     []error
	formatFunc ErrorFormatFunc
}

// Error 实现 error 接口
func (e *Errors) Error() string {
	fn := e.formatFunc
	if fn == nil {
		fn = ListFormatFunc
	}
	return fn(e.errors)
}

// Push 推入一个错误信息，err如果为nil则丢弃
func (e *Errors) Push(err error) {
	if err == nil {
		return
	}
	e.errors = append(e.errors, err)
}

// LastErr 返回最后一个错误信息，如果没有错误则返回nil
func (e *Errors) LastErr() error {
	if e == nil || len(e.errors) == 0 {
		return nil
	}
	return e.errors[len(e.errors)-1]
}

// Err 返回标准error对象，如果错误列表为空则返回nil
func (e *Errors) Err() error {
	if e == nil || len(e.errors) == 0 {
		return nil
	}
	return e
}

// String
func (e *Errors) String() string { return fmt.Sprintf("*%#v", *e) }

// WrappedErrors 返回内部所有的 error
func (e *Errors) WrappedErrors() []error { return e.errors }

// SetFormatFunc 设置格式化 error 数组函数，默认 ListFormatFunc
func (e *Errors) SetFormatFunc(f ErrorFormatFunc) { e.formatFunc = f }

// DotFormatFunc 多个 error，通过 ',' 进行分割输出
// 如输出: error 1,error 2
var DotFormatFunc = func(es []error) string {
	var errStr = make([]string, 0)
	for i := 0; i < len(es); i++ {
		errStr = append(errStr, es[i].Error())
	}
	return strings.Join(errStr, ",")
}

// ListFormatFunc 多个 error，列表输出
// 如输出: 2 errors occurred:
//
//	#1: error 1
//	#2: error 2
var ListFormatFunc = func(es []error) string {
	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("#%d: %s", i+1, err)
	}
	return fmt.Sprintf(
		"%d errors occurred:\n%s",
		len(es), strings.Join(points, "\n"))
}

// isNoScriptError 包内向后兼容别名；新代码请使用 IsNoScriptError。
func isNoScriptError(err error) bool { return IsNoScriptError(err) }
