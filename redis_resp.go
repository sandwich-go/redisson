package redisson

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-semver/semver"
	goredis "github.com/go-redis/redis/v8"
	"regexp"
	"strings"
	"sync"
	"time"
)

type RESP int

const (
	RESP2 RESP = 0
	RESP3 RESP = 1
)

var (
	errTooManyArguments                    = errors.New("too many arguments")
	errGeoRadiusByMemberNotSupportStore    = errors.New("GeoRadiusByMember does not support Store or StoreDist")
	errGeoRadiusNotSupportStore            = errors.New("GeoRadius does not support Store or StoreDist")
	errGeoRadiusStoreRequiresStore         = errors.New("GeoRadiusStore requires Store or StoreDist")
	errGeoRadiusByMemberStoreRequiresStore = errors.New("GeoRadiusByMemberStore requires Store or StoreDist")
	errMemoryUsageArgsCount                = errors.New("MemoryUsage expects single sample count")
)

const Nil = goredis.Nil

func IsNil(err error) bool { return err == Nil }

type client struct {
	v            ConfVisitor
	cmdable      Cmdable
	cacheCmdable CacheCmdable
	handler      handler
	version      semver.Version
	once         sync.Once
}

var versionRE = regexp.MustCompile(`redis_version:(.+)`)

func (c *client) initVersion() (err error) {
	var res string
	res, err = c.cmdable.Info(context.Background(), SERVER).Result()
	if err != nil {
		return
	}
	match := versionRE.FindAllStringSubmatch(res, -1)
	if len(match) < 1 {
		err = fmt.Errorf("could not extract redis version")
		return
	}
	c.version, err = newSemVersion(strings.TrimSpace(match[0][1]))
	if err != nil {
		return
	}
	c.handler.setVersion(&c.version)
	return err
}

func MustNewClient(v ConfVisitor) Cmdable {
	cmd, err := Connect(v)
	if err != nil {
		panic(err)
	}
	return cmd
}

func Connect(v ConfVisitor) (Cmdable, error) {
	var err error
	c := &client{v: v, handler: newBaseHandler(v)}
	switch v.GetResp() {
	case RESP2:
		c.cmdable, err = connectResp2(v, c.handler)
	case RESP3:
		c.cmdable, err = connectResp3(v, c.handler)
	default:
		err = fmt.Errorf("unknown RESP version, %d", v.GetResp())
	}
	if err != nil {
		return nil, err
	}
	// 初始化版本号
	if err = c.initVersion(); err != nil {
		return nil, err
	}
	c.cacheCmdable = c.cmdable
	c.handler.setSilentErrCallback(func(err error) bool { return err == Nil })
	return c, nil
}

func (c *client) copy() *client {
	return &client{
		v:            c.v,
		cmdable:      c.cmdable,
		cacheCmdable: c.cacheCmdable,
		handler:      c.handler,
		version:      c.version,
	}
}
func (c *client) Cache(ttl time.Duration) CacheCmdable {
	cp := c.copy()
	cp.cacheCmdable = c.cmdable.Cache(ttl)
	return cp
}
func (c *client) PoolStats() PoolStats { return c.cmdable.PoolStats() }
func (c *client) Close() error         { return c.cmdable.Close() }
