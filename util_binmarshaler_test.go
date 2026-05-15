package redisson

import (
	"errors"
	"strings"
	"testing"
)

// errBinaryMarshaler 让 MarshalBinary 报错。
type errBinaryMarshaler struct{ msg string }

func (e errBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return nil, errors.New(e.msg)
}

type okBinaryMarshaler struct{ data []byte }

func (o okBinaryMarshaler) MarshalBinary() ([]byte, error) { return o.data, nil }

// TestStrBinaryMarshalerErrorPanics 验证 str() 在 BinaryMarshaler 返回 error 时
// panic ParameterError,而不是静默把 "&{...}" 之类乱码写入 Redis。
func TestStrBinaryMarshalerErrorPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when MarshalBinary returns error")
		}
		if !IsParameterError(r) {
			t.Fatalf("expected *ParameterError, got %T(%v)", r, r)
		}
		err := r.(error)
		if !strings.Contains(err.Error(), "MarshalBinary") {
			t.Fatalf("error should mention MarshalBinary, got: %s", err.Error())
		}
	}()
	_ = str(errBinaryMarshaler{msg: "boom"})
}

// TestStrBinaryMarshalerOKPath 验证正常路径仍工作。
func TestStrBinaryMarshalerOKPath(t *testing.T) {
	got := str(okBinaryMarshaler{data: []byte("hello")})
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}
