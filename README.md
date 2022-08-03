# redisson

A Type-safe Golang Redis RESP2 client.

## Features

* Check Redis commands version in development mode.
* Check Redis deprecated commands in development mode.
* Check Redis slot when use multiple keys in development mode.
* Check forbid Redis commands in development mode.
* Monitoring cost of Redis commands.
* Monitoring status of connections.
* Support Redis RESP2.

## Requirement

* Currently, only supports Redis < 7.x
* Golang >= 1.6

## Links
* [English](https://github.com/sandwich-go/redisson/blob/version/1.0/README.md)
* [中文文档](https://github.com/sandwich-go/redisson/blob/version/1.0/README_CN.md)

## Getting Started

```golang
package main

import (
	"context"
	"github.com/sandwich-go/redisson"
)

func main() {
	c := redisson.MustNewClient(redisson.NewConf(
	      redisson.WithResp(redisson.RESP2), 
	      redisson.WithDevelopment(false), 
	))
	defer c.Close()

	ctx := context.Background()

	// SET key val NX
	_ = c.SetNX(ctx, "key", "val", 0).Err()
	// GET key
	_ = c.Get(ctx, "key").Val()
}
```

## Check
### Check version
if Redis < 6.0
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP2), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.Set(ctx, "key", "10", -1)
```
Output:
```go
Line 34: - redis 'SET KEEPTTL' are not supported in version "5.0.0", available since 6.0.0
```

### Check deprecated
if Redis >= 4.0
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP2), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.HMSet(ctx, "key", "10")
```
Output:
```go
As of Redis version 4.0.0, this command is regarded as deprecated.
It can be replaced by HSET with multiple field-value pairs when migrating or writing new code.
```

### Check slot for multiple keys
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP2), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.MSet(ctx, "key1", "10", "key2", "20")
```
Output:
```go
Line 34: - multi key command with different key slots are not allowed 
```

### Check forbid
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP2), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.ClusterFailover(ctx)
```
Output:
```go
Line 34: - command 'CLUSTER FAILOVER' not allowed 
```

## Monitor

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/sandwich-go/redisson"
)

var DefaultPrometheusRegistry = prometheus.NewRegistry()

c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3),
      redisson.WithDevelopment(true),
))
defer c.Close()

c.RegisterCollector(func(c prometheus.Collector) {
    DefaultPrometheusRegistry.Register(c)
})
```

* [Opt-in client side caching](https://redis.io/docs/manual/client-side-caching/)
* [RESP](https://redis.io/docs/reference/protocol-spec/)
* [RESP2](https://github.com/redis/redis-specifications/blob/master/protocol/RESP2.md)
* [RESP3](https://github.com/antirez/RESP3/blob/master/spec.md)