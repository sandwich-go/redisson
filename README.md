# redisson

A Type-safe Golang Redis RESP2/RESP3 client.

## Features

* Check Redis commands version in development mode.
* Check Redis deprecated commands in development mode.
* Check Redis slot when use multiple keys in development mode.
* Check forbid Redis commands in development mode.
* Monitoring cost of Redis commands.
* Monitoring status of connections.
* Monitoring hits/miss of Redis RESP3 client side caching.
* Support Redis RESP2/RESP3.
* Opt-in client side caching.
* Auto pipeline for non-blocking Redis RESP3 commands.
* Connection pooling for blocking Redis RESP3 commands.

## Requirement

* Currently, only supports Redis < 7.x
* Golang >= 1.18

If you can't upgrade Golang to 1.18, install redisson/version/0.1.

## Links
* [English](https://github.com/sandwich-go/redisson/blob/master/README.md)
* [中文文档](https://github.com/sandwich-go/redisson/blob/master/README_CN.md)

## Getting Started

```golang
package main

import (
	"context"
	"github.com/sandwich-go/redisson"
)

func main() {
	c := redisson.MustNewClient(redisson.NewConf(
	      redisson.WithResp(redisson.RESP3), 
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
If Redis < 6.0
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.Set(ctx, "key", "10", -1)
```
Output:
```text
[SET KEEPTTL]: redis command are not supported in version "5.0.0", available since 6.0.0
```

### Check deprecated
If Redis >= 4.0
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.HMSet(ctx, "key", "10")
```
Output:
```text
[HMSET]: As of Redis version 4.0.0, this command is regarded as deprecated.
It can be replaced by HSET with multiple field-value pairs when migrating or writing new code.
```

### Check slot for multiple keys
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.MSet(ctx, "key1", "10", "key2", "20")
```
Output:
```text
[MSET]: multiple keys command with different key slots are not allowed .
```

### Check forbid
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.ClusterFailover(ctx)
```
Output:
```text
[CLUSTER FAILOVER]: redis command not allowed 
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


## Auto Pipeline

All non-blocking commands sending to a single Redis node are automatically pipelined through one tcp connection,
which reduces the overall round trips and system calls, and gets higher throughput.

Notice: Only supports when use Redis RESP3 client.


## Client Side Caching

The Opt-In mode of server-assisted client side caching is always enabled.

```golang
c.Cache(time.Minute).Get(ctx, "key").Val()
```

An explicit client side TTL is required because Redis server may not send invalidation message in time when
a key is expired on the server. Please follow [#6833](https://github.com/redis/redis/issues/6833) and [#6867](https://github.com/redis/redis/issues/6867)

Although an explicit client side TTL is required, the `Cache()` still sends a `PTTL` command to server and make sure that
the client side TTL is not longer than the TTL on server side.

Notice: Only supports when use Redis RESP3 client.


* [Opt-in client side caching](https://redis.io/docs/manual/client-side-caching/)
* [RESP](https://redis.io/docs/reference/protocol-spec/)
* [RESP2](https://github.com/redis/redis-specifications/blob/master/protocol/RESP2.md)
* [RESP3](https://github.com/antirez/RESP3/blob/master/spec.md)