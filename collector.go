package redisson

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "redis"
	// collectScrapeTimeout Prometheus 采集 delay queue 长度的最大耗时；
	// 控制单次 scrape 时间，避免被 Redis 慢查询拖死。
	collectScrapeTimeout = 2 * time.Second
)

type RegisterCollectorFunc func(prometheus.Collector)

func (c *client) RegisterCollector(rc RegisterCollectorFunc) {
	c.once.Do(func() {
		if rc != nil {
			c.handler.setRegisterCollector(rc)
		}
		if c.v.GetEnableMonitor() {
			registerCollector(rc, c)
		}
	})
}

var colOnce sync.Once
var col = &collector{
	delayLengthDesc: prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "delay", "queue_length"),
		"length of delay queue.",
		[]string{"queue"},
		prometheus.Labels{},
	),
}

type collector struct {
	cs              sync.Map
	delayLengthDesc *prometheus.Desc
}

func registerCollector(rc RegisterCollectorFunc, c *client) {
	col.cs.Store(c, struct{}{})
	colOnce.Do(func() {
		rc(col)
	})
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.delayLengthDesc
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.cs.Range(func(key, value any) bool {
		cli := key.(*client)
		if cli == nil {
			return true
		}
		cli.delayQueues.Range(func(key, value any) bool {
			ctx, cancel := context.WithTimeout(context.Background(), collectScrapeTimeout)
			l, _ := value.(*delayQueue).Length(ctx)
			cancel()
			ch <- prometheus.MustNewConstMetric(
				c.delayLengthDesc,
				prometheus.GaugeValue,
				float64(l),
				key.(string),
			)
			return true
		})
		return true
	})
}
