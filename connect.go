package redisson

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/modern-go/reflect2"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
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
	var isCluster bool
	if len(match) < 1 || len(strings.TrimSpace(match[0][1])) == 0 || strings.TrimSpace(match[0][1]) == "0" {
		isCluster = false
	} else {
		isCluster = true
	}
	c.isCluster.Store(isCluster)
	c.handler.setIsCluster(isCluster)
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
	v, err := newSemVersion(strings.TrimSpace(match[0][1]))
	if err != nil {
		return err
	}
	c.version.Store(&v)
	c.handler.setVersion(&v)
	return nil
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
	errMsgClusterInfoElements = "elements in cluster info address, expected 2 or 3"
	errMsgUnsupportedHello    = "unsupported command `hello`"
	errMsgSlotHasNoRedisNode  = "the slot has no redis node"
)

// clientCloseWaitTimeout client.Close 等所有 delayQueue worker 收尾的兜底超时。
// 比 q.Close 的 5s timeout (delayCloseWorkerWaitTimeout) 略长,给 q.Close 自身的
// in-flight ack 留出余量;反模式 (callback 内同时阻塞用户 chan + 调 q.Close) 下,
// 这个超时打破死锁,后续 c.cmd.Close() 让 worker 中正在执行的 Redis 调用快速失败。
const clientCloseWaitTimeout = 10 * time.Second

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

// reconnectWhenError 尝试根据 err 调整 Conf 后 reconnect。
// 第二个返回值 recognized 标识 err 是否被任一 matcher 识别并触发了重连：
//   - recognized=false：err 不是已知可恢复的错误；调用方应直接结束重试循环。
//   - recognized=true ：触发了一次 Close+connect；返回的 err 为新一轮 connect 的结果。
func (c *client) reconnectWhenError(err error) (error, bool) {
	if err == nil {
		return nil, false
	}
	for _, f := range reconnectErrors {
		if ok := f(c, err); ok {
			warning(fmt.Sprintf("%s, reconnect...", err.Error()))
			_ = c.Close()
			return c.connect(), true
		}
	}
	return err, false
}

// Version 返回当前 Redis server 版本。Connect 之前调用返回 nil。
func (c *client) Version() *semver.Version { return c.version.Load() }

func (c *client) Close() error {
	// 第一阶段:关掉所有 queue 的 ticker (q.Close 不等 worker, 仅阻止新 worker spawn)。
	c.delayQueues.Range(func(_, value any) bool {
		_ = value.(*delayQueue).Close()
		return true
	})
	// 第二阶段:带超时等所有在途 worker 收尾。
	// 反模式 (callback 内调 q.Close 且后续阻塞在用户 chan) 下 worker 永等 callback,
	// 这里超时打破死锁。
	waitWGWithTimeout(&c.delayWorkerWG, clientCloseWaitTimeout)
	c.delayQueues = sync.Map{}
	// 从全局 collector 摘除自己,否则长期运行 (频繁 New/Close) 会泄漏 client root 与
	// associated handler/metrics 引用,且 Prometheus scrape 仍会枚举到已 Close 的实例。
	col.cs.Delete(c)
	// 第三阶段:关闭 rueidis client。
	// 注意不再把 c.cmd / c.adapter 设为 nil,避免 worker 超时漏出后的 ackSuccess /
	// ackRetry 等调用对 nil c.cmd 直接 panic;rueidis client 在 Close 后调命令会
	// 返回 ErrClosing 而非 panic,worker 内有 errors.Is(err, context.Canceled) 等
	// silent 路径,这样配合可让漏出的 worker 优雅收尾。
	if c.cmd != nil && !reflect2.IsNil(c.cmd) {
		c.cmd.Close()
	}
	return nil
}

func Connect(v ConfInterface) (Cmdable, error) {
	revise(v)
	c := &client{v: v, handler: newBaseHandler(v), maxp: runtime.GOMAXPROCS(0)}
	err := c.connect()
	// 重连尝试上限：每个 matcher 最多触发一次 ApplyOption + reconnect，
	// 上限设为 len(reconnectErrors) 以容许一次连接错误链经过所有识别器。
	// 直到 err 不再被任何 matcher 识别（recognized=false），跳出循环。
	for attempt := 0; err != nil && attempt < len(reconnectErrors); attempt++ {
		var recognized bool
		err, recognized = c.reconnectWhenError(err)
		if !recognized {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	c.handler.setSilentErrCallback(func(err error) bool { return errors.Is(err, Nil) })
	return c, nil
}
