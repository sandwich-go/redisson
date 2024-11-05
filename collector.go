package redisson

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

const (
	namespace = "redis"
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

var colOnceMap sync.Map
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
	m, _ := colOnceMap.LoadOrStore(rc, &sync.Once{})
	m.(*sync.Once).Do(func() {
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
			l, _ := value.(*delayQueue).Length(context.Background())
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
