package redisson

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var moveDelayTaskLua = `
local source_set, target_set  = KEYS[1], KEYS[2]
local max_priority, score = ARGV[1], ARGV[2]
local items = redis.call('ZRANGEBYSCORE', source_set, '-inf', max_priority, 'WITHSCORES')
for i, value in ipairs(items) do
	if i % 2 ~= 0 then
		redis.call('ZADD', target_set, score or 0.0, value)
		redis.call('ZREM', source_set, value)
	end
end
return items
`

var delayTaskLengthLua = `
local delay_set, doing_set  = KEYS[1], KEYS[2]
local l1 = redis.call('ZCARD', delay_set)
local l2 = redis.call('ZCARD', doing_set)
return l1+l2
`

var addDelayTaskLua = `
local delay_set  = KEYS[1]
local value, score = ARGV[1], ARGV[2]
redis.call('ZADD', delay_set, score or 0.0, value)
return {true}
`

var consumeDelayTaskSuccessLua = `
local doing_set  = KEYS[1]
local value = ARGV[1]
redis.call('ZREM', doing_set, value)
return {true}
`

var consumeDelayTaskFailedLua = `
local delay_set, doing_set  = KEYS[1], KEYS[2]
local value, score = ARGV[1], tonumber(ARGV[2])
redis.call('ZREM', doing_set, value)
redis.call('ZADD', delay_set, score-1, value)
return {true}
`

var (
	ErrEmptyDelayQueueName     = errors.New(" delay queue name cannot be empty")
	ErrEmptyDelayQueueCallback = errors.New(" delay queue callback cannot be empty")
	ErrDelayQueueHasClosed     = errors.New("delay queue has closed")
	ErrDelayQueueHasStarted    = errors.New("delay queue has started")
)

const (
	delayLogPrefix      = "[redis-delay]:"
	delayKeyFormat      = "do:{%s}"
	delayDoingKeyFormat = "doing:{%s}"
)

type DelayQueue interface {
	Add(ctx context.Context, bytes []byte, seconds time.Duration) error
	Length(ctx context.Context) (int64, error)
	Close() error
}

type delayQueue struct {
	c       *client
	spec    DelayOptionsVisitor
	wg      sync.WaitGroup
	exitC   chan struct{}
	name    string
	running atomic.Bool

	pollKeys    []string
	reclaimKeys []string

	moveScript           Scripter
	addScript            Scripter
	lengthScript         Scripter
	consumeSuccessScript Scripter
	consumeFailedScript  Scripter

	callback func([]byte) error
}

func newDelayQueue(c *client, name string, f func([]byte) error, opts ...DelayOption) (*delayQueue, error) {
	if name == "" {
		return nil, ErrEmptyDelayQueueName
	}
	if f == nil {
		return nil, ErrEmptyDelayQueueCallback
	}
	spec := newDelayOptions(opts...)
	q := &delayQueue{
		c:                    c,
		spec:                 spec,
		name:                 name,
		moveScript:           c.CreateScript(moveDelayTaskLua),
		addScript:            c.CreateScript(addDelayTaskLua),
		lengthScript:         c.CreateScript(delayTaskLengthLua),
		consumeSuccessScript: c.CreateScript(consumeDelayTaskSuccessLua),
		consumeFailedScript:  c.CreateScript(consumeDelayTaskFailedLua),
		callback:             f,
	}
	delayKey := fmt.Sprintf(delayKeyFormat, name)
	doingKey := fmt.Sprintf(delayDoingKeyFormat, name)
	if prefix := spec.GetPrefix(); prefix != "" {
		delayKey = fmt.Sprintf("%s:%s", prefix, delayKey)
		doingKey = fmt.Sprintf("%s:%s", prefix, doingKey)
	}
	q.pollKeys = []string{delayKey, doingKey}
	q.reclaimKeys = []string{doingKey, delayKey}
	err := q.run(ticker{d: time.Second, f: q.poll}, ticker{d: time.Second, f: q.reclaim})
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (q *delayQueue) Add(ctx context.Context, bytes []byte, seconds time.Duration) error {
	sec := formatSec(seconds)
	now := nowFunc().Unix()
	return q.addScript.Run(ctx, q.pollKeys, bytes, now+sec).Err()
}

func (q *delayQueue) Length(ctx context.Context) (int64, error) {
	return q.lengthScript.Run(ctx, q.pollKeys).Int64()
}

func (q *delayQueue) move(ctx context.Context, from, to string, offset float64) ([]any, error) {
	now := nowFunc().Unix()
	return q.moveScript.Run(ctx, []string{from, to}, now, float64(now)+offset).Slice()
}

func (q *delayQueue) isRunning() bool { return q.running.Load() }

func (q *delayQueue) Close() error {
	if !q.isRunning() {
		return ErrDelayQueueHasClosed
	}
	q.running.Store(false)
	close(q.exitC)
	q.wg.Wait()
	q.c.delayQueues.Delete(q.name)
	return nil
}

type ticker struct {
	d time.Duration
	f func() error
}

func (q *delayQueue) run(ts ...ticker) error {
	if q.isRunning() {
		return ErrDelayQueueHasStarted
	}
	q.running.Store(true)
	q.wg.Add(len(ts))
	q.exitC = make(chan struct{})
	var doTicker = func(ti ticker) {
		t := time.NewTimer(0)
		defer func() {
			_ = t.Stop()
			q.wg.Done()
		}()
		for {
			select {
			case <-t.C:
				_ = t.Reset(ti.d)
				go func() {
					if err := ti.f(); err != nil {
						// 输出日志
						e(fmt.Sprintf("%s ticker error, %v", delayLogPrefix, err))
					}
				}()
			case <-q.exitC:
				return
			}
		}
	}
	for _, ti := range ts {
		go func(_ti ticker) {
			doTicker(_ti)
		}(ti)
	}
	return nil
}

func (q *delayQueue) poll() error {
	now := nowFunc()
	res, err := q.moveScript.Run(context.Background(), q.pollKeys, now.Unix(), float64(now.Add(q.spec.GetTimeout()).Unix())).Slice()
	if err != nil {
		q.c.handler.delayPollError(q.name)
		return err
	}
	for i := 0; i < len(res); i += 2 {
		data := []byte(res[i].(string))
		score, _ := strconv.ParseInt(res[i+1].(string), 10, 64)
		if score < 0 {
			// 曾经失败过
			if int(math.Abs(float64(score))) > q.spec.GetRetryTimes() {
				// 死信
				q.handleDeadLetter(data)
				continue
			}
		} else {
			score = 0
		}
		if err0 := q.handle(data); err0 != nil {
			// 处理失败
			_ = q.retryAdd(data, float64(score))
		} else {
			// 处理成功
			_ = q.ackOK(data)
		}
	}
	return nil
}

func (q *delayQueue) handle(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handle task panic, %v", r)
			e(fmt.Sprintf("%s %v", delayLogPrefix, err))
			return
		}
	}()
	err = q.callback(data)
	return
}

func (q *delayQueue) handleDeadLetter(data []byte) {
	_ = q.ackOK(data)
	if f := q.spec.GetHandleDeadLetter(); f != nil {
		defer func() {
			if r := recover(); r != nil {
				e(fmt.Sprintf("%s handle dead letter panic, %v", delayLogPrefix, r))
				return
			}
		}()
		f(data)
	}
}

func (q *delayQueue) reclaim() error {
	now := nowFunc()
	res, err := q.moveScript.Run(context.Background(), q.reclaimKeys, now.Unix(), float64(now.Add(q.spec.GetTimeout()).Unix())).Slice()
	if err != nil {
		q.c.handler.delayReclaimError(q.name)
	} else if len(res) > 0 {
		q.c.handler.delayReclaim(q.name, len(res)/2)
	}
	return err
}

func (q *delayQueue) ackOK(data []byte) error {
	err := q.consumeSuccessScript.Run(context.Background(), q.reclaimKeys, data).Err()
	if err != nil {
		e(fmt.Sprintf("%s ack failed, %v, %v", delayLogPrefix, data, err))
	}
	return err
}

func (q *delayQueue) retryAdd(data []byte, score float64) error {
	err := q.consumeFailedScript.Run(context.Background(), q.pollKeys, data, score).Err()
	if err != nil {
		e(fmt.Sprintf("%s retry add failed, %v, %f, %v", delayLogPrefix, data, score, err))
	}
	return err
}

// NewDelayQueue 新建一个延迟队列
func (c *client) NewDelayQueue(name string, f func([]byte) error, opts ...DelayOption) (DelayQueue, error) {
	if val, ok := c.delayQueues.Load(name); ok {
		return val.(*delayQueue), nil
	}
	q, err := newDelayQueue(c, name, f, opts...)
	if err != nil {
		return nil, err
	}
	c.delayQueues.Store(q.name, q)
	return q, nil
}
