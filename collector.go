package redisson

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
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
			rc(newCollector(c))
		}
	})
}

type collector struct {
	c               *client
	delayLengthDesc *prometheus.Desc
}

func newCollector(c *client) prometheus.Collector {
	return &collector{
		c: c,
		delayLengthDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "delay", "queue_length"),
			"length of delay queue.",
			[]string{"queue"},
			prometheus.Labels{},
		),
	}
}

func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.delayLengthDesc
}

func (c collector) Collect(ch chan<- prometheus.Metric) {
	c.c.delayQueues.Range(func(key, value any) bool {
		l, _ := value.(*delayQueue).Length(context.Background())
		ch <- prometheus.MustNewConstMetric(
			c.delayLengthDesc,
			prometheus.GaugeValue,
			float64(l),
			key.(string),
		)
		return true
	})
}
