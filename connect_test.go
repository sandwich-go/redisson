package redisson

import (
	"errors"
	"fmt"
	"testing"

	"github.com/redis/rueidis"
)

// 这些测试仅验证 reconnectErrors 中各识别函数的分类逻辑，
// 不实际进行网络连接，因而没有 //go:build integration 标签。

// fakeClientForReconnect 构造一个仅 v 字段被设置的 *client，
// 使 reconnectErrors[i](c, err) 可以被单独执行。
func fakeClientForReconnect(opts ...ConfOption) *client {
	v := NewConf(opts...)
	return &client{v: v}
}

// TestReconnectErrors_CacheNotSupported 验证 ErrNoCache 命中后关闭 EnableCache。
func TestReconnectErrors_CacheNotSupported(t *testing.T) {
	c := fakeClientForReconnect(WithEnableCache(true))
	wrapped := fmt.Errorf("connection failed: %w", rueidis.ErrNoCache)

	if !reconnectErrors[0](c, wrapped) {
		t.Fatal("ErrNoCache should be recognized by reconnectErrors[0]")
	}
	if c.v.GetEnableCache() {
		t.Fatal("EnableCache should be flipped to false after ErrNoCache")
	}
}

// TestReconnectErrors_CacheAlreadyDisabled 验证已关闭 cache 时不再重复处理。
func TestReconnectErrors_CacheAlreadyDisabled(t *testing.T) {
	c := fakeClientForReconnect(WithEnableCache(false))
	if reconnectErrors[0](c, rueidis.ErrNoCache) {
		t.Fatal("with cache already disabled, reconnectErrors[0] should not match")
	}
}

// TestReconnectErrors_FallbackToRESP2 验证 cluster info 元素错误触发 RESP2 fallback。
func TestReconnectErrors_FallbackToRESP2(t *testing.T) {
	c := fakeClientForReconnect()
	err := errors.New("got 1 elements in cluster info address, expected 2 or 3")
	if !reconnectErrors[1](c, err) {
		t.Fatal("cluster-info-elements error should match")
	}
	if !c.v.GetAlwaysRESP2() {
		t.Fatal("AlwaysRESP2 should be set to true")
	}
}

// TestReconnectErrors_FallbackHello 验证 unsupported HELLO 错误触发 RESP2 fallback。
func TestReconnectErrors_FallbackHello(t *testing.T) {
	c := fakeClientForReconnect()
	err := errors.New("unsupported command `hello`")
	if !reconnectErrors[1](c, err) {
		t.Fatal("unsupported HELLO should match")
	}
	if !c.v.GetAlwaysRESP2() {
		t.Fatal("AlwaysRESP2 should be set to true")
	}
}

// TestReconnectErrors_AlreadyRESP2 验证已经 RESP2 时不再 match。
func TestReconnectErrors_AlreadyRESP2(t *testing.T) {
	c := fakeClientForReconnect(WithAlwaysRESP2(true))
	err := errors.New("got 1 elements in cluster info address, expected 2 or 3")
	if reconnectErrors[1](c, err) {
		t.Fatal("already-RESP2 client should not match RESP2 fallback")
	}
}

// TestReconnectErrors_ForceSingleClient 验证 slot has no redis node 触发 ForceSingleClient。
func TestReconnectErrors_ForceSingleClient(t *testing.T) {
	c := fakeClientForReconnect()
	err := errors.New("the slot has no redis node")
	if !reconnectErrors[2](c, err) {
		t.Fatal("slot-no-node error should match")
	}
	if !c.v.GetForceSingleClient() {
		t.Fatal("ForceSingleClient should be set to true")
	}
}

// TestReconnectErrors_NoMatch 验证一般错误不命中任何分支。
func TestReconnectErrors_NoMatch(t *testing.T) {
	c := fakeClientForReconnect()
	err := errors.New("connection refused")
	for i, f := range reconnectErrors {
		if f(c, err) {
			t.Fatalf("reconnectErrors[%d] should not match a generic connection-refused error", i)
		}
	}
}

// TestReconnectWhenError_NilPassthrough 验证 nil error 直接返回 nil 且 recognized=false。
func TestReconnectWhenError_NilPassthrough(t *testing.T) {
	c := fakeClientForReconnect()
	err, recognized := c.reconnectWhenError(nil)
	if err != nil {
		t.Fatalf("nil err should pass through, got %v", err)
	}
	if recognized {
		t.Fatal("nil err should not be recognized as reconnect-able")
	}
}
