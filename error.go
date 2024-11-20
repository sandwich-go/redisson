package redisson

import (
	"fmt"
	"strings"
)

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

func isNoScriptError(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "NOSCRIPT ")
}
