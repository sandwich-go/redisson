//go:build integration

package redisson

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 这一套 integration 测试在真实 Redis 上验证虚拟 Stream Hub。
// 需要 build tag `integration` 与可用的 Redis 实例。

func itXAdd(t *testing.T, c Cmdable, ctx context.Context, stream, value string) {
	t.Helper()
	r := c.XAdd(ctx, XAddArgs{Stream: stream, Values: map[string]any{"v": value}})
	if r.Err() != nil {
		t.Fatalf("xadd %s=%s: %v", stream, value, r.Err())
	}
}

func TestVirtualStreamIntegration_BasicReadFromStart(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualStreamHub()
	defer hub.Close()

	ctx := context.Background()
	stream := "vsint." + strconv.FormatInt(time.Now().UnixNano(), 10)
	for i := 0; i < 3; i++ {
		itXAdd(t, c, ctx, stream, "v"+strconv.Itoa(i))
	}
	v, err := hub.Subscribe(ctx, []string{stream}, "0")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer v.Close()

	got := 0
	deadline := time.After(3 * time.Second)
	for got < 3 {
		select {
		case <-v.Channel():
			got++
		case <-deadline:
			t.Fatalf("only got %d/3 messages", got)
		}
	}
}

func TestVirtualStreamIntegration_DefaultDollar(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualStreamHub()
	defer hub.Close()

	ctx := context.Background()
	stream := "vsint.$." + strconv.FormatInt(time.Now().UnixNano(), 10)
	itXAdd(t, c, ctx, stream, "old")
	v, err := hub.Subscribe(ctx, []string{stream})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer v.Close()
	time.Sleep(300 * time.Millisecond)
	itXAdd(t, c, ctx, stream, "new")
	select {
	case m := <-v.Channel():
		if got := m.Message.Values["v"].(string); got != "new" {
			t.Fatalf("got %q, want new (default $ should skip)", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("no message")
	}
}

func TestVirtualStreamIntegration_Fanout(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualStreamHub()
	defer hub.Close()

	ctx := context.Background()
	stream := "vsint.fan." + strconv.FormatInt(time.Now().UnixNano(), 10)
	v1, _ := hub.Subscribe(ctx, []string{stream}, "0")
	v2, _ := hub.Subscribe(ctx, []string{stream}, "0")
	defer v1.Close()
	defer v2.Close()
	time.Sleep(300 * time.Millisecond)
	itXAdd(t, c, ctx, stream, "hi")

	get := func(v VirtualStream) string {
		select {
		case m := <-v.Channel():
			return m.Message.Values["v"].(string)
		case <-time.After(3 * time.Second):
			return ""
		}
	}
	if g := get(v1); g != "hi" {
		t.Fatalf("v1 got %q", g)
	}
	if g := get(v2); g != "hi" {
		t.Fatalf("v2 got %q", g)
	}
}

func TestVirtualStreamIntegration_HubCloseCascades(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualStreamHub()

	ctx := context.Background()
	stream := "vsint.cascade." + strconv.FormatInt(time.Now().UnixNano(), 10)
	v, err := hub.Subscribe(ctx, []string{stream}, "0")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range v.Channel() {
		}
	}()
	time.Sleep(200 * time.Millisecond)
	hub.Close()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("Channel not closed after Hub.Close")
	}
}

func TestVirtualStreamIntegration_ManyStreamsSameSlot(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualStreamHub()
	defer hub.Close()

	ctx := context.Background()
	const N = 200
	subs := make([]VirtualStream, 0, N)
	for i := 0; i < N; i++ {
		stream := "{vsmany}.s." + strconv.Itoa(i)
		v, err := hub.Subscribe(ctx, []string{stream}, "0")
		if err != nil {
			t.Fatalf("subscribe %s: %v", stream, err)
		}
		subs = append(subs, v)
	}
	st := hub.Stats()
	if st.Streams != N {
		t.Fatalf("Streams=%d, want %d", st.Streams, N)
	}
	if st.Workers != 1 {
		t.Fatalf("Workers=%d, want 1 (hashtag share slot)", st.Workers)
	}
	for _, v := range subs {
		_ = v.Close()
	}
}
