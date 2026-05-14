// 示例：String 命令的基础用法 + Pipeline 批量写入。
//
// 运行:
//
//	go run ./examples/string
//
// 环境变量 REDIS_ADDR 可覆盖默认地址 127.0.0.1:6379。
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sandwich-go/redisson"
)

func addr() string {
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		return v
	}
	return "127.0.0.1:6379"
}

func main() {
	c := redisson.MustNewClient(redisson.NewConf(
		redisson.WithAddrs(addr()),
		redisson.WithDevelopment(false),
	))
	defer c.Close()

	ctx := context.Background()

	// 基础 SET / GET
	if err := c.Set(ctx, "greeting", "hello", 0).Err(); err != nil {
		log.Fatalf("Set: %v", err)
	}
	v, err := c.Get(ctx, "greeting").Result()
	if err != nil {
		log.Fatalf("Get: %v", err)
	}
	fmt.Printf("greeting = %q\n", v)

	// 带过期：30 秒
	if err := c.Set(ctx, "tmp", "expires", 30*time.Second).Err(); err != nil {
		log.Fatalf("Set with TTL: %v", err)
	}
	ttl, _ := c.TTL(ctx, "tmp").Result()
	fmt.Printf("tmp ttl ≈ %v\n", ttl)

	// 原子自增
	c.Set(ctx, "counter", "0", 0)
	for i := 0; i < 5; i++ {
		_ = c.Incr(ctx, "counter").Err()
	}
	cnt, _ := c.Get(ctx, "counter").Int64()
	fmt.Printf("counter = %d\n", cnt)

	// Pipeline 批量
	pip := c.Pipeline()
	redisson.CommandSet.P(pip).Cmd("k1", "v1", 0)
	redisson.CommandSet.P(pip).Cmd("k2", "v2", 0)
	redisson.CommandSet.P(pip).Cmd("k3", "v3", 0)
	pipeRes, err := pip.Exec(ctx)
	if err != nil {
		log.Fatalf("Pipeline: %v", err)
	}
	fmt.Printf("pipeline executed %d commands\n", len(pipeRes))

	// 清理
	c.Del(ctx, "greeting", "tmp", "counter", "k1", "k2", "k3")
}
