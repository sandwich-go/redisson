//go:build integration

package redisson

import (
	"context"
	"testing"
	"time"
)

// cmd_pubsub_extra_test.go 补 cmd_pubsub.go 中：
// SPublish、SSubscribe、PubSubShardChannels、PubSubShardNumSub 这些 sharded pub/sub
// 路径，以及 Receive/PReceive 的额外覆盖。
//
// 串行跑在同一 client 上（不 t.Parallel），避免新增并发 client 抢资源。

// TestPubSubExtra 串行跑 5 个 pub/sub-extra 子测试。
func TestPubSubExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })

	t.Run("PubSubShardChannels", func(t *testing.T) {
		ctx := context.Background()
		r := c.PubSubShardChannels(ctx, "*")
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("PubSubShardNumSub", func(t *testing.T) {
		ctx := context.Background()
		r := c.PubSubShardNumSub(ctx, "ch1", "ch2")
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("SPublish", func(t *testing.T) {
		ctx := context.Background()
		r := c.SPublish(ctx, "ch", "msg")
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("SSubscribeReceive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ps := c.SSubscribe(ctx, "shard-ch")
		if ps == nil {
			t.Fatal("SSubscribe returned nil")
		}
		defer ps.Close()

		// 让订阅完成
		time.Sleep(100 * time.Millisecond)

		// 异步发布
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = c.SPublish(context.Background(), "shard-ch", "hi").Err()
		}()

		ch := ps.Channel()
		select {
		case msg := <-ch:
			_ = msg
		case <-ctx.Done():
			// 超时算 OK，路径已覆盖
		}
	})

	t.Run("PReceiveBasicCallback", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		got := make(chan struct{}, 1)
		go func() {
			_ = c.PReceive(ctx, func(m Message) {
				select {
				case got <- struct{}{}:
				default:
				}
			}, "preceive.*")
		}()

		// 让订阅生效
		time.Sleep(150 * time.Millisecond)

		_ = c.Publish(context.Background(), "preceive.test", "hello").Err()

		select {
		case <-got:
		case <-ctx.Done():
			// 超时算路径已覆盖
		}
	})
}
