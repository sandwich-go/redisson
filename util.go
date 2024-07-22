package redisson

import (
	"encoding"
	"fmt"
	"github.com/redis/rueidis"
	"strconv"
	"time"
)

const (
	OK                   = "OK"
	LIMIT                = "LIMIT"
	GET                  = "GET"
	STORE                = "STORE"
	COUNT                = "COUNT"
	M                    = "M"
	MI                   = "MI"
	FT                   = "FT"
	KM                   = "KM"
	EMPTY                = ""
	SCAN                 = "SCAN"
	SET                  = "SET"
	KEEPTTL              = "KEEPTTL"
	EXAT                 = "EXAT"
	PX                   = "PX"
	EX                   = "EX"
	BITFIELD             = "BITFIELD"
	GEORADIUS            = "GEORADIUS"
	GEORADIUS_RO         = "GEORADIUS_RO"
	GEORADIUSBYMEMBER    = "GEORADIUSBYMEMBER"
	GEORADIUSBYMEMBER_RO = "GEORADIUSBYMEMBER_RO"
	GEOSEARCH            = "GEOSEARCH"
	GEOSEARCHSTORE       = "GEOSEARCHSTORE"
	STOREDIST            = "STOREDIST"
	BYRADIUS             = "BYRADIUS"
	FROMMEMBER           = "FROMMEMBER"
	FROMLONLAT           = "FROMLONLAT"
	BYBOX                = "BYBOX"
	ANY                  = "ANY"
	WITHCOORD            = "WITHCOORD"
	WITHDIST             = "WITHDIST"
	WITHHASH             = "WITHHASH"
	LPOS                 = "LPOS"
	RANK                 = "RANK"
	MAXLEN               = "MAXLEN"
	MINID                = "MINID"
	NOMKSTREAM           = "NOMKSTREAM"
	XADD                 = "XADD"
	XX                   = "XX"
	NX                   = "NX"
	WITHSCORES           = "WITHSCORES"
	BYSCORE              = "BYSCORE"
	BYLEX                = "BYLEX"
	REV                  = "REV"
	ZRANGE               = "ZRANGE"
	SERVER               = "SERVER"
	CLUSTER              = "CLUSTER"
	LADDR                = "LADDR"
	BitCountIndexByte    = "BYTE"
	BitCountIndexBit     = "BIT"
	TYPE                 = "TYPE"
	HSCAN                = "HSCAN"
	MATCH                = "MATCH"
	BEFORE               = "BEFORE"
	AFTER                = "AFTER"
	LMPOP                = "LMPOP"
	LMOVE                = "LMOVE"
	SSCAN                = "SSCAN"
	XPENDING             = "XPENDING"
	IDLE                 = "IDLE"
	XTRIM                = "XTRIM"
	ZADD                 = "ZADD"
	GT                   = "GT"
	LT                   = "LT"
	CH                   = "CH"
	INCR                 = "INCR"
	ZRANGESTORE          = "ZRANGESTORE"
	ZINTER               = "ZINTER"
	ZINTERSTORE          = "ZINTERSTORE"
	ZUNION               = "ZUNION"
	ZUNIONSTORE          = "ZUNIONSTORE"
	WEIGHTS              = "WEIGHTS"
	AGGREGATE            = "AGGREGATE"
	ZMPOP                = "ZMPOP"
	ZSCAN                = "ZSCAN"
	FUNCTION             = "FUNCTION"
	LIST                 = "LIST"
	LIBRARYNAME          = "LIBRARYNAME"
	WITHCODE             = "WITHCODE"
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
