package redisson

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	timingMetricName            = "redis_exec_timing"
	errorMetricName             = "redis_exec_error"
	hitsMetricName              = "redis_cache_hits"
	missMetricName              = "redis_cache_miss"
	delayPollErrorMetricName    = "redis_delay_poll_error"
	delayReclaimErrorMetricName = "redis_delay_reclaim_error"
	delayReclaimCountMetricName = "redis_delay_reclaim"
)

var (
	labelKeys      = []string{"command", "s_command"}
	queueLabelKeys = []string{"queue"}
)

// metricsSet 是一组 prometheus 指标的实例集合。
// 每个 client 持有一份实例（构造期生成），不再使用全局 init() 单例，
// 这样多 client 实例就能各自独立计数，避免相互污染。
//
// 兼容性说明：包级变量 errMetric 等仍保留并被 init() 初始化为同一份实例，
// 由 RegisterCollector 默认路径使用，保持外部行为不变。
type metricsSet struct {
	timing            *prometheus.SummaryVec
	err               *prometheus.CounterVec
	hits              *prometheus.CounterVec
	miss              *prometheus.CounterVec
	delayPollError    *prometheus.CounterVec
	delayReclaimError *prometheus.CounterVec
	delayReclaimCount *prometheus.CounterVec
}

// newMetricsSet 创建一组新的 metric 实例。
func newMetricsSet() *metricsSet {
	return &metricsSet{
		err: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: errorMetricName,
		}, labelKeys),
		hits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: hitsMetricName,
		}, labelKeys),
		miss: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: missMetricName,
		}, labelKeys),
		delayPollError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: delayPollErrorMetricName,
		}, queueLabelKeys),
		delayReclaimError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: delayReclaimErrorMetricName,
		}, queueLabelKeys),
		delayReclaimCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: delayReclaimCountMetricName,
		}, queueLabelKeys),
		timing: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       timingMetricName,
			Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.02, 0.99: 0.001, 1: 0},
			MaxAge:     time.Minute,
		}, labelKeys),
	}
}

// register 把 set 中所有指标注册到给定的 collector hook。
func (m *metricsSet) register(rc RegisterCollectorFunc) {
	if m == nil || rc == nil {
		return
	}
	rc(m.err)
	rc(m.hits)
	rc(m.miss)
	rc(m.delayPollError)
	rc(m.delayReclaimError)
	rc(m.delayReclaimCount)
	rc(m.timing)
}

// 全局兼容层 ============================================================
//
// 旧版本通过 init() 创建一组全局 *prometheus.CounterVec 等指标，
// 由 RegisterCollector + registerMetric 流程注册。为保持向后兼容，
// 这里继续暴露同名包级变量，并在 init() 时复用 newMetricsSet 的实例。
var (
	defaultMetrics = newMetricsSet()

	// 旧版包级别名（保留以兼容外部可能直接读取的代码；新代码使用 metricsSet 实例）。
	//
	//nolint:unused // 公共包级变量，外部用户可能引用
	metric = defaultMetrics.timing
	//nolint:unused
	errMetric = defaultMetrics.err
	//nolint:unused
	hitsMetric = defaultMetrics.hits
	//nolint:unused
	missMetric = defaultMetrics.miss
	//nolint:unused
	delayPollErrorMetric = defaultMetrics.delayPollError
	//nolint:unused
	delayReclaimErrorMetric = defaultMetrics.delayReclaimError
	//nolint:unused
	delayReclaimCountMetric = defaultMetrics.delayReclaimCount

	metricOnce sync.Once
)

// registerMetric 旧版默认路径：注册全局指标实例（一次性）。
//
// 新代码可以直接调用 metricsSet.register。
func registerMetric(rc RegisterCollectorFunc) {
	metricOnce.Do(func() {
		defaultMetrics.register(rc)
	})
}
