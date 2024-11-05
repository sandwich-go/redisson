package redisson

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
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
	metricOnceMap                                                          sync.Map
	metric                                                                 *prometheus.SummaryVec
	errMetric, hitsMetric, missMetric                                      *prometheus.CounterVec
	delayPollErrorMetric, delayReclaimErrorMetric, delayReclaimCountMetric *prometheus.CounterVec
)

var (
	labelKeys      = []string{"command", "s_command"}
	queueLabelKeys = []string{"queue"}
)

func init() {
	errMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: errorMetricName,
	}, labelKeys)
	hitsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: hitsMetricName,
	}, labelKeys)
	missMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: missMetricName,
	}, labelKeys)
	delayPollErrorMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: delayPollErrorMetricName,
	}, queueLabelKeys)
	delayReclaimErrorMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: delayReclaimErrorMetricName,
	}, queueLabelKeys)
	delayReclaimCountMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: delayReclaimCountMetricName,
	}, queueLabelKeys)
	metric = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       timingMetricName,
		Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.02, 0.99: 0.001, 1: 0},
		MaxAge:     time.Minute,
	}, labelKeys)
}

func registerMetric(rc RegisterCollectorFunc) {
	m, _ := metricOnceMap.LoadOrStore(rc, &sync.Once{})
	m.(*sync.Once).Do(func() {
		rc(errMetric)
		rc(hitsMetric)
		rc(missMetric)
		rc(delayPollErrorMetric)
		rc(delayReclaimErrorMetric)
		rc(delayReclaimCountMetric)
		rc(metric)
	})
}
