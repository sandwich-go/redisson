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

* 当前只支持 Redis < 7.x
* Golang >= 1.8

## 链接
* [English](https://github.com/sandwich-go/redisson/README.md)
* [中文文档](https://github.com/sandwich-go/redisson/README_CN.md)

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
```go
Line 34: - redis 'SET KEEPTTL' are not supported in version "5.0.0", available since 6.0.0
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
```go
As of Redis version 4.0.0, this command is regarded as deprecated.
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
```go
Line 34: - multi key command with different key slots are not allowed 
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
```go
Line 34: - command 'CLUSTER FAILOVER' not allowed 
```

## 监控

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
服务器上的密钥已过期。请遵循 [#6833](https://github.com/redis/redis/issues/6833) and [#6867](https://github.com/redis/redis/issues/6867)

尽管需要显式的指定客户端`TTL`，`Cache()`仍然向服务器发送`PTTL`命令，并确保客户端`TTL`不长于服务器端`TTL`。

注意：仅在使用`Redis RESP3`客户端时支持。


* [Opt-in client side caching](https://redis.io/docs/manual/client-side-caching/)
* [RESP](https://redis.io/docs/reference/protocol-spec/)
* [RESP2](https://github.com/redis/redis-specifications/blob/master/protocol/RESP2.md)
* [RESP3](https://github.com/antirez/RESP3/blob/master/spec.md)