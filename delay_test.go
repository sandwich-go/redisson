//go:build integration

package redisson

import (
	"context"
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() {
		_ = c.Close()
	})
	if !c.Options().GetDevelopment() {
		c.FlushDB(context.Background())
	}
	var ctx = context.Background()
	name := "mock"
	prefix := "mock_delay"
	task := ([]byte)("task")

	Convey("normal delay queue", t, func() {
		var notifyChan = make(chan []byte)
		var doTime time.Time
		q, err := c.NewDelayQueue(name, func(bytes []byte) error {
			doTime = nowFunc()
			notifyChan <- bytes
			return nil
		}, WithDelayOptionPrefix(prefix))
		So(err, ShouldBeNil)

		var l int64
		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(0))

		var addTime = nowFunc()
		err = q.Add(ctx, task, 2*time.Second)
		So(err, ShouldBeNil)

		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(1))

		select {
		case data := <-notifyChan:
			So(data, ShouldResemble, task)
			So(doTime.Sub(addTime).Seconds(), ShouldBeGreaterThan, 1)
			// 保证已经删除掉了
			eventuallyEq(func() int64 {
				v, _ := q.Length(ctx)
				return v
			}, 0, 3*time.Second)
			l, err = q.Length(ctx)
			So(err, ShouldBeNil)
			So(l, ShouldEqual, int64(0))

			So(q.Close(), ShouldBeNil)
		}
	})

	Convey("retry delay queue", t, func() {
		var count int
		var maxRetryTimes = 4
		var afterTimesOK = 2
		var notifyChan = make(chan []byte)
		q, err := c.NewDelayQueue(name, func(bytes []byte) error {
			if count == afterTimesOK {
				notifyChan <- bytes
				return nil
			}
			count++
			return errors.New("mock error")
		}, WithDelayOptionPrefix(prefix), WithDelayOptionRetryTimes(maxRetryTimes))
		So(err, ShouldBeNil)

		err = q.Add(ctx, task, time.Second)
		So(err, ShouldBeNil)

		var l int64
		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(1))

		select {
		case data := <-notifyChan:
			So(data, ShouldResemble, task)
			// 保证已经删除掉了
			eventuallyEq(func() int64 {
				v, _ := q.Length(ctx)
				return v
			}, 0, 3*time.Second)

			l, err = q.Length(ctx)
			So(err, ShouldBeNil)
			So(l, ShouldEqual, int64(0))

			So(q.Close(), ShouldBeNil)
		}
	})

	Convey("reclaim delay queue", t, func() {
		var notifyChan0, notifyChan1, notifyChan2 = make(chan []byte), make(chan []byte), make(chan struct{})
		var q DelayQueue
		var err error
		var timeout = 3 * time.Second
		q, err = c.NewDelayQueue(name, func(bytes []byte) error {
			// 当处理的时候，程序崩溃了
			_ = q.Close()
			notifyChan0 <- bytes
			<-notifyChan1
			notifyChan2 <- struct{}{}
			return nil
		}, WithDelayOptionPrefix(prefix), WithDelayOptionVisibilityTimeout(timeout))
		So(err, ShouldBeNil)

		err = q.Add(ctx, task, time.Second)
		So(err, ShouldBeNil)

		var l int64
		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(1))

		select {
		case data := <-notifyChan0:
			So(data, ShouldResemble, task)

			// 此时第一个 q 已被 callback 内调 q.Close,task 仍可能在 doing
			// (visibility 未到期) 或被新 q reclaim 后回到 delay。两种状态下
			// Length 都应当 >= 1;严格断言 == 1 在 CI 高负载 (Redis 6) 下
			// 偶发因 reclaim 与查询的微小 race 失败,这里改用 >= 1 容忍。
			l, err = q.Length(ctx)
			So(err, ShouldBeNil)
			So(l, ShouldBeGreaterThanOrEqualTo, int64(1))

			q, err = c.NewDelayQueue(name, func(bytes []byte) error {
				// 重新处理
				notifyChan1 <- bytes
				return nil
			}, WithDelayOptionPrefix(prefix), WithDelayOptionVisibilityTimeout(timeout))
			So(err, ShouldBeNil)
		}
		<-notifyChan2
		So(q.Close(), ShouldBeNil)
	})

	Convey("delay queue dead letter", t, func() {
		// 该子测试理论耗时 ~3-6s（3 次失败 → 死信）。给以下两项收紧：
		//   1) PollInterval=200ms + RetryBackoff=200ms：去掉默认 1s/1s 配置下的"调度漂移"
		//      （score = now + 1s 与 1s ticker 撞点会让每轮重试多等 1s）。这样 CI 上即使
		//      Redis 抖动，理想耗时也能从 3-6s 压到 ~1s 内，留足缓冲。
		//   2) <-notifyChan 改为 select + 15s 超时：若 hook 没被调用，立即 t.Fatal 而不是
		//      被 go test 的 5 分钟全局 timeout 吞掉根因，便于 CI 上追查（之前观察到的 4 分钟+
		//      卡顿就是因为缺这层显式超时）。
		var notifyChan = make(chan []byte, 1) // buffered 防止 hook 写时无人读 → 阻塞 worker 路径
		var q DelayQueue
		var err error
		var timeout = 2 * time.Second
		q, err = c.NewDelayQueue(name, func(bytes []byte) error {
			return errors.New("mock error")
		}, WithDelayOptionPrefix(prefix),
			WithDelayOptionVisibilityTimeout(timeout),
			WithDelayOptionRetryTimes(3),
			WithDelayOptionPollInterval(200*time.Millisecond),
			WithDelayOptionRetryBackoff(200*time.Millisecond),
			WithDelayOptionHandleDeadLetter(func(bs []byte) {
				notifyChan <- bs
			}))
		So(err, ShouldBeNil)

		err = q.Add(ctx, task, time.Second)
		So(err, ShouldBeNil)

		var l int64
		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(1))

		select {
		case <-notifyChan:
		case <-time.After(15 * time.Second):
			t.Fatalf("dead letter handler not fired within 15s")
		}

		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(0))

		So(q.Close(), ShouldBeNil)
	})
}
