package redisson

import (
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// TestCollector_Describe 默认包级 col 的 Describe 应至少推一个 desc。
func TestCollector_Describe(t *testing.T) {
	c := &collector{
		delayLengthDesc: prometheus.NewDesc(
			"redis_delay_queue_length_test",
			"len",
			[]string{"queue"},
			prometheus.Labels{},
		),
	}
	ch := make(chan *prometheus.Desc, 1)
	c.Describe(ch)
	close(ch)
	cnt := 0
	for range ch {
		cnt++
	}
	if cnt != 1 {
		t.Errorf("Describe sent %d, want 1", cnt)
	}
}

// TestCollector_Collect_NoClients 没注册 client 时 Collect 不发出任何 metric。
func TestCollector_Collect_NoClients(t *testing.T) {
	c := &collector{
		delayLengthDesc: prometheus.NewDesc(
			"redis_delay_queue_length_empty_test",
			"len",
			[]string{"queue"},
			prometheus.Labels{},
		),
	}
	ch := make(chan prometheus.Metric, 8)
	c.Collect(ch)
	close(ch)
	cnt := 0
	for range ch {
		cnt++
	}
	if cnt != 0 {
		t.Errorf("Collect with no clients emitted %d, want 0", cnt)
	}
}

// TestCollector_Collect_NilClientSkipped cs 中 nil key 被跳过，不 panic。
func TestCollector_Collect_NilClientSkipped(t *testing.T) {
	c := &collector{
		delayLengthDesc: prometheus.NewDesc(
			"redis_delay_queue_length_nilcli_test",
			"len",
			[]string{"queue"},
			prometheus.Labels{},
		),
	}
	// 注入一个 nil *client
	c.cs.Store((*client)(nil), struct{}{})
	ch := make(chan prometheus.Metric, 8)
	c.Collect(ch)
	close(ch)
	for range ch {
		// 不期望任何 metric
		t.Errorf("nil client should be skipped")
	}
}

// TestColPackageOnceInitialized 验证包级 col 字段已被 init。
func TestColPackageOnceInitialized(t *testing.T) {
	if col == nil || col.delayLengthDesc == nil {
		t.Errorf("package-level col not initialized")
	}
}

// TestCollector_ClientRemovedOnClose 验证 client.Close 从全局 col.cs 摘除自己,
// 否则长期运行(频繁 New/Close) 会泄漏 client root 与 handler 引用,
// 且 Prometheus scrape 仍会枚举到已 Close 实例。
func TestCollector_ClientRemovedOnClose(t *testing.T) {
	// 构造最小 client (不真正 connect),手动注册到 col.cs 后 Close 验证摘除。
	c := &client{v: NewConf(), handler: newBaseHandler(NewConf())}
	col.cs.Store(c, struct{}{})
	if _, ok := col.cs.Load(c); !ok {
		t.Fatalf("precondition failed: client should be in col.cs")
	}
	_ = c.Close()
	if _, ok := col.cs.Load(c); ok {
		t.Fatalf("client should be removed from col.cs after Close")
	}
}

// 防止编译器把 sync 当未用
var _ = sync.Once{}
