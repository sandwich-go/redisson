// 示例：Hash 命令的常用用法（HSet/HGetAll/HMSet/HExpire）。
//
// 运行:
//
//	go run ./examples/hash
//
// HExpire 系列需要 Redis 7.4+；7.2 上仅打印错误，不影响其余演示。
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
	c := redisson.MustNewClient(redisson.NewConf(redisson.WithAddrs(addr())))
	defer c.Close()
	ctx := context.Background()

	const key = "user:1001"
	defer c.Del(ctx, key)

	// 一次写入多 field
	if err := c.HSet(ctx, key,
		"name", "alice",
		"age", "30",
		"email", "alice@example.com",
	).Err(); err != nil {
		log.Fatalf("HSet: %v", err)
	}

	// 读单 field
	name, _ := c.HGet(ctx, key, "name").Result()
	fmt.Printf("name=%q\n", name)

	// 读全部
	all, _ := c.HGetAll(ctx, key).Result()
	fmt.Printf("all = %+v\n", all)

	// 计数
	c.HIncrBy(ctx, key, "age", 1)
	age, _ := c.HGet(ctx, key, "age").Int64()
	fmt.Printf("age after incr = %d\n", age)

	// HMSet 兼容 API
	if err := c.HMSet(ctx, key, map[string]any{
		"city":    "Beijing",
		"country": "CN",
	}).Err(); err != nil {
		log.Fatalf("HMSet: %v", err)
	}

	// Field 级别过期（Redis 7.4+，旧版本会返回 unknown command）
	if err := c.HExpire(ctx, key, 5*time.Minute, "email").Err(); err != nil {
		fmt.Printf("HExpire (need Redis 7.4+): %v\n", err)
	} else {
		fmt.Println("email field set to expire in 5min")
	}
}
