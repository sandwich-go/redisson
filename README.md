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

* Golang >= 1.18

If you can't upgrade Golang to 1.18, install redisson/version/0.1.

## Base Library
- RESP2, using [go-redis/redis](https://github.com/go-redis/redis) library.
- RESP3, using [rueian/rueidis](https://github.com/rueian/rueidis) library.

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

> :warning: Development Mode, will ***Panic*** when check version failed.

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
[MSET]: multiple keys command with different key slots are not allowed
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
[CLUSTER FAILOVER]: redis command are not allowed 
```

## Monitor

Import Grafana dashboard id `16768`

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/sandwich-go/redisson"
)

var DefaultPrometheusRegistry = prometheus.NewRegistry()

c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3),
      redisson.WithEnableMonitor(true),
))
defer c.Close()

c.RegisterCollector(func(c prometheus.Collector) {
    DefaultPrometheusRegistry.Register(c)
})
```

![grafana_dashboard](https://github.com/sandwich-go/redisson/blob/version/1.0/grafana_dashboard.png)

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

## Benchmark
### Environment
- [go-redis/redis](https://github.com/go-redis/redis) v8.11.5
- [joomcode/redispipe](https://github.com/joomcode/redispipe) v0.9.4
- [mediocregopher/radix](https://github.com/mediocregopher/radix) v4.1.1
- [rueian/rueidis](https://github.com/rueian/rueidis) v0.0.74
- [sandwich-go/redisson](https://github.com/sandwich-go/redisson) v1.1.14

### Benchmarking Result
##### Single, Parallel mode, Get Command
```markdown
+---------------------------------------------------+-----------+-------+-------+-----------+
| Single Parallel(128) Get                          | iteration | ns/op | B/op  | allocs/op |
+===================================================+===========+=======+=======+===========+
| sandwich-go/redisson/RESP2:Val(64):Pool(100)      | 362365    | 6136  | 279   |        6  |
| sandwich-go/redisson/RESP2:Val(64):Pool(1000)     | 504202    | 4731  | 286   |        6  |
| sandwich-go/redisson/RESP2:Val(256):Pool(100)     | 362181    | 6334  | 487   |        6  |
| sandwich-go/redisson/RESP2:Val(256):Pool(1000)    | 481341    | 4946  | 495   |        6  |
| sandwich-go/redisson/RESP2:Val(1024):Pool(100)    | 332634    | 6822  | 1351  |        6  |
| sandwich-go/redisson/RESP2:Val(1024):Pool(1000)   | 451609    | 5299  | 1360  |        6  |
| sandwich-go/redisson/RESP3:Val(64):Pool(100)      | 1208716   | 1923  | 320   |        4  |
| sandwich-go/redisson/RESP3:Val(256):Pool(100)     | 1000000   | 2013  | 512   |        4  |
| sandwich-go/redisson/RESP3:Val(1024):Pool(100)    | 728786    | 2816  | 1281  |        4  |
| rueian/rueidis/rueidiscompat:Val(64):Pool(100)    | 1253146   | 1847  | 256   |        4  |
| rueian/rueidis/rueidiscompat:Val(256):Pool(100)   | 1000000   | 2034  | 448   |        4  |
| rueian/rueidis/rueidiscompat:Val(1024):Pool(100)  | 792254    | 2686  | 1217  |        4  |
| go-redis/redis/v8:Val(64):Pool(100)               | 369186    | 6098  | 279   |        6  |
| go-redis/redis/v8:Val(64):Pool(1000)              | 506796    | 4750  | 286   |        6  |
| go-redis/redis/v8:Val(256):Pool(100)              | 357454    | 6266  | 487   |        6  |
| go-redis/redis/v8:Val(256):Pool(1000)             | 486217    | 4919  | 495   |        6  |
| go-redis/redis/v8:Val(1024):Pool(100)             | 331382    | 6779  | 1351  |        6  |
| go-redis/redis/v8:Val(1024):Pool(1000)            | 452067    | 5307  | 1360  |        6  |
| mediocregopher/radix/v4:Val(64):Pool(100)         | 596540    | 4284  | 26    |        1  |
| mediocregopher/radix/v4:Val(64):Pool(1000)        | 589083    | 4902  | 54    |        1  |
| mediocregopher/radix/v4:Val(256):Pool(100)        | 576108    | 4384  | 27    |        1  |
| mediocregopher/radix/v4:Val(256):Pool(1000)       | 597157    | 4993  | 54    |        1  |
| mediocregopher/radix/v4:Val(1024):Pool(100)       | 573411    | 4539  | 27    |        1  |
| mediocregopher/radix/v4:Val(1024):Pool(1000)      | 559611    | 5062  | 56    |        1  |
| joomcode/redispipe:Val(64):Pool(100)              | 1109589   | 2137  | 168   |        5  |
| joomcode/redispipe:Val(256):Pool(100)             | 1000000   | 2170  | 377   |        5  |
| joomcode/redispipe:Val(1024):Pool(100)            | 958350    | 2442  | 1241  |        5  |
+---------------------------------------------------+-----------+-------+-------+-----------+  
```

![BenchmarkSingleClientGetParallel](https://github.com/sandwich-go/go-redis-client-benchmark/blob/master/BenchmarkSingleClientGetParallel.png)

##### Cluster, Parallel mode, Get Command
```markdown
+---------------------------------------------------+-----------+-------+-------+-----------+ 
| Cluster Parallel(128) Get                         | iteration | ns/op | B/op  | allocs/op | 
+===================================================+===========+=======+=======+===========+ 
| sandwich-go/redisson/RESP2:Val(64):Pool(100)      | 361689    | 6246  | 279   |        6  |
| sandwich-go/redisson/RESP2:Val(64):Pool(1000)     | 494625    | 4819  | 286   |        6  |
| sandwich-go/redisson/RESP2:Val(256):Pool(100)     | 353413    | 6439  | 487   |        6  |
| sandwich-go/redisson/RESP2:Val(256):Pool(1000)    | 478305    | 5035  | 494   |        6  |
| sandwich-go/redisson/RESP2:Val(1024):Pool(100)    | 324940    | 6992  | 1351  |        6  |
| sandwich-go/redisson/RESP2:Val(1024):Pool(1000)   | 441291    | 5472  | 1360  |        6  |
| sandwich-go/redisson/RESP3:Val(64):Pool(100)      | 1036126   | 2275  | 320   |        4  |
| sandwich-go/redisson/RESP3:Val(256):Pool(100)     | 1008175   | 2420  | 513   |        4  |
| sandwich-go/redisson/RESP3:Val(1024):Pool(100)    | 766168    | 2906  | 1282  |        4  |
| rueian/rueidis/rueidiscompat:Val(64):Pool(100)    | 946216    | 2266  | 256   |        4  |
| rueian/rueidis/rueidiscompat:Val(256):Pool(100)   | 924811    | 2292  | 448   |        4  |
| rueian/rueidis/rueidiscompat:Val(1024):Pool(100)  | 856582    | 2802  | 1218  |        4  |
| go-redis/redis/v8:Val(64):Pool(100)               | 351850    | 6251  | 279   |        6  |
| go-redis/redis/v8:Val(64):Pool(1000)              | 489259    | 4821  | 286   |        6  |
| go-redis/redis/v8:Val(256):Pool(100)              | 356703    | 6385  | 487   |        6  |
| go-redis/redis/v8:Val(256):Pool(1000)             | 478236    | 5012  | 494   |        6  |
| go-redis/redis/v8:Val(1024):Pool(100)             | 333362    | 6972  | 1351  |        6  |
| go-redis/redis/v8:Val(1024):Pool(1000)            | 443264    | 5386  | 1360  |        6  |
| mediocregopher/radix/v4:Val(64):Pool(100)         | 477573    | 4598  | 113   |        2  |
| mediocregopher/radix/v4:Val(64):Pool(1000)        | 386779    | 5431  | 114   |        2  |
| mediocregopher/radix/v4:Val(256):Pool(100)        | 459818    | 4737  | 113   |        2  |
| mediocregopher/radix/v4:Val(256):Pool(1000)       | 383200    | 5656  | 114   |        2  |
| mediocregopher/radix/v4:Val(1024):Pool(100)       | 451070    | 4911  | 114   |        2  |
| mediocregopher/radix/v4:Val(1024):Pool(1000)      | 356745    | 5745  | 114   |        2  |
| joomcode/redispipe:Val(64):Pool(100)              | 1091751   | 2147  | 170   |        5  |
| joomcode/redispipe:Val(256):Pool(100)             | 1088572   | 2298  | 379   |        5  |
| joomcode/redispipe:Val(1024):Pool(100)            | 800530    | 2548  | 1246  |        5  |
+---------------------------------------------------+-----------+-------+-------+-----------+ 
```

![BenchmarkClusterClientGetParallel](https://github.com/sandwich-go/go-redis-client-benchmark/blob/master/BenchmarkClusterClientGetParallel.png)

See [Benchmark Detail Result](https://github.com/sandwich-go/go-redis-client-benchmark)


* [Opt-in client side caching](https://redis.io/docs/manual/client-side-caching/)
* [RESP](https://redis.io/docs/reference/protocol-spec/)
* [RESP2](https://github.com/redis/redis-specifications/blob/master/protocol/RESP2.md)
* [RESP3](https://github.com/antirez/RESP3/blob/master/spec.md)