package redisson

import (
	"context"
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	t.Cleanup(func() {
		_ = c.Close()
	})
	if !c.Options().GetDevelopment() {
		c.FlushAll(context.Background())
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
			time.Sleep(1 * time.Second)
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
			time.Sleep(1 * time.Second)

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
		}, WithDelayOptionPrefix(prefix), WithDelayOptionTimeout(timeout))
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

			l, err = q.Length(ctx)
			So(err, ShouldBeNil)
			So(l, ShouldEqual, int64(1))

			q, err = c.NewDelayQueue(name, func(bytes []byte) error {
				// 重新处理
				notifyChan1 <- bytes
				return nil
			}, WithDelayOptionPrefix(prefix), WithDelayOptionTimeout(timeout))
			So(err, ShouldBeNil)
		}
		<-notifyChan2
		So(q.Close(), ShouldBeNil)
	})

	Convey("delay queue dead letter", t, func() {
		var notifyChan = make(chan []byte)
		var q DelayQueue
		var err error
		var timeout = 2 * time.Second
		q, err = c.NewDelayQueue(name, func(bytes []byte) error {
			return errors.New("mock error")
		}, WithDelayOptionPrefix(prefix), WithDelayOptionTimeout(timeout), WithDelayOptionRetryTimes(3), WithDelayOptionHandleDeadLetter(func(bs []byte) {
			notifyChan <- bs
		}))
		So(err, ShouldBeNil)

		err = q.Add(ctx, task, time.Second)
		So(err, ShouldBeNil)

		var l int64
		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(1))

		<-notifyChan

		l, err = q.Length(ctx)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, int64(0))

		So(q.Close(), ShouldBeNil)
	})
}
