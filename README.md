# redisson

A Type-safe Golang Redis RESP2/RESP3 client.

## Features

* Check Redis commands version in development mode.
* Check Redis deprecated commands in development mode.
* Check Redis slot when use multiple keys in development mode.
* Monitoring cost of Redis commands.
* Monitoring status of connections.
* Monitoring hits/miss of Redis RESP3 client side caching.
* Support Redis RESP2/RESP3.
* Opt-in client side caching.
* Auto pipeline for non-blocking Redis RESP3 commands.
* Connection pooling for blocking Redis RESP3 commands.

## Requirement

* Currently, only supports Redis < 7.x

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
	defer func() {
		_ = c.Close()
    }()

	ctx := context.Background()

	// SET key val NX
	_ = c.SetNX(ctx, "key", "val", 0).Err()
	// GET key
	_ = c.Get(ctx, "key").Val()
}
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


