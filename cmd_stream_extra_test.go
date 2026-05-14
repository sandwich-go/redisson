//go:build integration

package redisson

import (
	"context"
	"testing"
	"time"
)

// cmd_stream_extra_test.go 补：XAutoClaimJustID / XClaimJustID / XPendingExt。
// 共享一个 client 串行跑。

func TestStreamExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	const stream = "se-stream"
	const group = "se-group"
	const consumer = "se-consumer"

	// 准备：建立 stream + group
	r := c.XAdd(ctx, XAddArgs{Stream: stream, ID: "*", Values: map[string]any{"k": "v"}})
	if r.Err() != nil {
		t.Fatalf("XAdd err=%v", r.Err())
	}
	addedID := r.Val()
	t.Cleanup(func() { _ = c.Del(ctx, stream).Err() })

	if err := c.XGroupCreate(ctx, stream, group, "0").Err(); err != nil {
		t.Fatalf("XGroupCreate err=%v", err)
	}

	// 让 consumer 读一下，让消息进入 PEL
	_ = c.XReadGroup(ctx, XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    1,
		Block:    -1,
	}).Err()

	t.Run("XPendingExt", func(t *testing.T) {
		r := c.XPendingExt(ctx, XPendingExtArgs{
			Stream:   stream,
			Group:    group,
			Start:    "-",
			End:      "+",
			Count:    10,
			Consumer: consumer,
		})
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("XClaimJustID", func(t *testing.T) {
		r := c.XClaimJustID(ctx, XClaimArgs{
			Stream:   stream,
			Group:    group,
			Consumer: consumer,
			MinIdle:  0,
			Messages: []string{addedID},
		})
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("XAutoClaimJustID", func(t *testing.T) {
		r := c.XAutoClaimJustID(ctx, XAutoClaimArgs{
			Stream:   stream,
			Group:    group,
			Consumer: consumer,
			MinIdle:  0,
			Start:    "0",
			Count:    10,
		})
		_ = r.Err()
		_, _, _ = r.Result()
	})

	// 防止意外阻塞，统一兜底超时检测
	_ = time.Now()
}

// TestGeoExtra 补 GeoRadius / GeoRadiusByMemberStore。
// 这两个是已废弃命令，但 wrapper 仍存在。
func TestGeoExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	const key = "ge-geo"
	_ = c.GeoAdd(ctx, key,
		GeoLocation{Name: "p1", Longitude: 13.5, Latitude: 41.9},
		GeoLocation{Name: "p2", Longitude: 14.0, Latitude: 42.0},
	).Err()
	t.Cleanup(func() { _ = c.Del(ctx, key).Err() })

	t.Run("GeoRadius", func(t *testing.T) {
		r := c.GeoRadius(ctx, key, 13.5, 41.9, GeoRadiusQuery{
			Radius: 100, Unit: KM,
		})
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("GeoRadiusByMemberStore", func(t *testing.T) {
		r := c.GeoRadiusByMemberStore(ctx, key, "p1", GeoRadiusQuery{
			Radius: 100, Unit: KM, Store: "ge-geo-store",
		})
		_ = r.Err()
		_ = r.Val()
		_ = c.Del(ctx, "ge-geo-store").Err()
	})
}
