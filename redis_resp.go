package redisson

import (
	"context"
	"fmt"
	"github.com/coreos/go-semver/semver"
	goredis "github.com/redis/go-redis/v9"
	"regexp"
	"strings"
	"sync"
	"time"
)

type RESP = string

const (
	RESP2 RESP = "RESP2"
	RESP3 RESP = "RESP3"
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
	return nil

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
	switch strings.ToUpper(c.v.GetResp()) {
	case RESP2:
		c.cmdable, err = connectResp2(c.v, c.handler)
	default:
		err = fmt.Errorf("unknown RESP version, %s", c.v.GetResp())
	}
	if err != nil {
		return err
	}
	// 初始化
	if err = c.initialize(); err != nil {
		_ = c.Close()
		return err
	}
	return nil
}

func (c *client) reconnectWhenError(err error) error {
	if err == nil {
		return nil
	}
	errString := err.Error()
	if strings.Contains(errString, "ERR This instance has cluster support disabled") ||
		strings.Contains(errString, "ERR Cluster setting conflict") {
		warning(fmt.Sprintf("%s, reconnect...", errString))
		c.v.ApplyOption(WithCluster(!c.v.GetCluster()))
		return c.connect()
	} else if c.v.GetResp() == RESP2 && strings.Contains(errString, "elements in cluster info address, expected 2 or 3") {
		warning(fmt.Sprintf("%s, using always resp2, reconnect...", errString))
		c.v.ApplyOption(WithResp(RESP3))
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
func (c *client) PoolStats() PoolStats { return c.cmdable.PoolStats() }
func (c *client) Close() error         { return c.cmdable.Close() }
