package redisson

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/modern-go/reflect2"
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
		info, err = c.Info(ctx, KwCluster).Result()
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
		info, err = c.Info(ctx, KwServer).Result()
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
	info, err := c.Info(ctx, KwCluster, KwServer).Result()
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
	// 当配置了 Tester 时，挂上 mock。该路径默认仅在 build tag `redisson_miniredis`
	// 下可用，避免下游用户的二进制链入 miniredis 及其传递依赖。
	if t := c.v.GetT(); t != nil {
		if err := setupMiniredisIfEnabled(c.v, t); err != nil {
			return err
		}
	}
	var err error
	c.cmd, err = rueidis.NewClient(confVisitor2ClientOption(c.v))
	if err != nil {
		return err
	}
	c.adapter = rueidiscompat.NewAdapter(c.cmd)
	c.builder = builder{c.cmd.B()}
	if t := c.v.GetT(); t == nil {
		// 启动期版本/集群探测使用 BootstrapTimeout，避免 Background 卡死初始化。
		// 默认 30s（参见 option.go 的 defaultBootstrapTimeout）。
		ctx, cancel := context.WithTimeout(context.Background(), c.v.GetBootstrapTimeout())
		defer cancel()
		if err = c.revise(ctx); err != nil {
			_ = c.Close()
			return err
		}
	}
	return nil
}

// reconnect 错误模式：因 rueidis 不暴露这些场景的强类型错误，仍需字符串匹配，
// 但集中定义在此，便于跟随上游 rueidis 改文案时修订。
const (
	errMsgClusterInfoElements   = "elements in cluster info address, expected 2 or 3"
	errMsgUnsupportedHello      = "unsupported command `hello`"
	errMsgSlotHasNoRedisNode    = "the slot has no redis node"
)

// reconnectErrors 是一组按顺序尝试的"识别 + 修正"函数。
// 每个函数:
//   - 接收 client 与原 error，
//   - 若识别命中则就地调整 ConfInterface 并返回 true（调用方会重连），
//   - 否则返回 false。
var reconnectErrors = []func(c *client, err error) bool{
	// rueidis 客户端缓存不可用 → 关闭 cache 后重连
	func(c *client, err error) bool {
		if c.v.GetEnableCache() && errors.Is(err, rueidis.ErrNoCache) {
			c.v.ApplyOption(WithEnableCache(false))
			return true
		}
		return false
	},
	// 服务端不支持 RESP3 / HELLO → 退回 RESP2
	func(c *client, err error) bool {
		if c.v.GetAlwaysRESP2() {
			return false
		}
		s := err.Error()
		if strings.Contains(s, errMsgClusterInfoElements) || strings.Contains(s, errMsgUnsupportedHello) {
			c.v.ApplyOption(WithAlwaysRESP2(true))
			return true
		}
		return false
	},
	// 集群拓扑探测失败 → 强制单连接
	func(c *client, err error) bool {
		if !c.v.GetForceSingleClient() && strings.Contains(err.Error(), errMsgSlotHasNoRedisNode) {
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
	for _, f := range reconnectErrors {
		if ok := f(c, err); ok {
			warning(fmt.Sprintf("%s, reconnect...", err.Error()))
			_ = c.Close()
			return c.connect()
		}
	}
	return err
}

func (c *client) Version() *semver.Version { return &c.version }

func (c *client) Close() error {
	c.delayQueues.Range(func(key, value any) bool {
		_ = value.(*delayQueue).Close()
		return true
	})
	c.delayQueues = sync.Map{}
	if c.cmd != nil && !reflect2.IsNil(c.cmd) {
		c.cmd.Close()
	}
	c.cmd = nil
	c.adapter = nil
	return nil
}

func Connect(v ConfInterface) (Cmdable, error) {
	revise(v)
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
