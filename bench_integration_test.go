//go:build integration

package redisson

import (
	"context"
	"strconv"
	"testing"
)

// BenchmarkSet_Pipeline 测量串行 SET 吞吐（集成层）。
// 用法：go test -tags integration -bench BenchmarkSet -benchmem ./...
func BenchmarkSet(b *testing.B) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	b.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()
	c.FlushAll(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := "bench:set:" + strconv.Itoa(i)
		if err := c.Set(ctx, k, "v", 0).Err(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGet 测量 GET 吞吐。
func BenchmarkGet(b *testing.B) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	b.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()
	c.FlushAll(ctx)
	const key = "bench:get:hot"
	if err := c.Set(ctx, key, "hello-world", 0).Err(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Get(ctx, key).Result(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGet_Cached 测量启用客户端缓存后的 GET 吞吐（应明显优于 BenchmarkGet）。
func BenchmarkGet_Cached(b *testing.B) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithEnableCache(true)))
	b.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()
	c.FlushAll(ctx)
	const key = "bench:get:cached"
	if err := c.Set(ctx, key, "hello-world", 0).Err(); err != nil {
		b.Fatal(err)
	}
	cached := c.Cache(60_000_000_000) // 60s

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := cached.Get(ctx, key).Result(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPipeline_Set 测量管道吞吐（每批 100 个 SET）。
func BenchmarkPipeline_Set(b *testing.B) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	b.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()
	c.FlushAll(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := c.Pipeline()
		for j := 0; j < 100; j++ {
			CommandSet.P(p).Cmd("bench:pipe:"+strconv.Itoa(i)+":"+strconv.Itoa(j), "v", 0)
		}
		if _, err := p.Exec(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
