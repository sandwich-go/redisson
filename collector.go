package redisson

import "github.com/prometheus/client_golang/prometheus"

type RegisterCollectorFunc func(prometheus.Collector)

func (c *client) RegisterCollector(rc RegisterCollectorFunc) {
	c.once.Do(func() {
		if rc != nil {
			c.handler.setRegisterCollector(rc)
		}
	})
}
