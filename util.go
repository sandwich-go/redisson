package redisson

import (
	"encoding"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/redis/rueidis"
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
	XXX_BY                   = "BY"
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
	XXX_ASC                  = "ASC"
	XXX_DESC                 = "DESC"
	XXX_ALPHA                = "ALPHA"
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

func e(msg string) {
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

func worker[V any](wg *sync.WaitGroup, ch chan V, fn func(V)) {
	for v := range ch {
		fn(v)
	}
	wg.Done()
}

func closeThenParallel[V any](maxp int, ch chan V, fn func(V)) {
	close(ch)
	concurrency := len(ch)
	if concurrency > maxp {
		concurrency = maxp
	}
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 1; i < concurrency; i++ {
		go worker(&wg, ch, fn)
	}
	worker(&wg, ch, fn)
	wg.Wait()
}

func parallelK[K comparable, V any](maxp int, p map[K]V, fn func(K)) {
	ch := make(chan K, len(p))
	for k := range p {
		ch <- k
	}
	closeThenParallel(maxp, ch, fn)
}

func atoi(b []byte) (int, error) {
	return strconv.Atoi(bytesToString(b))
}

func parseInt(b []byte, base int, bitSize int) (int64, error) {
	return strconv.ParseInt(bytesToString(b), base, bitSize)
}

func parseUint(b []byte, base int, bitSize int) (uint64, error) {
	return strconv.ParseUint(bytesToString(b), base, bitSize)
}

func parseFloat(b []byte, bitSize int) (float64, error) {
	return strconv.ParseFloat(bytesToString(b), bitSize)
}

// bytesToString converts byte slice to string.
func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// stringToBytes converts string to byte slice.
func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// Scan parses bytes `b` to `v` with appropriate type.
//
//nolint:gocyclo
func scan(b []byte, v interface{}) error {
	switch v := v.(type) {
	case nil:
		return fmt.Errorf("redis: Scan(nil)")
	case *string:
		*v = bytesToString(b)
		return nil
	case *[]byte:
		*v = b
		return nil
	case *int:
		var err error
		*v, err = atoi(b)
		return err
	case *int8:
		n, err := parseInt(b, 10, 8)
		if err != nil {
			return err
		}
		*v = int8(n)
		return nil
	case *int16:
		n, err := parseInt(b, 10, 16)
		if err != nil {
			return err
		}
		*v = int16(n)
		return nil
	case *int32:
		n, err := parseInt(b, 10, 32)
		if err != nil {
			return err
		}
		*v = int32(n)
		return nil
	case *int64:
		n, err := parseInt(b, 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *uint:
		n, err := parseUint(b, 10, 64)
		if err != nil {
			return err
		}
		*v = uint(n)
		return nil
	case *uint8:
		n, err := parseUint(b, 10, 8)
		if err != nil {
			return err
		}
		*v = uint8(n)
		return nil
	case *uint16:
		n, err := parseUint(b, 10, 16)
		if err != nil {
			return err
		}
		*v = uint16(n)
		return nil
	case *uint32:
		n, err := parseUint(b, 10, 32)
		if err != nil {
			return err
		}
		*v = uint32(n)
		return nil
	case *uint64:
		n, err := parseUint(b, 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *float32:
		n, err := parseFloat(b, 32)
		if err != nil {
			return err
		}
		*v = float32(n)
		return err
	case *float64:
		var err error
		*v, err = parseFloat(b, 64)
		return err
	case *bool:
		*v = len(b) == 1 && b[0] == '1'
		return nil
	case *time.Time:
		var err error
		*v, err = time.Parse(time.RFC3339Nano, bytesToString(b))
		return err
	case *time.Duration:
		n, err := parseInt(b, 10, 64)
		if err != nil {
			return err
		}
		*v = time.Duration(n)
		return nil
	case encoding.BinaryUnmarshaler:
		return v.UnmarshalBinary(b)
	case *net.IP:
		*v = b
		return nil
	default:
		return fmt.Errorf(
			"redis: can't unmarshal %T (consider implementing BinaryUnmarshaler)", v)
	}
}

func scanSlice(data []string, slice interface{}) error {
	v := reflect.ValueOf(slice)
	if !v.IsValid() {
		return fmt.Errorf("redis: ScanSlice(nil)")
	}
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("redis: ScanSlice(non-pointer %T)", slice)
	}
	v = v.Elem()
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("redis: ScanSlice(non-slice %T)", slice)
	}

	next := makeSliceNextElemFunc(v)
	for i, s := range data {
		elem := next()
		if err := scan([]byte(s), elem.Addr().Interface()); err != nil {
			err = fmt.Errorf("redis: ScanSlice index=%d value=%q failed: %w", i, s, err)
			return err
		}
	}

	return nil
}

func makeSliceNextElemFunc(v reflect.Value) func() reflect.Value {
	elemType := v.Type().Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
		return func() reflect.Value {
			if v.Len() < v.Cap() {
				v.Set(v.Slice(0, v.Len()+1))
				elem := v.Index(v.Len() - 1)
				if elem.IsNil() {
					elem.Set(reflect.New(elemType))
				}
				return elem.Elem()
			}

			elem := reflect.New(elemType)
			v.Set(reflect.Append(v, elem))
			return elem.Elem()
		}
	}

	zero := reflect.Zero(elemType)
	return func() reflect.Value {
		if v.Len() < v.Cap() {
			v.Set(v.Slice(0, v.Len()+1))
			return v.Index(v.Len() - 1)
		}

		v.Set(reflect.Append(v, zero))
		return v.Index(v.Len() - 1)
	}
}
