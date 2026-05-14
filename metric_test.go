package redisson

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// TestNewMetricsSet 验证构造的 metric 集所有字段都被初始化。
func TestNewMetricsSet(t *testing.T) {
	m := newMetricsSet()
	if m.timing == nil ||
		m.err == nil ||
		m.hits == nil ||
		m.miss == nil ||
		m.delayPollError == nil ||
		m.delayReclaimError == nil ||
		m.delayReclaimCount == nil {
		t.Errorf("newMetricsSet returned partially initialized set: %+v", m)
	}
}

// TestMetricsSet_Register 把所有 metric 注册到 hook。
func TestMetricsSet_Register(t *testing.T) {
	m := newMetricsSet()
	registered := make([]prometheus.Collector, 0, 7)
	m.register(func(c prometheus.Collector) {
		registered = append(registered, c)
	})
	// 7 个 metric: err/hits/miss/delayPollError/delayReclaimError/delayReclaimCount/timing
	if len(registered) != 7 {
		t.Errorf("registered %d, want 7", len(registered))
	}
}

// TestMetricsSet_Register_Nil 安全 no-op 路径。
func TestMetricsSet_Register_Nil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("register(nil) panicked: %v", r)
		}
	}()
	m := newMetricsSet()
	m.register(nil) // 应直接返回，不触发 nil deref
}

// TestMetricsSet_Register_NilReceiver nil 接收者路径。
func TestMetricsSet_Register_NilReceiver(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("nil set register panicked: %v", r)
		}
	}()
	var m *metricsSet
	m.register(func(_ prometheus.Collector) {})
}

// TestRegisterMetric_OnceProtected 重复调 registerMetric 只生效一次。
func TestRegisterMetric_OnceProtected(t *testing.T) {
	calls := 0
	registerMetric(func(_ prometheus.Collector) { calls++ })
	first := calls
	registerMetric(func(_ prometheus.Collector) { calls++ })
	if calls != first {
		t.Errorf("registerMetric called twice: first=%d second=%d", first, calls)
	}
}
