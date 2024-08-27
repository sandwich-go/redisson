package redisson

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
	"net"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	versionRE      = regexp.MustCompile(`redis_version:(.+)`)
	clusterEnabled = regexp.MustCompile(`cluster_enabled:(.+)`)
)

func (c *client) reviseCluster(ctx context.Context, info string) (err error) {
	if len(info) == 0 {
		info, err = c.Info(ctx, XXX_CLUSTER).Result()
		if err != nil {
			return
		}
	}
	match := clusterEnabled.FindAllStringSubmatch(info, -1)
	if len(match) < 1 || len(strings.TrimSpace(match[0][1])) == 0 || strings.TrimSpace(match[0][1]) == "0" {
		c.isCluster = false
	} else {
		c.isCluster = true
	}
	c.handler.setIsCluster(c.isCluster)
	return
}

func (c *client) reviseVersion(ctx context.Context, info string) (err error) {
	if len(info) == 0 {
		info, err = c.Info(ctx, XXX_SERVER).Result()
		if err != nil {
			return err
		}
	}
	match := versionRE.FindAllStringSubmatch(info, -1)
	if len(match) < 1 {
		err = fmt.Errorf("could not extract redis server version")
		return
	}
	c.version, err = newSemVersion(strings.TrimSpace(match[0][1]))
	if err != nil {
		return
	}
	c.handler.setVersion(&c.version)
	return err
}

func (c *client) revise(ctx context.Context) error {
	info, err := c.Info(ctx, XXX_CLUSTER, XXX_SERVER).Result()
	if err != nil {
		info = ""
	}
	if err = c.reviseVersion(ctx, info); err != nil {
		return err
	}
	if err = c.reviseCluster(ctx, info); err != nil {
		return err
	}
	return nil
}

func confVisitor2ClientOption(v ConfVisitor) rueidis.ClientOption {
	opt := rueidis.ClientOption{
		Username:          v.GetUsername(),
		Password:          v.GetPassword(),
		InitAddress:       v.GetAddrs(),
		SelectDB:          v.GetDB(),
		CacheSizeEachConn: v.GetCacheSizeEachConn(),
		RingScaleEachConn: v.GetRingScaleEachConn(),
		BlockingPoolSize:  v.GetConnPoolSize(),
		ConnWriteTimeout:  v.GetWriteTimeout(),
		DisableCache:      !v.GetEnableCache(),
		ShuffleInit:       true,
		AlwaysRESP2:       v.GetAlwaysRESP2(),
		ForceSingleClient: v.GetForceSingleClient(),
		Sentinel: rueidis.SentinelOption{
			Username:   v.GetUsername(),
			Password:   v.GetPassword(),
			ClientName: v.GetName(),
			MasterSet:  v.GetMasterName(),
		},
	}
	switch strings.ToLower(v.GetNet()) {
	case "unix":
		opt.DialFn = func(s string, dialer *net.Dialer, _ *tls.Config) (net.Conn, error) {
			return dialer.Dial("unix", s)
		}
	}
	return opt
}

func (c *client) connect() error {
	if t := c.v.GetT(); t != nil {
		_ = c.v.ApplyOption(WithAddrs(miniredis.RunT(t).Addr()))
	}
	var err error
	c.cmd, err = rueidis.NewClient(confVisitor2ClientOption(c.v))
	if err != nil {
		return err
	}
	c.adapter = rueidiscompat.NewAdapter(c.cmd)
	if t := c.v.GetT(); t == nil {
		if err = c.revise(context.Background()); err != nil {
			_ = c.Close()
			return err
		}
	}
	c.builder = builder{c.cmd.B()}
	return nil
}

var reconnectErrors = []func(*client, string) bool{
	func(c *client, errString string) bool {
		if c.v.GetEnableCache() && strings.Contains(errString, rueidis.ErrNoCache.Error()) {
			c.v.ApplyOption(WithEnableCache(false))
			return true
		}
		return false
	},
	func(c *client, errString string) bool {
		if !c.v.GetAlwaysRESP2() && strings.Contains(errString, "elements in cluster info address, expected 2 or 3") || strings.Contains(errString, "unsupported command `hello`") {
			c.v.ApplyOption(WithAlwaysRESP2(true))
			return true
		}
		return false
	},
	func(c *client, errString string) bool {
		if !c.v.GetForceSingleClient() && strings.Contains(errString, "the slot has no redis node") {
			c.v.ApplyOption(WithForceSingleClient(true))
			return true
		}
		return false
	},
}

func (c *client) reconnectWhenError(err error) error {
	if err == nil {
		return nil
	}
	errString := err.Error()
	for _, f := range reconnectErrors {
		if ok := f(c, errString); ok {
			warning(fmt.Sprintf("%s, reconnect...", errString))
			_ = c.Close()
			return c.connect()
		}
	}
	return err
}

func (c *client) Close() error {
	c.delayQueues.Range(func(key, value any) bool {
		_ = value.(*delayQueue).Close()
		return true
	})
	c.delayQueues = sync.Map{}
	if c.cmd != nil {
		c.cmd.Close()
	}
	c.cmd = nil
	c.adapter = nil
	return nil
}

func Connect(v ConfInterface) (Cmdable, error) {
	c := &client{v: v, handler: newBaseHandler(v), maxp: runtime.GOMAXPROCS(0)}
	err := c.connect()
	if err != nil {
		for i := 0; i < len(reconnectErrors); i++ {
			err = c.reconnectWhenError(err)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}
	c.handler.setSilentErrCallback(func(err error) bool { return errors.Is(err, Nil) })
	return c, nil
}
