package redisson

import (
	"context"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
	"regexp"
	"strings"
)

var (
	versionRE      = regexp.MustCompile(`redis_version:(.+)`)
	clusterEnabled = regexp.MustCompile(`cluster_enabled:(.+)`)
)

func (c *client) reviseCluster(info string) error {
	match := clusterEnabled.FindAllStringSubmatch(info, -1)
	if len(match) < 1 || len(strings.TrimSpace(match[0][1])) == 0 || strings.TrimSpace(match[0][1]) == "0" {
		c.isCluster = false
	} else {
		c.isCluster = true
	}
	c.handler.setIsCluster(c.isCluster)
	return nil
}

func (c *client) reviseVersion(info string) (err error) {
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
	info, err := c.Info(ctx, CLUSTER, SERVER).Result()
	if err != nil {
		return err
	}
	if err = c.reviseVersion(info); err != nil {
		return err
	}
	if err = c.reviseCluster(info); err != nil {
		return err
	}
	return nil
}

func confVisitor2ClientOption(v ConfVisitor) rueidis.ClientOption {
	return rueidis.ClientOption{
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
	if err = c.revise(context.Background()); err != nil {
		_ = c.Close()
		return err
	}
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
		if strings.Contains(errString, "ERR This instance has cluster support disabled") || strings.Contains(errString, "ERR Cluster setting conflict") {
			c.v.ApplyOption(WithCluster(!c.v.GetCluster()))
			return true
		}
		return false
	},
	func(c *client, errString string) bool {
		if !c.v.GetAlwaysRESP2() && strings.Contains(errString, "elements in cluster info address, expected 2 or 3") {
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
	if c.cmd != nil {
		c.cmd.Close()
	}
	c.cmd = nil
	c.adapter = nil
	return nil
}

func Connect(v ConfInterface) (Cmdable, error) {
	c := &client{v: v, handler: newBaseHandler(v)}
	err := c.connect()
	if err == nil && c.isCluster != c.v.GetCluster() {
		err = fmt.Errorf("ERR Cluster setting conflict, server's cluster_enabled is %t, but client's cluster_enabled is %t", c.isCluster, c.v.GetCluster())
	}
	for i := 0; i < len(reconnectErrors); i++ {
		err = c.reconnectWhenError(err)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	c.handler.setSilentErrCallback(func(err error) bool { return errors.Is(err, Nil) })
	return c, nil
}
