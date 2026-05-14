// 示例：Locker 互斥锁的常用模式（WithContext / TryWithContext）。
//
// 运行:
//
//	go run ./examples/locker
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/rueidis/rueidislock"
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

	locker, err := c.NewLocker(
		redisson.WithLockerOptionKeyPrefix("examples-lock"),
		redisson.WithLockerOptionKeyValidity(10*time.Second),
	)
	if err != nil {
		log.Fatalf("NewLocker: %v", err)
	}

	parentCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 基础获取/释放
	lockCtx, release, err := locker.WithContext(parentCtx, "resource-1")
	if err != nil {
		log.Fatalf("acquire: %v", err)
	}
	fmt.Println("locked resource-1, doing work ...")
	doWork(lockCtx)
	release() // 释放
	fmt.Println("released resource-1")

	// TryWithContext：拿不到立即返回 ErrNotLocked，不阻塞
	// 模拟竞争：先抢一份
	holdCtx, holdRelease, err := locker.WithContext(parentCtx, "resource-2")
	if err != nil {
		log.Fatalf("first acquire: %v", err)
	}
	defer holdRelease()
	_ = holdCtx

	// 第二次 TryWithContext 应该立即拿不到（rueidislock.ErrNotLocked）
	_, _, err = locker.TryWithContext(parentCtx, "resource-2")
	if errors.Is(err, rueidislock.ErrNotLocked) {
		fmt.Println("TryWithContext correctly returned ErrNotLocked")
	} else if err != nil {
		fmt.Printf("TryWithContext err=%v\n", err)
	}
}

func doWork(ctx context.Context) {
	// 业务逻辑期间 ctx 在锁丢失时会被 cancel
	select {
	case <-time.After(200 * time.Millisecond):
	case <-ctx.Done():
		fmt.Printf("lock lost: %v\n", ctx.Err())
	}
}
