package redisson

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (b builder) AppendCompleted(key, value string) Completed {
	return b.Append().Key(key).Value(value).Build()
}

func (b builder) DecrCompleted(key string) Completed {
	return b.Decr().Key(key).Build()
}

func (b builder) DecrByCompleted(key string, decrement int64) Completed {
	return b.Decrby().Key(key).Decrement(decrement).Build()
}

func (b builder) GetCompleted(key string) Completed {
	return b.Get().Key(key).Build()
}

func (b builder) GetDelCompleted(key string) Completed {
	return b.Getdel().Key(key).Build()
}

func (b builder) GetExCompleted(key string, expiration time.Duration) Completed {
	if expiration > 0 {
		if usePrecise(expiration) {
			return b.Getex().Key(key).PxMilliseconds(formatMs(expiration)).Build()
		} else {
			return b.Getex().Key(key).ExSeconds(formatSec(expiration)).Build()
		}
	} else {
		return b.Getex().Key(key).Build()
	}
}

func (b builder) GetRangeCompleted(key string, start, end int64) Completed {
	return b.Getrange().Key(key).Start(start).End(end).Build()
}

func (b builder) GetSetCompleted(key string, value any) Completed {
	return b.Getset().Key(key).Value(str(value)).Build()
}

func (b builder) IncrCompleted(key string) Completed {
	return b.Incr().Key(key).Build()
}

func (b builder) IncrByCompleted(key string, increment int64) Completed {
	return b.Incrby().Key(key).Increment(increment).Build()
}

func (b builder) IncrByFloatCompleted(key string, increment float64) Completed {
	return b.Incrbyfloat().Key(key).Increment(increment).Build()
}

func (b builder) MGetCompleted(keys ...string) Completed {
	return b.Mget().Key(keys...).Build()
}

func (b builder) MSetCompleted(values ...any) Completed {
	partial := b.Mset().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		partial = partial.KeyValue(args[i], args[i+1])
	}
	return partial.Build()
}

func (b builder) MSetNXCompleted(values ...any) Completed {
	partial := b.Msetnx().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		partial = partial.KeyValue(args[i], args[i+1])
	}
	return partial.Build()
}

func (b builder) SetKeepTTLCompleted(key string, value any) Completed {
	return b.SetCompleted(key, value, KeepTTL)
}

func (b builder) SetCompleted(key string, value any, expiration time.Duration) Completed {
	if expiration > 0 {
		if usePrecise(expiration) {
			return b.Set().Key(key).Value(str(value)).PxMilliseconds(formatMs(expiration)).Build()
		} else {
			return b.Set().Key(key).Value(str(value)).ExSeconds(formatSec(expiration)).Build()
		}
	} else if expiration == KeepTTL {
		return b.Set().Key(key).Value(str(value)).Keepttl().Build()
	} else {
		return b.Set().Key(key).Value(str(value)).Build()
	}
}

func (b builder) SetArgsCompleted(key string, value any, a SetArgs) Completed {
	cmd := b.Arbitrary(XXX_SET).Keys(key).Args(str(value))
	if a.KeepTTL {
		cmd = cmd.Args(XXX_KEEPTTL)
	}
	if !a.ExpireAt.IsZero() {
		cmd = cmd.Args(XXX_EXAT, strconv.FormatInt(a.ExpireAt.Unix(), 10))
	}
	if a.TTL > 0 {
		if usePrecise(a.TTL) {
			cmd = cmd.Args(XXX_PX, strconv.FormatInt(formatMs(a.TTL), 10))
		} else {
			cmd = cmd.Args(XXX_EX, strconv.FormatInt(formatSec(a.TTL), 10))
		}
	}
	switch mode := strings.ToUpper(a.Mode); mode {
	case XX, NX:
		cmd = cmd.Args(mode)
	case "":
	default:
		panic(fmt.Sprintf("invalid mode for SET: %s", a.Mode))
	}
	if a.Get {
		cmd = cmd.Args(XXX_GET)
	}
	return cmd.Build()
}

func (b builder) SetEXCompleted(key string, value any, expiration time.Duration) Completed {
	return b.Setex().Key(key).Seconds(formatSec(expiration)).Value(str(value)).Build()
}

func (b builder) SetNXCompleted(key string, value any, expiration time.Duration) Completed {
	switch expiration {
	case 0:
		return b.Setnx().Key(key).Value(str(value)).Build()
	case KeepTTL:
		return b.Set().Key(key).Value(str(value)).Nx().Keepttl().Build()
	default:
		if usePrecise(expiration) {
			return b.Set().Key(key).Value(str(value)).Nx().PxMilliseconds(formatMs(expiration)).Build()
		} else {
			return b.Set().Key(key).Value(str(value)).Nx().ExSeconds(formatSec(expiration)).Build()
		}
	}
}

func (b builder) SetXXCompleted(key string, value any, expiration time.Duration) Completed {
	if expiration > 0 {
		if usePrecise(expiration) {
			return b.Set().Key(key).Value(str(value)).Xx().PxMilliseconds(formatMs(expiration)).Build()
		} else {
			return b.Set().Key(key).Value(str(value)).Xx().ExSeconds(formatSec(expiration)).Build()
		}
	} else if expiration == KeepTTL {
		return b.Set().Key(key).Value(str(value)).Xx().Keepttl().Build()
	} else {
		return b.Set().Key(key).Value(str(value)).Xx().Build()
	}
}

func (b builder) SetRangeCompleted(key string, offset int64, value string) Completed {
	return b.Setrange().Key(key).Offset(offset).Value(value).Build()
}

func (b builder) StrLenCompleted(key string) Completed {
	return b.Strlen().Key(key).Build()
}
