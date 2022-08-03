package sandwich_redis

import (
	"fmt"
	"strconv"
	"time"
)

const (
	OK                   = "OK"
	BITFIELD             = "BITFIELD"
	AND                  = "AND"
	OR                   = "OR"
	XOR                  = "XOR"
	NOT                  = "NOT"
	CLIENT               = "CLIENT"
	KILL                 = "KILL"
	SORT                 = "SORT"
	BY                   = "BY"
	LIMIT                = "LIMIT"
	GET                  = "GET"
	ALPHA                = "ALPHA"
	STORE                = "STORE"
	TYPE                 = "TYPE"
	SCAN                 = "SCAN"
	SLAVEOF              = "SLAVEOF"
	MATCH                = "MATCH"
	COUNT                = "COUNT"
	M                    = "M"
	MI                   = "MI"
	FT                   = "FT"
	KM                   = "KM"
	EMPTY                = ""
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
	HMGET                = "hmget"
	HSCAN                = "HSCAN"
	BLMOVE               = "BLMOVE"
	LMOVE                = "LMOVE"
	BEFORE               = "BEFORE"
	AFTER                = "AFTER"
	LPOS                 = "LPOS"
	RANK                 = "RANK"
	MINID                = "MINID"
	MAXLEN               = "MAXLEN"
	SSCAN                = "SSCAN"
	ZADD                 = "ZADD"
	XX                   = "XX"
	NX                   = "NX"
	GT                   = "GT"
	LT                   = "LT"
	CH                   = "CH"
	INCR                 = "INCR"
	ZINTER               = "ZINTER"
	WEIGHTS              = "WEIGHTS"
	AGGREGATE            = "AGGREGATE"
	WITHSCORES           = "WITHSCORES"
	ZINTERSTORE          = "ZINTERSTORE"
	BYSCORE              = "BYSCORE"
	BYLEX                = "BYLEX"
	REV                  = "REV"
	ZRANGE               = "ZRANGE"
	ZRANGESTORE          = "ZRANGESTORE"
	ZSCAN                = "ZSCAN"
	ZUNION               = "ZUNION"
	ZUNIONSTORE          = "ZUNIONSTORE"
	XADD                 = "XADD"
	NOMKSTREAM           = "NOMKSTREAM"
	XPENDING             = "XPENDING"
	IDLE                 = "IDLE"
	XREAD                = "XREAD"
	BLOCK                = "BLOCK"
	STREAMS              = "STREAMS"
	XREADGROUP           = "XREADGROUP"
	GROUP                = "GROUP"
	NOACK                = "NOACK"
	XTRIM                = "XTRIM"
	SET                  = "SET"
	KEEPTTL              = "KEEPTTL"
	EXAT                 = "EXAT"
	PX                   = "PX"
	EX                   = "EX"
	SERVER               = "SERVER"
	LADDR                = "LADDR"
)

var (
	nowFunc   = time.Now
	sinceFunc = time.Since
)

func appendString(s string, ss ...string) []string {
	sss := make([]string, 0, len(ss)+1)
	sss = append(sss, s)
	sss = append(sss, ss...)
	return sss
}

func str(arg interface{}) string {
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
	}
	return fmt.Sprint(arg)
}

func argsToSliceWithValues(src []interface{}) []string {
	if len(src) == 2 {
		return argToSlice(src[0])
	}
	dst := make([]string, 0, len(src)/2)
	for i := 0; i < len(src); i += 2 {
		dst = append(dst, str(src[i]))
	}
	return dst
}

func argToSlice(a interface{}) []string {
	switch arg := a.(type) {
	case []string:
		return arg
	case []interface{}:
		dst := make([]string, 0, len(arg))
		for _, v := range arg {
			dst = append(dst, str(v))
		}
		return dst
	case map[string]interface{}:
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
