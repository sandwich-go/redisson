package redisson

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type builder struct {
	Builder
}

func (b builder) GetBitCompleted(key string, offset int64) Completed {
	return b.Getbit().Key(key).Offset(offset).Build()
}

func (b builder) SetBitCompleted(key string, offset int64, value int64) Completed {
	return b.Setbit().Key(key).Offset(offset).Value(value).Build()
}

func (b builder) BitCountCompleted(key string, bc *BitCount) Completed {
	if bc == nil {
		return b.Bitcount().Key(key).Build()
	}
	if bc.Unit == "" {
		return b.Bitcount().Key(key).Start(bc.Start).End(bc.End).Build()
	}
	switch strings.ToUpper(bc.Unit) {
	case BitCountIndexByte:
		return b.Bitcount().Key(key).Start(bc.Start).End(bc.End).Byte().Build()
	case BitCountIndexBit:
		return b.Bitcount().Key(key).Start(bc.Start).End(bc.End).Bit().Build()
	default:
		panic(fmt.Sprintf("invalid unit %s", bc.Unit))
	}
}

func (b builder) BitOpAndCompleted(destKey string, keys ...string) Completed {
	return b.Bitop().And().Destkey(destKey).Key(keys...).Build()
}

func (b builder) BitOpOrCompleted(destKey string, keys ...string) Completed {
	return b.Bitop().Or().Destkey(destKey).Key(keys...).Build()
}

func (b builder) BitOpXorCompleted(destKey string, keys ...string) Completed {
	return b.Bitop().Xor().Destkey(destKey).Key(keys...).Build()
}

func (b builder) BitOpNotCompleted(destKey string, key string) Completed {
	return b.Bitop().Not().Destkey(destKey).Key(key).Build()
}

func (b builder) BitPosCompleted(key string, bit int64, pos ...int64) Completed {
	switch len(pos) {
	case 0:
		return b.Bitpos().Key(key).Bit(bit).Build()
	case 1:
		return b.Bitpos().Key(key).Bit(bit).Start(pos[0]).Build()
	case 2:
		return b.Bitpos().Key(key).Bit(bit).Start(pos[0]).End(pos[1]).Build()
	default:
		panic("too many arguments")
	}
}

func (b builder) BitPosSpanCompleted(key string, bit, start, end int64, span string) Completed {
	if strings.ToUpper(span) == BitCountIndexBit {
		return b.Bitpos().Key(key).Bit(bit).Start(start).End(end).Bit().Build()
	} else {
		return b.Bitpos().Key(key).Bit(bit).Start(start).End(end).Byte().Build()
	}
}

func (b builder) BitFieldCompleted(key string, args ...any) Completed {
	cmd := b.Arbitrary(BITFIELD).Keys(key)
	for _, v := range args {
		cmd = cmd.Args(str(v))
	}
	return cmd.Build()
}

func (b builder) EchoCompleted(message any) Completed { return b.Echo().Message(str(message)).Build() }
func (b builder) PingCompleted() Completed            { return b.Ping().Build() }
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
	cmd := b.Arbitrary(SCAN, strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, strconv.FormatInt(count, 10))
	}
	return cmd.ReadOnly()
}

func (b builder) ScanTypeCompleted(cursor uint64, match string, count int64, keyType string) Completed {
	cmd := b.Arbitrary(SCAN, strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, strconv.FormatInt(count, 10))
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
func (b builder) GeoAddCompleted(key string, geoLocation ...GeoLocation) Completed {
	cmd := b.Geoadd().Key(key).LongitudeLatitudeMember()
	for _, loc := range geoLocation {
		cmd = cmd.LongitudeLatitudeMember(loc.Longitude, loc.Latitude, loc.Name)
	}
	return cmd.Build()
}

func (b builder) GeoDistCompleted(key, member1, member2, unit string) Completed {
	switch strings.ToUpper(unit) {
	case M:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).M().Build()
	case MI:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).Mi().Build()
	case FT:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).Ft().Build()
	case EMPTY, KM:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).Km().Build()
	default:
		panic(fmt.Sprintf("invalid unit %s", unit))
	}
}

func (b builder) GeoHashCompleted(key string, members ...string) Completed {
	return b.Geohash().Key(key).Member(members...).Build()
}

func (b builder) GeoPosCompleted(key string, members ...string) Completed {
	return b.Geopos().Key(key).Member(members...).Build()
}

func (b builder) GeoRadiusByMemberCompleted(key, member string, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(GEORADIUSBYMEMBER_RO).Keys(key).Args(member)
	if query.Store != "" || query.StoreDist != "" {
		panic("GeoRadiusByMember does not support Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoRadiusByMemberStoreCompleted(key, member string, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(GEORADIUSBYMEMBER).Keys(key).Args(member)
	if query.Store == "" && query.StoreDist == "" {
		panic("GeoRadiusByMemberStore requires Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoRadiusCompleted(key string, longitude, latitude float64, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(GEORADIUS_RO).Keys(key).Args(strconv.FormatFloat(longitude, 'f', -1, 64), strconv.FormatFloat(latitude, 'f', -1, 64))
	if query.Store != "" || query.StoreDist != "" {
		panic("GeoRadius does not support Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoRadiusStoreCompleted(key string, longitude, latitude float64, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(GEORADIUS).Keys(key).Args(strconv.FormatFloat(longitude, 'f', -1, 64), strconv.FormatFloat(latitude, 'f', -1, 64))
	if query.Store == "" && query.StoreDist == "" {
		panic("GeoRadiusStore requires Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoSearchCompleted(key string, q GeoSearchQuery) Completed {
	return b.Arbitrary(GEOSEARCH).Keys(key).Args(geoSearchQueryArgs(q)...).Build()
}

func (b builder) GeoSearchStoreCompleted(src, dest string, q GeoSearchStoreQuery) Completed {
	cmd := b.Arbitrary(GEOSEARCHSTORE).Keys(dest, src)
	cmd = cmd.Args(geoSearchQueryArgs(q.GeoSearchQuery)...)
	if q.StoreDist {
		cmd = cmd.Args(STOREDIST)
	}
	return cmd.Build()
}

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
	cmd := b.Arbitrary(HSCAN).Keys(key).Args(strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, strconv.FormatInt(count, 10))
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

func (b builder) PFAddCompleted(key string, els ...any) Completed {
	return b.Pfadd().Key(key).Element(argsToSlice(els)...).Build()
}

func (b builder) PFCountCompleted(keys ...string) Completed {
	return b.Pfcount().Key(keys...).Build()
}

func (b builder) PFMergeCompleted(dest string, keys ...string) Completed {
	return b.Pfmerge().Destkey(dest).Sourcekey(keys...).Build()
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

func (b builder) LLenCompleted(key string) Completed {
	return b.Llen().Key(key).Build()
}

func (b builder) LMoveCompleted(source, destination, srcpos, destpos string) Completed {
	return b.Arbitrary(LMOVE).Keys(source, destination).Args(srcpos, destpos).Build()
}

func (b builder) LPopCompleted(key string) Completed {
	return b.Lpop().Key(key).Build()
}

func (b builder) LPopCountCompleted(key string, count int64) Completed {
	return b.Lpop().Key(key).Count(count).Build()
}

func (b builder) LMPopCompleted(direction string, count int64, keys ...string) Completed {
	cmd := b.Arbitrary(LMPOP, strconv.Itoa(len(keys))).Keys(keys...)
	cmd = cmd.Args(direction)
	if count > 0 {
		cmd = cmd.Args(COUNT, strconv.FormatInt(count, 10))
	}
	return cmd.Build()
}

func (b builder) LPosCompleted(key string, element string, a LPosArgs) Completed {
	cmd := b.Arbitrary(LPOS).Keys(key).Args(element)
	if a.Rank != 0 {
		cmd = cmd.Args(RANK, strconv.FormatInt(a.Rank, 10))
	}
	if a.MaxLen != 0 {
		cmd = cmd.Args(MAXLEN, strconv.FormatInt(a.MaxLen, 10))
	}
	return cmd.Build()
}

func (b builder) LPosCountCompleted(key string, element string, count int64, a LPosArgs) Completed {
	cmd := b.Arbitrary(LPOS).Keys(key).Args(element).Args(COUNT, strconv.FormatInt(count, 10))
	if a.Rank != 0 {
		cmd = cmd.Args(RANK, strconv.FormatInt(a.Rank, 10))
	}
	if a.MaxLen != 0 {
		cmd = cmd.Args(MAXLEN, strconv.FormatInt(a.MaxLen, 10))
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

func (b builder) TimeCompleted() Completed {
	return b.Time().Build()
}
func (b builder) SAddCompleted(key string, member any) Completed {
	return b.Sadd().Key(key).Member(str(member)).Build()
}

func (b builder) SMAddCompleted(key string, members ...any) Completed {
	cmd := b.Sadd().Key(key).Member()
	for _, m := range argsToSlice(members) {
		cmd = cmd.Member(str(m))
	}
	return cmd.Build()
}

func (b builder) SCardCompleted(key string) Completed {
	return b.Scard().Key(key).Build()
}

func (b builder) SDiffCompleted(keys ...string) Completed {
	return b.Sdiff().Key(keys...).Build()
}

func (b builder) SDiffStoreCompleted(destination string, keys ...string) Completed {
	return b.Sdiffstore().Destination(destination).Key(keys...).Build()
}

func (b builder) SInterCompleted(keys ...string) Completed {
	return b.Sinter().Key(keys...).Build()
}

func (b builder) SInterStoreCompleted(destination string, keys ...string) Completed {
	return b.Sinterstore().Destination(destination).Key(keys...).Build()
}

func (b builder) SInterCardCompleted(limit int64, keys ...string) Completed {
	return b.Sintercard().Numkeys(int64(len(keys))).Key(keys...).Limit(limit).Build()
}

func (b builder) SIsMemberCompleted(key string, member any) Completed {
	return b.Sismember().Key(key).Member(str(member)).Build()
}

func (b builder) SMIsMemberCompleted(key string, members ...any) Completed {
	return b.Smismember().Key(key).Member(argsToSlice(members)...).Build()
}

func (b builder) SMembersCompleted(key string) Completed {
	return b.Smembers().Key(key).Build()
}

func (b builder) SMoveCompleted(source, destination string, member any) Completed {
	return b.Smove().Source(source).Destination(destination).Member(str(member)).Build()
}

func (b builder) SPopCompleted(key string) Completed {
	return b.Spop().Key(key).Build()
}

func (b builder) SPopNCompleted(key string, count int64) Completed {
	return b.Spop().Key(key).Count(count).Build()
}

func (b builder) SRandMemberCompleted(key string) Completed {
	return b.Srandmember().Key(key).Build()
}

func (b builder) SRandMemberNCompleted(key string, count int64) Completed {
	return b.Srandmember().Key(key).Count(count).Build()
}

func (b builder) SRemCompleted(key string, member any) Completed {
	return b.Srem().Key(key).Member(str(member)).Build()
}

func (b builder) SMRemCompleted(key string, members ...any) Completed {
	return b.Srem().Key(key).Member(argsToSlice(members)...).Build()
}

func (b builder) SScanCompleted(key string, cursor uint64, match string, count int64) Completed {
	cmd := b.Arbitrary(SSCAN).Keys(key).Args(strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, strconv.FormatInt(count, 10))
	}
	return cmd.ReadOnly()
}

func (b builder) SUnionCompleted(keys ...string) Completed {
	return b.Sunion().Key(keys...).Build()
}

func (b builder) SUnionStoreCompleted(destination string, keys ...string) Completed {
	return b.Sunionstore().Destination(destination).Key(keys...).Build()
}

func (b builder) XAckCompleted(stream, group string, ids ...string) Completed {
	return b.Xack().Key(stream).Group(group).Id(ids...).Build()
}

func (b builder) XAddCompleted(a XAddArgs) Completed {
	cmd := b.Arbitrary(XADD).Keys(a.Stream)
	if a.NoMkStream {
		cmd = cmd.Args(NOMKSTREAM)
	}
	switch {
	case a.MaxLen > 0:
		if a.Approx {
			cmd = cmd.Args(MAXLEN, "~", strconv.FormatInt(a.MaxLen, 10))
		} else {
			cmd = cmd.Args(MAXLEN, strconv.FormatInt(a.MaxLen, 10))
		}
	case a.MinID != "":
		if a.Approx {
			cmd = cmd.Args(MINID, "~", a.MinID)
		} else {
			cmd = cmd.Args(MINID, a.MinID)
		}
	}
	if a.Limit > 0 {
		cmd = cmd.Args(LIMIT, strconv.FormatInt(a.Limit, 10))
	}
	if a.ID != "" {
		cmd = cmd.Args(a.ID)
	} else {
		cmd = cmd.Args("*")
	}
	cmd = cmd.Args(argToSlice(a.Values)...)
	return cmd.Build()
}

func (b builder) XAutoClaimCompleted(a XAutoClaimArgs) Completed {
	if a.Count > 0 {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Count(a.Count).Build()
	} else {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Build()
	}
}

func (b builder) XAutoClaimJustIDCompleted(a XAutoClaimArgs) Completed {
	if a.Count > 0 {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Count(a.Count).Justid().Build()
	} else {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Justid().Build()
	}
}

func (b builder) XClaimCompleted(a XClaimArgs) Completed {
	return b.Xclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Id(a.Messages...).Build()
}

func (b builder) XClaimJustIDCompleted(a XClaimArgs) Completed {
	return b.Xclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Id(a.Messages...).Justid().Build()
}

func (b builder) XDelCompleted(stream string, ids ...string) Completed {
	return b.Xdel().Key(stream).Id(ids...).Build()
}

func (b builder) XGroupCreateCompleted(stream, group, start string) Completed {
	return b.XgroupCreate().Key(stream).Group(group).Id(start).Build()
}

func (b builder) XGroupCreateMkStreamCompleted(stream, group, start string) Completed {
	return b.XgroupCreate().Key(stream).Group(group).Id(start).Mkstream().Build()
}

func (b builder) XGroupCreateConsumerCompleted(stream, group, consumer string) Completed {
	return b.XgroupCreateconsumer().Key(stream).Group(group).Consumer(consumer).Build()
}

func (b builder) XGroupDelConsumerCompleted(stream, group, consumer string) Completed {
	return b.XgroupDelconsumer().Key(stream).Group(group).Consumername(consumer).Build()
}

func (b builder) XGroupDestroyCompleted(stream, group string) Completed {
	return b.XgroupDestroy().Key(stream).Group(group).Build()
}

func (b builder) XGroupSetIDCompleted(stream, group, start string) Completed {
	return b.XgroupSetid().Key(stream).Group(group).Id(start).Build()
}

func (b builder) XInfoConsumersCompleted(key, group string) Completed {
	return b.XinfoConsumers().Key(key).Group(group).Build()
}

func (b builder) XInfoGroupsCompleted(key string) Completed {
	return b.XinfoGroups().Key(key).Build()
}

func (b builder) XInfoStreamCompleted(key string) Completed {
	return b.XinfoStream().Key(key).Build()
}

func (b builder) XInfoStreamFullCompleted(key string, count int64) Completed {
	return b.XinfoStream().Key(key).Full().Count(count).Build()
}

func (b builder) XLenCompleted(stream string) Completed {
	return b.Xlen().Key(stream).Build()
}

func (b builder) XPendingCompleted(stream, group string) Completed {
	return b.Xpending().Key(stream).Group(group).Build()
}

func (b builder) XPendingExtCompleted(a XPendingExtArgs) Completed {
	cmd := b.Arbitrary(XPENDING).Keys(a.Stream).Args(a.Group)
	if a.Idle != 0 {
		cmd = cmd.Args(IDLE, strconv.FormatInt(formatMs(a.Idle), 10))
	}
	cmd = cmd.Args(a.Start, a.End, strconv.FormatInt(a.Count, 10))
	if a.Consumer != "" {
		cmd = cmd.Args(a.Consumer)
	}
	return cmd.Build()
}

func (b builder) XRangeCompleted(stream, start, stop string) Completed {
	return b.Xrange().Key(stream).Start(start).End(stop).Build()
}

func (b builder) XRangeNCompleted(stream, start, stop string, count int64) Completed {
	return b.Xrange().Key(stream).Start(start).End(stop).Count(count).Build()
}

func (b builder) XRevRangeCompleted(stream, stop, start string) Completed {
	return b.Xrevrange().Key(stream).End(stop).Start(start).Build()
}

func (b builder) XRevRangeNCompleted(stream, stop, start string, count int64) Completed {
	return b.Xrevrange().Key(stream).End(stop).Start(start).Count(count).Build()
}

func (b builder) xTrim(key, strategy string,
	approx bool, threshold string, limit int64) Completed {
	cmd := b.Arbitrary(XTRIM).Keys(key).Args(strategy)
	if approx {
		cmd = cmd.Args("~")
	}
	cmd = cmd.Args(threshold)
	if limit > 0 {
		cmd = cmd.Args(LIMIT, strconv.FormatInt(limit, 10))
	}
	return cmd.Build()
}

func (b builder) XTrimCompleted(key string, maxLen int64) Completed {
	return b.xTrim(key, MAXLEN, false, strconv.FormatInt(maxLen, 10), 0)
}

func (b builder) XTrimMaxLenApproxCompleted(key string, maxLen, limit int64) Completed {
	return b.xTrim(key, MAXLEN, true, strconv.FormatInt(maxLen, 10), limit)
}

func (b builder) XTrimMinIDCompleted(key string, minID string) Completed {
	return b.xTrim(key, MINID, false, minID, 0)
}

func (b builder) XTrimMinIDApproxCompleted(key string, minID string, limit int64) Completed {
	return b.xTrim(key, MINID, true, minID, limit)
}

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
	cmd := b.Arbitrary(SET).Keys(key).Args(str(value))
	if a.KeepTTL {
		cmd = cmd.Args(KEEPTTL)
	}
	if !a.ExpireAt.IsZero() {
		cmd = cmd.Args(EXAT, strconv.FormatInt(a.ExpireAt.Unix(), 10))
	}
	if a.TTL > 0 {
		if usePrecise(a.TTL) {
			cmd = cmd.Args(PX, strconv.FormatInt(formatMs(a.TTL), 10))
		} else {
			cmd = cmd.Args(EX, strconv.FormatInt(formatSec(a.TTL), 10))
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
		cmd = cmd.Args(GET)
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

func (b builder) zAddArgs(key string, incr bool, args ZAddArgs) Completed {
	cmd := b.Arbitrary(ZADD).Keys(key)
	// The GT, LT and NX options are mutually exclusive.
	if args.NX {
		cmd = cmd.Args(NX)
	} else {
		if args.XX {
			cmd = cmd.Args(XX)
		}
		if args.GT {
			cmd = cmd.Args(GT)
		} else if args.LT {
			cmd = cmd.Args(LT)
		}
	}
	if args.Ch {
		cmd = cmd.Args(CH)
	}
	if incr {
		cmd = cmd.Args(INCR)
	}
	for _, v := range args.Members {
		cmd = cmd.Args(strconv.FormatFloat(v.Score, 'f', -1, 64), v.Member)
	}
	return cmd.Build()
}

func (b builder) ZAddCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members})
}

func (b builder) ZAddNXCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, NX: true})
}

func (b builder) ZAddXXCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, XX: true})
}

func (b builder) ZAddLTCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, LT: true})
}

func (b builder) ZAddGTCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, GT: true})
}

func (b builder) ZAddChCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, Ch: true})
}

func (b builder) ZAddArgsCompleted(key string, args ZAddArgs) Completed {
	return b.zAddArgs(key, false, args)
}

func (b builder) ZAddArgsIncrCompleted(key string, args ZAddArgs) Completed {
	return b.zAddArgs(key, true, args)
}

func (b builder) ZCardCompleted(key string) Completed {
	return b.Zcard().Key(key).Build()
}

func (b builder) ZCountCompleted(key, min, max string) Completed {
	return b.Zcount().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZLexCountCompleted(key, min, max string) Completed {
	return b.Zlexcount().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZIncrByCompleted(key string, increment float64, member string) Completed {
	return b.Zincrby().Key(key).Increment(increment).Member(member).Build()
}

func (b builder) zInter(store ZStore, withScores bool) Completed {
	cmd := b.Arbitrary(ZINTER).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(WEIGHTS)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(AGGREGATE, store.Aggregate)
	}
	if withScores {
		cmd = cmd.Args(WITHSCORES)
	}
	return cmd.ReadOnly()
}

func (b builder) ZInterCompleted(store ZStore) Completed           { return b.zInter(store, false) }
func (b builder) ZInterWithScoresCompleted(store ZStore) Completed { return b.zInter(store, true) }

func (b builder) ZInterCardCompleted(limit int64, keys ...string) Completed {
	return b.Zintercard().Numkeys(int64(len(keys))).Key(keys...).Limit(limit).Build()
}

func (b builder) ZInterStoreCompleted(destination string, store ZStore) Completed {
	cmd := b.Arbitrary(ZINTERSTORE).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(WEIGHTS)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(AGGREGATE, store.Aggregate)
	}
	return cmd.Keys(destination).Build()
}

func (b builder) ZMPopCompleted(order string, count int64, keys ...string) Completed {
	cmd := b.Arbitrary(ZMPOP, strconv.Itoa(len(keys))).Keys(keys...)
	cmd = cmd.Args(order)
	if count > 0 {
		cmd = cmd.Args(COUNT, strconv.FormatInt(count, 10))
	}
	return cmd.Build()
}

func (b builder) ZMScoreCompleted(key string, members ...string) Completed {
	return b.Zmscore().Key(key).Member(members...).Build()
}

func (b builder) ZPopMaxCompleted(key string, count ...int64) Completed {
	switch len(count) {
	case 0:
		return b.Zpopmax().Key(key).Build()
	case 1:
		return b.Zpopmax().Key(key).Count(count[0]).Build()
	default:
		panic("too many arguments")
	}
}

func (b builder) ZPopMinCompleted(key string, count ...int64) Completed {
	switch len(count) {
	case 0:
		return b.Zpopmin().Key(key).Build()
	case 1:
		return b.Zpopmin().Key(key).Count(count[0]).Build()
	default:
		panic("too many arguments")
	}
}

func (b builder) zRangeArgs(withScores bool, z ZRangeArgs) Completed {
	cmd := b.Arbitrary(ZRANGE).Keys(z.Key)
	if z.Rev && (z.ByScore || z.ByLex) {
		cmd = cmd.Args(str(z.Stop), str(z.Start))
	} else {
		cmd = cmd.Args(str(z.Start), str(z.Stop))
	}
	if z.ByScore {
		cmd = cmd.Args(BYSCORE)
	} else if z.ByLex {
		cmd = cmd.Args(BYLEX)
	}
	if z.Rev {
		cmd = cmd.Args(REV)
	}
	if z.Offset != 0 || z.Count != 0 {
		cmd = cmd.Args(LIMIT, strconv.FormatInt(z.Offset, 10), strconv.FormatInt(z.Count, 10))
	}
	if withScores {
		cmd = cmd.Args(WITHSCORES)
	}
	return cmd.Build()
}

func (b builder) ZRangeCompleted(key string, start, stop int64) Completed {
	return b.zRangeArgs(false, ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  stop,
	})
}

func (b builder) ZRangeWithScoresCompleted(key string, start, stop int64) Completed {
	return b.zRangeArgs(true, ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  stop,
	})
}

func (b builder) ZRangeByScoreCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Build()
	}
}

func (b builder) ZRangeByLexCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrangebylex().Key(key).Min(opt.Min).Max(opt.Max).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrangebylex().Key(key).Min(opt.Min).Max(opt.Max).Build()
	}
}

func (b builder) ZRangeByScoreWithScoresCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Withscores().Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Withscores().Build()
	}
}

func (b builder) ZRangeArgsCompleted(z ZRangeArgs) Completed {
	return b.zRangeArgs(false, z)
}

func (b builder) ZRangeArgsWithScoresCompleted(z ZRangeArgs) Completed {
	return b.zRangeArgs(true, z)
}

func (b builder) ZRangeStoreCompleted(dst string, z ZRangeArgs) Completed {
	cmd := b.Arbitrary(ZRANGESTORE).Keys(dst, z.Key)
	if z.Rev && (z.ByScore || z.ByLex) {
		cmd = cmd.Args(str(z.Stop), str(z.Start))
	} else {
		cmd = cmd.Args(str(z.Start), str(z.Stop))
	}
	if z.ByScore {
		cmd = cmd.Args(BYSCORE)
	} else if z.ByLex {
		cmd = cmd.Args(BYLEX)
	}
	if z.Rev {
		cmd = cmd.Args(REV)
	}
	if z.Offset != 0 || z.Count != 0 {
		cmd = cmd.Args(LIMIT, strconv.FormatInt(z.Offset, 10), strconv.FormatInt(z.Count, 10))
	}
	return cmd.Build()
}

func (b builder) ZRankCompleted(key, member string) Completed {
	return b.Zrank().Key(key).Member(member).Build()
}

func (b builder) ZRankWithScoreCompleted(key, member string) Completed {
	return b.Zrank().Key(key).Member(member).Withscore().Build()
}

func (b builder) ZRemCompleted(key string, members ...any) Completed {
	return b.Zrem().Key(key).Member(argsToSlice(members)...).Build()
}

func (b builder) ZRemRangeByRankCompleted(key string, start, stop int64) Completed {
	return b.Zremrangebyrank().Key(key).Start(start).Stop(stop).Build()
}
func (b builder) ZRemRangeByScoreCompleted(key, min, max string) Completed {
	return b.Zremrangebyscore().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZRemRangeByLexCompleted(key string, min, max string) Completed {
	return b.Zremrangebylex().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZRevRangeCompleted(key string, start, stop int64) Completed {
	return b.Zrevrange().Key(key).Start(start).Stop(stop).Build()
}

func (b builder) ZRevRangeWithScoresCompleted(key string, start, stop int64) Completed {
	return b.Zrevrange().Key(key).Start(start).Stop(stop).Withscores().Build()
}

func (b builder) ZRevRangeByScoreCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Build()
	}
}

func (b builder) ZRevRangeByLexCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrevrangebylex().Key(key).Max(opt.Max).Min(opt.Min).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrevrangebylex().Key(key).Max(opt.Max).Min(opt.Min).Build()
	}
}

func (b builder) ZRevRangeByScoreWithScoresCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Withscores().Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Withscores().Build()
	}
}

func (b builder) ZRevRankCompleted(key, member string) Completed {
	return b.Zrevrank().Key(key).Member(member).Build()
}

func (b builder) ZRevRankWithScoreCompleted(key, member string) Completed {
	return b.Zrevrank().Key(key).Member(member).Withscore().Build()
}

func (b builder) ZScoreCompleted(key, member string) Completed {
	return b.Zscore().Key(key).Member(member).Build()
}

func (b builder) zUnion(store ZStore, withScores bool) Completed {
	cmd := b.Arbitrary(ZUNION).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(WEIGHTS)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(AGGREGATE, store.Aggregate)
	}
	if withScores {
		cmd = cmd.Args(WITHSCORES)
	}
	return cmd.ReadOnly()
}

func (b builder) ZUnionStoreCompleted(dest string, store ZStore) Completed {
	cmd := b.Arbitrary(ZUNIONSTORE).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(WEIGHTS)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(AGGREGATE, store.Aggregate)
	}
	return cmd.Keys(dest).Build()
}

func (b builder) ZUnionCompleted(store ZStore) Completed           { return b.zUnion(store, false) }
func (b builder) ZUnionWithScoresCompleted(store ZStore) Completed { return b.zUnion(store, true) }

func (b builder) ZRandMemberCompleted(key string, count int64) Completed {
	return b.Zrandmember().Key(key).Count(count).Build()
}

func (b builder) ZRandMemberWithScoresCompleted(key string, count int64) Completed {
	return b.Zrandmember().Key(key).Count(count).Withscores().Build()
}

func (b builder) ZDiffCompleted(keys ...string) Completed {
	return b.Zdiff().Numkeys(int64(len(keys))).Key(keys...).Build()
}

func (b builder) ZDiffWithScoresCompleted(keys ...string) Completed {
	return b.Zdiff().Numkeys(int64(len(keys))).Key(keys...).Withscores().Build()
}

func (b builder) ZDiffStoreCompleted(destination string, keys ...string) Completed {
	return b.Zdiffstore().Destination(destination).Numkeys(int64(len(keys))).Key(keys...).Build()
}

func (b builder) ZScanCompleted(key string, cursor uint64, match string, count int64) Completed {
	cmd := b.Arbitrary(ZSCAN).Keys(key).Args(strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, strconv.FormatInt(count, 10))
	}
	return cmd.ReadOnly()
}
