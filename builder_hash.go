package redisson

import (
	"strconv"
	"time"
)

func (b builder) HDelCompleted(key string, fieldS ...string) Completed {
	return b.Hdel().Key(key).Field(fieldS...).Build()
}

func (b builder) HExistsCompleted(key, field string) Completed {
	return b.Hexists().Key(key).Field(field).Build()
}

func (b builder) HExpireCompleted(key string, seconds time.Duration, fields ...string) Completed {
	return b.Hexpire().Key(key).Seconds(formatSec(seconds)).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireNXCompleted(key string, seconds time.Duration, fields ...string) Completed {
	return b.Hexpire().Key(key).Seconds(formatSec(seconds)).Nx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireXXCompleted(key string, seconds time.Duration, fields ...string) Completed {
	return b.Hexpire().Key(key).Seconds(formatSec(seconds)).Xx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireGTCompleted(key string, seconds time.Duration, fields ...string) Completed {
	return b.Hexpire().Key(key).Seconds(formatSec(seconds)).Gt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireLTCompleted(key string, seconds time.Duration, fields ...string) Completed {
	return b.Hexpire().Key(key).Seconds(formatSec(seconds)).Lt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireAtCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hexpireat().Key(key).UnixTimeSeconds(tm.Unix()).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireAtNXCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hexpireat().Key(key).UnixTimeSeconds(tm.Unix()).Nx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireAtXXCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hexpireat().Key(key).UnixTimeSeconds(tm.Unix()).Xx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireAtGTCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hexpireat().Key(key).UnixTimeSeconds(tm.Unix()).Gt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireAtLTCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hexpireat().Key(key).UnixTimeSeconds(tm.Unix()).Lt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HExpireTimeCompleted(key string, fields ...string) Completed {
	return b.Hexpiretime().Key(key).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HGetCompleted(key, field string) Completed {
	return b.Hget().Key(key).Field(field).Build()
}

func (b builder) HGetAllCompleted(key string) Completed {
	return b.Hgetall().Key(key).Build()
}

func (b builder) HIncrByCompleted(key, field string, incr int64) Completed {
	return b.Hincrby().Key(key).Field(field).Increment(incr).Build()
}

func (b builder) HIncrByFloatCompleted(key, field string, incr float64) Completed {
	return b.Hincrbyfloat().Key(key).Field(field).Increment(incr).Build()
}

func (b builder) HKeysCompleted(key string) Completed {
	return b.Hkeys().Key(key).Build()
}

func (b builder) HLenCompleted(key string) Completed {
	return b.Hlen().Key(key).Build()
}

func (b builder) HMGetCompleted(key string, fields ...string) Completed {
	return b.Hmget().Key(key).Field(fields...).Build()
}

func (b builder) HMSetCompleted(key string, values ...any) Completed {
	partial := b.Hset().Key(key).FieldValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		partial = partial.FieldValue(args[i], args[i+1])
	}
	return partial.Build()
}

func (b builder) HPersistCompleted(key string, fields ...string) Completed {
	return b.Hpersist().Key(key).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireCompleted(key string, milliseconds time.Duration, fields ...string) Completed {
	return b.Hpexpire().Key(key).Milliseconds(formatMs(milliseconds)).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireNXCompleted(key string, milliseconds time.Duration, fields ...string) Completed {
	return b.Hpexpire().Key(key).Milliseconds(formatMs(milliseconds)).Nx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireXXCompleted(key string, milliseconds time.Duration, fields ...string) Completed {
	return b.Hpexpire().Key(key).Milliseconds(formatMs(milliseconds)).Xx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireGTCompleted(key string, milliseconds time.Duration, fields ...string) Completed {
	return b.Hpexpire().Key(key).Milliseconds(formatMs(milliseconds)).Gt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireLTCompleted(key string, milliseconds time.Duration, fields ...string) Completed {
	return b.Hpexpire().Key(key).Milliseconds(formatMs(milliseconds)).Lt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireAtCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hpexpireat().Key(key).UnixTimeMilliseconds(tm.UnixNano() / int64(time.Millisecond)).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireAtNXCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hpexpireat().Key(key).UnixTimeMilliseconds(tm.UnixNano() / int64(time.Millisecond)).Nx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireAtXXCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hpexpireat().Key(key).UnixTimeMilliseconds(tm.UnixNano() / int64(time.Millisecond)).Xx().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireAtGTCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hpexpireat().Key(key).UnixTimeMilliseconds(tm.UnixNano() / int64(time.Millisecond)).Gt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireAtLTCompleted(key string, tm time.Time, fields ...string) Completed {
	return b.Hpexpireat().Key(key).UnixTimeMilliseconds(tm.UnixNano() / int64(time.Millisecond)).Lt().Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPExpireTimeCompleted(key string, fields ...string) Completed {
	return b.Hpexpiretime().Key(key).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HTTLCompleted(key string, fields ...string) Completed {
	return b.Httl().Key(key).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HPTTLCompleted(key string, fields ...string) Completed {
	return b.Hpttl().Key(key).Fields().Numfields(int64(len(fields))).Field(fields...).Build()
}

func (b builder) HRandFieldCompleted(key string, count int64) Completed {
	return b.Hrandfield().Key(key).Count(count).Build()
}

func (b builder) HRandFieldWithValuesCompleted(key string, count int64) Completed {
	return b.Hrandfield().Key(key).Count(count).Withvalues().Build()
}

func (b builder) HScanCompleted(key string, cursor uint64, match string, count int64) Completed {
	cmd := b.Arbitrary(XXX_HSCAN).Keys(key).Args(strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(XXX_MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(XXX_COUNT, strconv.FormatInt(count, 10))
	}
	return cmd.ReadOnly()
}

func (b builder) HSetCompleted(key, field string, value any) Completed {
	partial := b.Hset().Key(key).FieldValue()
	partial = partial.FieldValue(field, str(value))
	return partial.Build()
}

func (b builder) HMSetXCompleted(key string, values ...any) Completed {
	partial := b.Hset().Key(key).FieldValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		partial = partial.FieldValue(args[i], args[i+1])
	}
	return partial.Build()
}

func (b builder) HSetNXCompleted(key, field string, value any) Completed {
	return b.Hsetnx().Key(key).Field(field).Value(str(value)).Build()
}

func (b builder) HValsCompleted(key string) Completed {
	return b.Hvals().Key(key).Build()
}

func (b builder) HStrLenCompleted(key, field string) Completed {
	return b.Hstrlen().Key(key).Field(field).Build()
}
