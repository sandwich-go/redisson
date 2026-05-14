package redisson

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (b builder) CopyCompleted(source string, destination string, db int64, replace bool) Completed {
	if replace {
		return b.Copy().Source(source).Destination(destination).Db(db).Replace().Build()
	} else {
		return b.Copy().Source(source).Destination(destination).Db(db).Build()
	}
}
func (b builder) DelCompleted(keys ...string) Completed { return b.Del().Key(keys...).Build() }
func (b builder) DumpCompleted(key string) Completed    { return b.Dump().Key(key).Build() }
func (b builder) ExistsCompleted(keys ...string) Completed {
	return b.Exists().Key(keys...).Build()
}

func (b builder) ExpireCompleted(key string, seconds time.Duration) Completed {
	return b.Expire().Key(key).Seconds(formatSec(seconds)).Build()
}

func (b builder) ExpireNXCompleted(key string, seconds time.Duration) Completed {
	return b.Expire().Key(key).Seconds(formatSec(seconds)).Nx().Build()
}

func (b builder) ExpireXXCompleted(key string, seconds time.Duration) Completed {
	return b.Expire().Key(key).Seconds(formatSec(seconds)).Xx().Build()
}

func (b builder) ExpireGTCompleted(key string, seconds time.Duration) Completed {
	return b.Expire().Key(key).Seconds(formatSec(seconds)).Gt().Build()
}

func (b builder) ExpireLTCompleted(key string, seconds time.Duration) Completed {
	return b.Expire().Key(key).Seconds(formatSec(seconds)).Lt().Build()
}

func (b builder) ExpireAtCompleted(key string, timestamp time.Time) Completed {
	return b.Expireat().Key(key).Timestamp(timestamp.Unix()).Build()
}

func (b builder) ExpireAtNXCompleted(key string, timestamp time.Time) Completed {
	return b.Expireat().Key(key).Timestamp(timestamp.Unix()).Nx().Build()
}

func (b builder) ExpireAtXXCompleted(key string, timestamp time.Time) Completed {
	return b.Expireat().Key(key).Timestamp(timestamp.Unix()).Xx().Build()
}

func (b builder) ExpireAtGTCompleted(key string, timestamp time.Time) Completed {
	return b.Expireat().Key(key).Timestamp(timestamp.Unix()).Gt().Build()
}

func (b builder) ExpireAtLTCompleted(key string, timestamp time.Time) Completed {
	return b.Expireat().Key(key).Timestamp(timestamp.Unix()).Lt().Build()
}

func (b builder) ExpireTimeCompleted(key string) Completed {
	return b.Expiretime().Key(key).Build()
}

func (b builder) KeysCompleted(pattern string) Completed {
	return b.Keys().Pattern(pattern).Build()
}

func (b builder) MigrateCompleted(host string, port int64, key string, db int64, timeout time.Duration) Completed {
	return b.Migrate().Host(host).Port(port).Key(key).DestinationDb(db).Timeout(formatSec(timeout)).Build()
}

func (b builder) MoveCompleted(key string, db int64) Completed {
	return b.Move().Key(key).Db(db).Build()
}

func (b builder) ObjectEncodingCompleted(key string) Completed {
	return b.ObjectEncoding().Key(key).Build()
}

func (b builder) ObjectIdleTimeCompleted(key string) Completed {
	return b.ObjectIdletime().Key(key).Build()
}

func (b builder) ObjectRefCountCompleted(key string) Completed {
	return b.ObjectRefcount().Key(key).Build()
}

func (b builder) PersistCompleted(key string) Completed {
	return b.Persist().Key(key).Build()
}

func (b builder) PExpireCompleted(key string, milliseconds time.Duration) Completed {
	return b.Pexpire().Key(key).Milliseconds(formatMs(milliseconds)).Build()
}

func (b builder) PExpireNXCompleted(key string, milliseconds time.Duration) Completed {
	return b.Pexpire().Key(key).Milliseconds(formatMs(milliseconds)).Nx().Build()
}

func (b builder) PExpireXXCompleted(key string, milliseconds time.Duration) Completed {
	return b.Pexpire().Key(key).Milliseconds(formatMs(milliseconds)).Xx().Build()
}

func (b builder) PExpireGTCompleted(key string, milliseconds time.Duration) Completed {
	return b.Pexpire().Key(key).Milliseconds(formatMs(milliseconds)).Gt().Build()
}

func (b builder) PExpireLTCompleted(key string, milliseconds time.Duration) Completed {
	return b.Pexpire().Key(key).Milliseconds(formatMs(milliseconds)).Lt().Build()
}

func (b builder) PExpireAtCompleted(key string, millisecondsTimestamp time.Time) Completed {
	return b.Pexpireat().Key(key).MillisecondsTimestamp(millisecondsTimestamp.UnixNano() / int64(time.Millisecond)).Build()
}

func (b builder) PExpireAtNXCompleted(key string, millisecondsTimestamp time.Time) Completed {
	return b.Pexpireat().Key(key).MillisecondsTimestamp(millisecondsTimestamp.UnixNano() / int64(time.Millisecond)).Nx().Build()
}

func (b builder) PExpireAtXXCompleted(key string, millisecondsTimestamp time.Time) Completed {
	return b.Pexpireat().Key(key).MillisecondsTimestamp(millisecondsTimestamp.UnixNano() / int64(time.Millisecond)).Xx().Build()
}

func (b builder) PExpireAtGTCompleted(key string, millisecondsTimestamp time.Time) Completed {
	return b.Pexpireat().Key(key).MillisecondsTimestamp(millisecondsTimestamp.UnixNano() / int64(time.Millisecond)).Gt().Build()
}

func (b builder) PExpireAtLTCompleted(key string, millisecondsTimestamp time.Time) Completed {
	return b.Pexpireat().Key(key).MillisecondsTimestamp(millisecondsTimestamp.UnixNano() / int64(time.Millisecond)).Lt().Build()
}

func (b builder) PExpireTimeCompleted(key string) Completed {
	return b.Pexpiretime().Key(key).Build()
}

func (b builder) PTTLCompleted(key string) Completed {
	return b.Pttl().Key(key).Build()
}

func (b builder) RandomKeyCompleted() Completed {
	return b.Randomkey().Build()
}

func (b builder) sort(command, key string, sort Sort) Completed {
	cmd := b.Arbitrary(command).Keys(key)
	if sort.By != "" {
		cmd = cmd.Args(KwBy, sort.By)
	}
	if sort.Offset != 0 || sort.Count != 0 {
		cmd = cmd.Args(KwLimit, strconv.FormatInt(sort.Offset, 10), strconv.FormatInt(sort.Count, 10))
	}
	for _, get := range sort.Get {
		cmd = cmd.Args(KwGet).Args(get)
	}
	switch order := strings.ToUpper(sort.Order); order {
	case ASC, DESC:
		cmd = cmd.Args(order)
	case "":
	default:
		panic(fmt.Sprintf("invalid sort order %s", sort.Order))
	}
	if sort.Alpha {
		cmd = cmd.Args(KwAlpha)
	}
	return cmd.Build()
}

func (b builder) SortCompleted(key string, sort Sort) Completed {
	return b.sort("SORT", key, sort)
}

func (b builder) SortROCompleted(key string, sort Sort) Completed {
	return b.sort("SORT_RO", key, sort)
}

func (b builder) RenameCompleted(key, newkey string) Completed {
	return b.Rename().Key(key).Newkey(newkey).Build()
}

func (b builder) RenameNXCompleted(key, newkey string) Completed {
	return b.Renamenx().Key(key).Newkey(newkey).Build()
}

func (b builder) RestoreCompleted(key string, ttl time.Duration, serializedValue string) Completed {
	return b.Restore().Key(key).Ttl(formatMs(ttl)).SerializedValue(serializedValue).Build()
}

func (b builder) RestoreReplaceCompleted(key string, ttl time.Duration, serializedValue string) Completed {
	return b.Restore().Key(key).Ttl(formatMs(ttl)).SerializedValue(serializedValue).Replace().Build()
}

func (b builder) ScanCompleted(cursor uint64, match string, count int64) Completed {
	cmd := b.Arbitrary(KwScan, strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(KwMatch, match)
	}
	if count > 0 {
		cmd = cmd.Args(KwCount, strconv.FormatInt(count, 10))
	}
	return cmd.ReadOnly()
}

func (b builder) ScanTypeCompleted(cursor uint64, match string, count int64, keyType string) Completed {
	cmd := b.Arbitrary(KwScan, strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(KwMatch, match)
	}
	if count > 0 {
		cmd = cmd.Args(KwCount, strconv.FormatInt(count, 10))
	}
	return cmd.Args(TYPE, keyType).ReadOnly()
}

func (b builder) TouchCompleted(keys ...string) Completed {
	return b.Touch().Key(keys...).Build()
}

func (b builder) TTLCompleted(key string) Completed {
	return b.Ttl().Key(key).Build()
}

func (b builder) TypeCompleted(key string) Completed {
	return b.Type().Key(key).Build()
}

func (b builder) UnlinkCompleted(keys ...string) Completed {
	return b.Unlink().Key(keys...).Build()
}
