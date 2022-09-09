# redisson

一个类型安全的`Golang Redis`客户端，支持`RESP2/RESP3`协议

## 特征

* 开发模式下，检查`Redis`命令的版本要求
* 开发模式下，检查在当前版本已过期的`Redis`命令
* 开发模式下，检查多`Key`是否属于同一`Redis`槽
* 开发模式下，检查禁止使用的`Redis`命令
* 监控`Redis`命令耗时时间 
* 监控`Redis`连接状态
* 监控`Redis RESP3`客户端缓存命中状态
* 支持`RESP2/RESP3`协议
* 支持`Redis RESP3`客户端缓存
* `Redis RESP3`客户端命令自动进行`pipeline`
* `Redis RESP3`客户端自动管理阻塞的连接

## 要求

* Golang >= 1.18

如果不能升级`Golang`至`1.18`，请使用`redisson/version/0.1`版本。

## 基础库
- RESP2, 使用 [go-redis/redis](https://github.com/go-redis/redis).
- RESP3, 使用 [rueian/rueidis](https://github.com/rueian/rueidis).

## 链接
* [English](https://github.com/sandwich-go/redisson/blob/master/README.md)
* [中文文档](https://github.com/sandwich-go/redisson/blob/master/README_CN.md)

## 开始

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

## 检查
### 版本检查
如果 Redis < 6.0
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.Set(ctx, "key", "10", -1)
```
输出:
```text
[SET KEEPTTL]: redis command are not supported in version "5.0.0", available since 6.0.0
```

### 检查过期
如果 Redis >= 4.0
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.HMSet(ctx, "key", "10")
```
输出:
```text
[HMSET]: As of Redis version 4.0.0, this command is regarded as deprecated.
It can be replaced by HSET with multiple field-value pairs when migrating or writing new code.
```

### 检查槽位
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.MSet(ctx, "key1", "10", "key2", "20")
```
输出:
```text
[MSET]: multiple keys command with different key slots are not allowed
```

### 命令禁用
```go
c := redisson.MustNewClient(redisson.NewConf(
      redisson.WithResp(redisson.RESP3), 
      redisson.WithDevelopment(true), 
))
defer c.Close()

res := c.ClusterFailover(ctx)
```
输出:
```text
[CLUSTER FAILOVER]: redis command are not allowed
```

## 监控

导入`Grafana dashboard id` `16768`

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

## 自动`pipeline`

所有发送到单个`Redis`节点的非阻塞命令都会通过一个tcp连接自动`pipeline`传输，
这减少了整体往返和系统调用，并获得了更高的吞吐量。

注意：仅在使用`Redis RESP3`客户端时支持。


## 客户端缓存

始终启用服务器辅助客户端缓存的加入模式

```golang
c.Cache(time.Minute).Get(ctx, "key").Val()
```

需要显式指定客户端`TTL`，因为`Redis`服务器在以下情况下可能无法及时发送失效消息：
服务器上的密钥已过期。请遵循 [#6833](https://github.com/redis/redis/issues/6833) 和 [#6867](https://github.com/redis/redis/issues/6867)

尽管需要显式的指定客户端`TTL`，`Cache()`仍然向服务器发送`PTTL`命令，并确保客户端`TTL`不长于服务器端`TTL`。

注意：仅在使用`Redis RESP3`客户端时支持。

## Benchmark
### 环境
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

详见 [Benchmark Detail Result](https://github.com/sandwich-go/go-redis-client-benchmark)



* [Opt-in client side caching](https://redis.io/docs/manual/client-side-caching/)
* [RESP](https://redis.io/docs/reference/protocol-spec/)
* [RESP2](https://github.com/redis/redis-specifications/blob/master/protocol/RESP2.md)
* [RESP3](https://github.com/antirez/RESP3/blob/master/spec.md)