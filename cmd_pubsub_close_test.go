package redisson

import (
	"context"
	"testing"
	"time"

	"github.com/redis/rueidis"
)

// TestPubSubForwardMessage_NonBlockingAfterClose 回归 Bug #4：
// pubSub.forwardMessage 在 ctx 已 cancel（即 Close 后）的情况下不应阻塞，
// 必须立即丢弃消息（避免 UnboundedChan process goroutine 退出后
// In 缓冲打满 → rueidis 回调永久阻塞 → 连接池泄漏）。
func TestPubSubForwardMessage_NonBlockingAfterClose(t *testing.T) {
	conf := NewConf()
	c := &client{v: conf, handler: newBaseHandler(conf)}
	p := newPubSub(context.Background(), c, c.handler).(*pubSub)
	// 立即 Close，模拟 UnboundedChan process goroutine 已退出 + In 缓冲已经满。
	_ = p.Close()

	// 直接调用 forwardMessage，必须立即返回（不依赖 In 是否被消费）。
	done := make(chan struct{})
	go func() {
		// 即便 isClosed 检查通过到了 select，ctx.Done 也会立即放行。
		p.forwardMessage("test", []string{"ch"}, rueidis.PubSubMessage{Channel: "ch", Message: "x"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("forwardMessage blocked after Close — goroutine leak risk")
	}
}

// TestPubSubForwardMessage_NonBlockingWhenBufferFullAndClosed 模拟
// "msgCh.In 缓冲打满 + p.ctx 已 cancel" 双重条件下，forwardMessage 必须不阻塞。
// 这是 Bug #4 最坏情况：Close 触发 process goroutine 退出，回调累积写到无人消费的 In。
func TestPubSubForwardMessage_NonBlockingWhenBufferFullAndClosed(t *testing.T) {
	// 这里跳过 isClosed 早返回路径，构造 ctx 已 cancel 但 isClosed=false 的极端竞态：
	// 通过手动重置 closed 字段实现。
	conf := NewConf()
	c := &client{v: conf, handler: newBaseHandler(conf)}
	p := newPubSub(context.Background(), c, c.handler).(*pubSub)
	// 先 cancel ctx（process goroutine 会退出），但保持 closed 为 0
	p.cancel()
	// 给 process goroutine 时间退出
	time.Sleep(50 * time.Millisecond)
	// 灌满 In 缓冲（默认 1024），让 send 必然阻塞
	for i := 0; i < conf.GetPubSubChanSize(); i++ {
		select {
		case p.msgCh.In <- rueidis.PubSubMessage{Channel: "x", Message: "y"}:
		default:
			// 缓冲满即停（process 可能还残留消费一个）
			i = conf.GetPubSubChanSize()
		}
	}
	// 再来一条：旧实现会永久阻塞；新实现因为 ctx.Done 会立即返回。
	done := make(chan struct{})
	go func() {
		p.forwardMessage("test", []string{"ch"}, rueidis.PubSubMessage{Channel: "ch", Message: "x"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("forwardMessage blocked when buffer full and ctx canceled — Bug #4 not fixed")
	}
}
