package redisson

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/rueidis"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-semver/semver"
	goredis "github.com/go-redis/redis/v8"
)

type RESP = string

const (
	RESP2 RESP = "RESP2"
	RESP3 RESP = "RESP3"
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
	v            ConfInterface
	cmdable      Cmdable
	cacheCmdable CacheCmdable
	handler      handler
	version      semver.Version
	once         sync.Once
	isCluster    bool
}

var (
	versionRE      = regexp.MustCompile(`redis_version:(.+)`)
	clusterEnabled = regexp.MustCompile(`cluster_enabled:(.+)`)
)

func (c *client) clusterEnable() error {
	res, err := c.cmdable.Info(context.Background(), CLUSTER).Result()
	if err != nil {
		return err
	}
	match := clusterEnabled.FindAllStringSubmatch(res, -1)
	if len(match) < 1 || len(strings.TrimSpace(match[0][1])) == 0 || strings.TrimSpace(match[0][1]) == "0" {
		c.isCluster = false
	} else {
		c.isCluster = true
	}
	c.handler.setIsCluster(c.isCluster)
	return nil

}

func (c *client) Options() ConfVisitor { return c.v }
func (c *client) IsCluster() bool      { return c.isCluster }
func (c *client) ForEachNodes(ctx context.Context, f func(context.Context, Cmdable) error) error {
	if !c.isCluster {
		return f(ctx, c.cmdable)
	}
	return c.cmdable.ForEachNodes(ctx, f)
}

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

func MustNewClient(v ConfInterface) Cmdable {
	cmd, err := Connect(v)
	if err != nil {
		panic(err)
	}
	return cmd
}

func (c *client) initialize() error {
	// 初始化版本号
	if err := c.initVersion(); err != nil {
		return err
	}
	if err := c.clusterEnable(); err != nil {
		return err
	}
	return nil
}

func (c *client) connect() error {
	var err error
	if c.v.GetT() == nil {
		switch strings.ToUpper(c.v.GetResp()) {
		case RESP2:
			if c.v.GetAlwaysRESP2() {
				c.cmdable, err = connectResp3(c.v, c.handler)
			} else {
				c.cmdable, err = connectResp2(c.v, c.handler)
			}
		case RESP3:
			c.cmdable, err = connectResp3(c.v, c.handler)
		default:
			err = fmt.Errorf("unknown RESP version, %s", c.v.GetResp())
		}
		if err != nil {
			if strings.Contains(err.Error(), rueidis.ErrNoCache.Error()) {
				warning(fmt.Sprintf("%v, reconnect...", err))
				c.v.ApplyOption(WithEnableCache(false))
				return c.connect()
			}
			return err
		}
		// 初始化
		if err = c.initialize(); err != nil {
			_ = c.Close()
			return err
		}
	} else {
		c.cmdable, err = connectMock(c.v, c.handler)
	}
	return nil
}

func (c *client) reconnectWhenError(err error) error {
	if err == nil {
		return nil
	}
	errString := err.Error()
	if strings.Contains(errString, "ERR This instance has cluster support disabled") || strings.Contains(errString, "ERR Cluster setting conflict") {
		warning(fmt.Sprintf("%s, reconnect...", errString))
		c.v.ApplyOption(WithCluster(!c.v.GetCluster()))
		return c.connect()
	}
	if !c.v.GetAlwaysRESP2() && strings.Contains(errString, "elements in cluster info address, expected 2 or 3") {
		warning(fmt.Sprintf("%s, using always resp2, reconnect...", errString))
		c.v.ApplyOption(WithAlwaysRESP2(true))
		return c.connect()
	}
	if !c.v.GetForceSingleClient() && strings.Contains(errString, "the slot has no redis node") {
		warning(fmt.Sprintf("%s, using force single model, reconnect...", errString))
		c.v.ApplyOption(WithForceSingleClient(true))
		return c.connect()
	}
	return err
}

var retryTimes = 3

func Connect(v ConfInterface) (Cmdable, error) {
	c := &client{v: v, handler: newBaseHandler(v)}
	err := c.connect()
	if err == nil && c.isCluster != c.v.GetCluster() {
		err = fmt.Errorf("ERR Cluster setting conflict, server's cluster_enabled is %t, but client's cluster_enabled is %t", c.isCluster, c.v.GetCluster())
	}

	for i := 0; i < retryTimes; i++ {
		err = c.reconnectWhenError(err)
		if err == nil {
			break
		}
	}
	if err != nil {
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
func (c *client) PoolStats() PoolStats                           { return c.cmdable.PoolStats() }
func (c *client) Close() error                                   { return c.cmdable.Close() }
func (c *client) NewLocker(opts ...LockerOption) (Locker, error) { return c.cmdable.NewLocker(opts...) }
func (c *client) Receive(ctx context.Context, cb func(Message), channels ...string) error {
	return c.cmdable.Receive(ctx, cb, channels...)
}
func (c *client) PReceive(ctx context.Context, cb func(Message), patterns ...string) error {
	return c.cmdable.PReceive(ctx, cb, patterns...)
}
