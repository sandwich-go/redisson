package redisson

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (b builder) BLMoveCompleted(source, destination, srcpos, destpos string, timeout time.Duration) Completed {
	return b.Arbitrary(KwBLMove).Keys(source, destination).Args(srcpos, destpos, strconv.FormatFloat(float64(formatSec(timeout)), 'f', -1, 64)).Blocking()
}

func (b builder) BLPopCompleted(timeout time.Duration, keys ...string) Completed {
	return b.Blpop().Key(keys...).Timeout(float64(formatSec(timeout))).Build()
}

func (b builder) BRPopCompleted(timeout time.Duration, keys ...string) Completed {
	return b.Brpop().Key(keys...).Timeout(float64(formatSec(timeout))).Build()
}

func (b builder) BRPopLPushCompleted(source, destination string, timeout time.Duration) Completed {
	return b.Brpoplpush().Source(source).Destination(destination).Timeout(float64(formatSec(timeout))).Build()
}

func (b builder) LIndexCompleted(key string, index int64) Completed {
	return b.Lindex().Key(key).Index(index).Build()
}

func (b builder) LInsertCompleted(key, op string, pivot, element any) Completed {
	switch strings.ToUpper(op) {
	case BEFORE:
		return b.Linsert().Key(key).Before().Pivot(str(pivot)).Element(str(element)).Build()
	case AFTER:
		return b.Linsert().Key(key).After().Pivot(str(pivot)).Element(str(element)).Build()
	default:
		panic(fmt.Sprintf("Invalid op argument value: %s", op))
	}
}

func (b builder) LInsertBeforeCompleted(key string, pivot, element any) Completed {
	return b.Linsert().Key(key).Before().Pivot(str(pivot)).Element(str(element)).Build()
}

func (b builder) LInsertAfterCompleted(key string, pivot, element any) Completed {
	return b.Linsert().Key(key).After().Pivot(str(pivot)).Element(str(element)).Build()
}

func (b builder) LLenCompleted(key string) Completed {
	return b.Llen().Key(key).Build()
}

func (b builder) LMoveCompleted(source, destination, srcpos, destpos string) Completed {
	return b.Arbitrary(KwLMove).Keys(source, destination).Args(srcpos, destpos).Build()
}

func (b builder) LPopCompleted(key string) Completed {
	return b.Lpop().Key(key).Build()
}

func (b builder) LPopCountCompleted(key string, count int64) Completed {
	return b.Lpop().Key(key).Count(count).Build()
}

func (b builder) LMPopCompleted(direction string, count int64, keys ...string) Completed {
	cmd := b.Arbitrary(KwLMPop, strconv.Itoa(len(keys))).Keys(keys...)
	cmd = cmd.Args(direction)
	if count > 0 {
		cmd = cmd.Args(KwCount, strconv.FormatInt(count, 10))
	}
	return cmd.Build()
}

func (b builder) LPosCompleted(key string, element string, a LPosArgs) Completed {
	cmd := b.Arbitrary(KwLPos).Keys(key).Args(element)
	if a.Rank != 0 {
		cmd = cmd.Args(KwRank, strconv.FormatInt(a.Rank, 10))
	}
	if a.MaxLen != 0 {
		cmd = cmd.Args(KwMaxLen, strconv.FormatInt(a.MaxLen, 10))
	}
	return cmd.Build()
}

func (b builder) LPosCountCompleted(key string, element string, count int64, a LPosArgs) Completed {
	cmd := b.Arbitrary(KwLPos).Keys(key).Args(element).Args(KwCount, strconv.FormatInt(count, 10))
	if a.Rank != 0 {
		cmd = cmd.Args(KwRank, strconv.FormatInt(a.Rank, 10))
	}
	if a.MaxLen != 0 {
		cmd = cmd.Args(KwMaxLen, strconv.FormatInt(a.MaxLen, 10))
	}
	return cmd.Build()
}

func (b builder) LPushCompleted(key string, element any) Completed {
	return b.Lpush().Key(key).Element(str(element)).Build()
}

func (b builder) LMPushCompleted(key string, elements ...any) Completed {
	return b.Lpush().Key(key).Element(argsToSlice(elements)...).Build()
}

func (b builder) LPushXCompleted(key string, element any) Completed {
	return b.Lpushx().Key(key).Element(str(element)).Build()
}

func (b builder) LMPushXCompleted(key string, elements ...any) Completed {
	return b.Lpushx().Key(key).Element(argsToSlice(elements)...).Build()
}

func (b builder) LRangeCompleted(key string, start, stop int64) Completed {
	return b.Lrange().Key(key).Start(start).Stop(stop).Build()
}

func (b builder) LRemCompleted(key string, count int64, element any) Completed {
	return b.Lrem().Key(key).Count(count).Element(str(element)).Build()
}

func (b builder) LSetCompleted(key string, index int64, element any) Completed {
	return b.Lset().Key(key).Index(index).Element(str(element)).Build()
}

func (b builder) LTrimCompleted(key string, start, stop int64) Completed {
	return b.Ltrim().Key(key).Start(start).Stop(stop).Build()
}

func (b builder) RPopCompleted(key string) Completed {
	return b.Rpop().Key(key).Build()
}

func (b builder) RPopCountCompleted(key string, count int64) Completed {
	return b.Rpop().Key(key).Count(count).Build()
}

func (b builder) RPopLPushCompleted(source, destination string) Completed {
	return b.Rpoplpush().Source(source).Destination(destination).Build()
}

func (b builder) RPushCompleted(key string, element any) Completed {
	return b.Rpush().Key(key).Element(str(element)).Build()
}

func (b builder) RMPushCompleted(key string, elements ...any) Completed {
	return b.Rpush().Key(key).Element(argsToSlice(elements)...).Build()
}

func (b builder) RPushXCompleted(key string, element any) Completed {
	return b.Rpushx().Key(key).Element(str(element)).Build()
}

func (b builder) RMPushXCompleted(key string, elements ...any) Completed {
	return b.Rpushx().Key(key).Element(argsToSlice(elements)...).Build()
}
