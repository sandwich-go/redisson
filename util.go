package redisson

import (
	"encoding"
	"fmt"
	"github.com/redis/rueidis"
	"strconv"
	"time"
)

const (
	XXX_BITFIELD             = "BITFIELD"
	XXX_SCAN                 = "SCAN"
	XXX_HSCAN                = "HSCAN"
	XXX_SSCAN                = "SSCAN"
	XXX_ZSCAN                = "ZSCAN"
	XXX_MATCH                = "MATCH"
	XXX_COUNT                = "COUNT"
	XXX_GEORADIUSBYMEMBER    = "GEORADIUSBYMEMBER"
	XXX_GEORADIUSBYMEMBER_RO = "GEORADIUSBYMEMBER_RO"
	XXX_GEORADIUS            = "GEORADIUS"
	XXX_GEORADIUS_RO         = "GEORADIUS_RO"
	XXX_GEOSEARCH            = "GEOSEARCH"
	XXX_GEOSEARCHSTORE       = "GEOSEARCHSTORE"
	XXX_STORE                = "STORE"
	XXX_STOREDIST            = "STOREDIST"
	XXX_LMOVE                = "LMOVE"
	XXX_LMPOP                = "LMPOP"
	XXX_LPOS                 = "LPOS"
	XXX_RANK                 = "RANK"
	XXX_MAXLEN               = "MAXLEN"
	XXX_MINID                = "MINID"
	XXX_FUNCTION             = "FUNCTION"
	XXX_LIST                 = "LIST"
	XXX_LIBRARYNAME          = "LIBRARYNAME"
	XXX_WITHCODE             = "WITHCODE"
	XXX_XADD                 = "XADD"
	XXX_NOMKSTREAM           = "NOMKSTREAM"
	XXX_LIMIT                = "LIMIT"
	XXX_XPENDING             = "XPENDING"
	XXX_IDLE                 = "IDLE"
	XXX_XTRIM                = "XTRIM"
	XXX_SET                  = "SET"
	XXX_KEEPTTL              = "KEEPTTL"
	XXX_EXAT                 = "EXAT"
	XXX_PX                   = "PX"
	XXX_EX                   = "EX"
	XXX_GET                  = "GET"
	XXX_ZADD                 = "ZADD"
	XXX_GT                   = "GT"
	XXX_LT                   = "LT"
	XXX_CH                   = "CH"
	XXX_INCR                 = "INCR"
	XXX_ZINTER               = "ZINTER"
	XXX_ZINTERSTORE          = "ZINTERSTORE"
	XXX_WEIGHTS              = "WEIGHTS"
	XXX_AGGREGATE            = "AGGREGATE"
	XXX_WITHSCORES           = "WITHSCORES"
	XXX_ZMPOP                = "ZMPOP"
	XXX_ZRANGE               = "ZRANGE"
	XXX_ZRANGESTORE          = "ZRANGESTORE"
	XXX_ZUNION               = "ZUNION"
	XXX_ZUNIONSTORE          = "ZUNIONSTORE"
	XXX_BYSCORE              = "BYSCORE"
	XXX_BYLEX                = "BYLEX"
	XXX_REV                  = "REV"
	XXX_BYRADIUS             = "BYRADIUS"
	XXX_BYBOX                = "BYBOX"
	XXX_FROMMEMBER           = "FROMMEMBER"
	XXX_FROMLONLAT           = "FROMLONLAT"
	XXX_WITHCOORD            = "WITHCOORD"
	XXX_WITHDIST             = "WITHDIST"
	XXX_WITHHASH             = "WITHHASH"
	XXX_ANY                  = "ANY"
	XXX_SERVER               = "SERVER"
	XXX_CLUSTER              = "CLUSTER"
)

var (
	nowFunc   = time.Now
	sinceFunc = time.Since
)

func usePrecise(dur time.Duration) bool {
	return dur < time.Second || dur%time.Second != 0
}

func formatSec(dur time.Duration) int64 {
	if dur > 0 && dur < time.Second {
		// too small, truncate too 1s
		return 1
	}
	return int64(dur / time.Second)
}

func formatMs(dur time.Duration) int64 {
	if dur > 0 && dur < time.Millisecond {
		// too small, truncate too 1ms
		return 1
	}
	return int64(dur / time.Millisecond)
}

func appendString(s string, ss ...string) []string {
	sss := make([]string, 0, len(ss)+1)
	sss = append(sss, s)
	sss = append(sss, ss...)
	return sss
}

func str(arg any) string {
	switch v := arg.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case uint64:
		return strconv.FormatUint(v, 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []byte:
		return string(v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case encoding.BinaryMarshaler:
		if data, err := v.MarshalBinary(); err == nil {
			return rueidis.BinaryString(data)
		}
	}
	return fmt.Sprint(arg)
}

func argsToSlice(src []any) []string {
	if len(src) == 1 {
		return argToSlice(src[0])
	}
	dst := make([]string, 0, len(src))
	for _, v := range src {
		dst = append(dst, str(v))
	}
	return dst
}

func argsToSliceWithValues(src []any) []string {
	if len(src) == 2 {
		return argToSlice(src[0])
	}
	dst := make([]string, 0, len(src)/2)
	for i := 0; i < len(src); i += 2 {
		dst = append(dst, str(src[i]))
	}
	return dst
}

func argToSlice(a any) []string {
	switch arg := a.(type) {
	case []string:
		return arg
	case []any:
		dst := make([]string, 0, len(arg))
		for _, v := range arg {
			dst = append(dst, str(v))
		}
		return dst
	case map[string]any:
		dst := make([]string, 0, len(arg))
		for k, v := range arg {
			dst = append(dst, k, str(v))
		}
		return dst
	case map[string]string:
		dst := make([]string, 0, len(arg))
		for k, v := range arg {
			dst = append(dst, k, v)
		}
		return dst
	default:
		return []string{str(arg)}
	}
}

func warning(msg string) {
	fmt.Println(msg)
}

func toFloat32(val any) (float32, error) {
	switch t := val.(type) {
	case int64:
		return float32(t), nil
	case string:
		f, err := strconv.ParseFloat(t, 32)
		if err != nil {
			return 0, err
		}
		return float32(f), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for Float32", t)
	}
}

func toFloat64(val any) (float64, error) {
	switch t := val.(type) {
	case int64:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(t, 64)
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for Float64", t)
	}
}
