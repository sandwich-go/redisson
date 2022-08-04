package redisson

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace   = "go_redis_pool_stats"
	subsystem   = "connections"
	defaultName = "default_redis"
)

type RegisterCollectorFunc func(prometheus.Collector)

func (c *client) RegisterCollector(rc RegisterCollectorFunc) {
	c.once.Do(func() {
		if rc != nil {
			c.handler.setRegisterCollector(rc)
			rc(newCollector(c))
		}
	})
}

type statsCollector struct {
	c *client
	// number of times free connection was found in the pool
	hitsDesc *prometheus.Desc
	// number of times free connection was NOT found in the pool
	missesDesc *prometheus.Desc
	//  number of times a wait timeout occurred
	timeoutsDesc *prometheus.Desc
	// number of total connections in the pool
	totalConnsDesc *prometheus.Desc
	// number of idle connections in the pool
	idleConnsDesc *prometheus.Desc
	// number of stale connections removed from the pool
	staleConnsDesc *prometheus.Desc
}

func newCollector(c *client) prometheus.Collector {
	labels := prometheus.Labels{"name": defaultName}
	return &statsCollector{
		c: c,
		hitsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "hits"),
			"number of times free connection was found in the pool.",
			nil,
			labels,
		),
		missesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "misses"),
			"number of times free connection was NOT found in the pool.",
			nil,
			labels,
		),
		timeoutsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "timeouts"),
			"number of times a wait timeout occurred.",
			nil,
			labels,
		),
		totalConnsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total_conns"),
			"number of total connections in the pool.",
			nil,
			labels,
		),
		idleConnsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle_conns"),
			"number of idle connections in the pool.",
			nil,
			labels,
		),
		staleConnsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "stale_conns"),
			"number of stale connections removed from the pool.",
			nil,
			labels,
		),
	}
}

// Describe implements the prometheus.Collector interface.
func (c statsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hitsDesc
	ch <- c.missesDesc
	ch <- c.timeoutsDesc
	ch <- c.totalConnsDesc
	ch <- c.idleConnsDesc
	ch <- c.staleConnsDesc
}

// Collect implements the prometheus.Collector interface.
func (c statsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.c.PoolStats()
	ch <- prometheus.MustNewConstMetric(
		c.hitsDesc,
		prometheus.GaugeValue,
		float64(stats.Hits),
	)
	ch <- prometheus.MustNewConstMetric(
		c.missesDesc,
		prometheus.GaugeValue,
		float64(stats.Misses),
	)
	ch <- prometheus.MustNewConstMetric(
		c.timeoutsDesc,
		prometheus.GaugeValue,
		float64(stats.Timeouts),
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalConnsDesc,
		prometheus.GaugeValue,
		float64(stats.TotalConns),
	)
	ch <- prometheus.MustNewConstMetric(
		c.idleConnsDesc,
		prometheus.GaugeValue,
		float64(stats.IdleConns),
	)
	ch <- prometheus.MustNewConstMetric(
		c.staleConnsDesc,
		prometheus.GaugeValue,
		float64(stats.StaleConns),
	)
}
