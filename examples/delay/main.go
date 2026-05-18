// 示例：DelayQueue 演示延迟投递、消费、重试与死信处理。
//
// 运行:
//
//	go run ./examples/delay
//
// 程序会在 2 秒延迟后处理一条任务，再投递一条永远失败的任务，
// 演示重试与死信路径。约 5 秒后自动退出。
package main

import (
	"context"
	"errors"
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

	// 演示 1：成功路径
	doneCh := make(chan struct{})
	q1, err := c.NewDelayQueue("demo-success", func(payload []byte) error {
		fmt.Printf("[demo-success] handled: %s\n", payload)
		close(doneCh)
		return nil
	},
		redisson.WithDelayOptionPrefix("examples"),
		redisson.WithDelayOptionVisibilityTimeout(10*time.Second),
		redisson.WithDelayOptionRetryTimes(3),
	)
	if err != nil {
		log.Fatalf("NewDelayQueue: %v", err)
	}
	defer q1.Close()

	// 投一条 2 秒延迟的任务
	if err := q1.Add(ctx, []byte("task-A"), 2*time.Second); err != nil {
		log.Fatalf("Add: %v", err)
	}
	fmt.Println("task-A scheduled after 2s ...")

	<-doneCh

	// 演示 2：死信路径
	deadCh := make(chan []byte, 1)
	q2, err := c.NewDelayQueue("demo-dead", func(payload []byte) error {
		fmt.Printf("[demo-dead] retry attempt for: %s\n", payload)
		return errors.New("simulated failure")
	},
		redisson.WithDelayOptionPrefix("examples"),
		redisson.WithDelayOptionVisibilityTimeout(10*time.Second),
		redisson.WithDelayOptionRetryTimes(2), // 失败 2 次即死信
		redisson.WithDelayOptionHandleDeadLetter(func(bs []byte) {
			fmt.Printf("[demo-dead] dead letter: %s\n", bs)
			deadCh <- bs
		}),
	)
	if err != nil {
		log.Fatalf("NewDelayQueue (dead): %v", err)
	}
	defer q2.Close()

	if err := q2.Add(ctx, []byte("task-B"), 500*time.Millisecond); err != nil {
		log.Fatalf("Add B: %v", err)
	}

	select {
	case <-deadCh:
		fmt.Println("done.")
	case <-time.After(15 * time.Second):
		log.Fatal("dead letter timeout")
	}
}
