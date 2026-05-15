package redisson

import (
	"context"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"sync"
	"sync/atomic"
	"time"
)

type isSilentError func(error) bool

type handler interface {
	setVersion(*semver.Version)
	setIsCluster(bool)
	setSilentErrCallback(isSilentError)
	setRegisterCollector(RegisterCollectorFunc)

	before(ctx context.Context, command Command) context.Context
	beforeWithKeys(ctx context.Context, command Command, getKeys func() []string) context.Context
	after(ctx context.Context, err error)
	cache(ctx context.Context, hit bool)
	isCluster() bool
	delayPollError(name string)
	delayReclaimError(name string)
	delayReclaim(name string, count int)
}

func newSemVersion(version string) (semver.Version, error) {
	v := semver.Version{}
	if err := v.Set(version); err != nil {
		return v, err
	}
	return v, nil
}

func mustNewSemVersion(version string) semver.Version {
	v, err := newSemVersion(version)
	if err != nil {
		panic(err)
	}
	return v
}

// silentErrCallbackBox 给 atomic.Value 提供单一具体类型，规避存入不同
// 函数实现时触发 "store of inconsistently typed value" panic。
type silentErrCallbackBox struct{ f isSilentError }

type baseHandler struct {
	// silentErrCallback / version / cluster 三者构造期由 Connect 写一次，
	// 命令路径并发读；reconnectWhenError 路径下还会再写一次。
	// 用 atomic 包装，避免 plain field 在 reconnect 时与命令读形成 data race。
	silentErrCallback atomic.Value // silentErrCallbackBox
	version           atomic.Pointer[semver.Version]
	cluster           atomic.Bool
	v                 ConfVisitor
	// metrics 是构造期注入的 prometheus 指标集合；
	// 默认指向 defaultMetrics（向后兼容），调用 newBaseHandlerWithMetrics 可注入独立实例。
	metrics *metricsSet

	mx                 sync.Mutex
	warningOnceMapping map[string]struct{}
}

func newBaseHandler(v ConfVisitor) handler {
	return newBaseHandlerWithMetrics(v, defaultMetrics)
}

// newBaseHandlerWithMetrics 用指定的 metrics 实例构造 handler。
// 多 client 隔离场景下使用 newMetricsSet() 生成独立实例并通过本函数注入。
func newBaseHandlerWithMetrics(v ConfVisitor, m *metricsSet) handler {
	if m == nil {
		m = defaultMetrics
	}
	return &baseHandler{
		v:                  v,
		metrics:            m,
		warningOnceMapping: make(map[string]struct{}),
	}
}

type (
	startTimeContextKeyType  struct{}
	commandContextKeyType    struct{}
	subCommandContextKeyType struct{}
	skipCheckContextKeyType  struct{}
)

func (*startTimeContextKeyType) String() string  { return "start_time" }
func (*commandContextKeyType) String() string    { return "command" }
func (*subCommandContextKeyType) String() string { return "sub_command" }
func (*skipCheckContextKeyType) String() string  { return "skip_check" }

var (
	startTimeContextKey  = startTimeContextKeyType(struct{}{})
	commandContextKey    = commandContextKeyType(struct{}{})
	subCommandContextKey = subCommandContextKeyType(struct{}{})
	skipCheckContextKey  = skipCheckContextKeyType(struct{}{})
)

// WithSkipCheck 是否跳过检测
// 在 Development 的情况下，会跳过 黑名单检验、版本检验、槽位检测以及警告输出
func WithSkipCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipCheckContextKey, true)
}

// WithSubCommandName 指定 sub command
func WithSubCommandName(ctx context.Context, s string) context.Context {
	if len(s) == 0 {
		return ctx
	}
	return context.WithValue(ctx, subCommandContextKey, s)
}

func (r *baseHandler) isCluster() bool              { return r.cluster.Load() }
func (r *baseHandler) setIsCluster(b bool)          { r.cluster.Store(b) }
func (r *baseHandler) setVersion(v *semver.Version) { r.version.Store(v) }
func (r *baseHandler) getVersion() *semver.Version  { return r.version.Load() }
func (r *baseHandler) setSilentErrCallback(b isSilentError) {
	r.silentErrCallback.Store(silentErrCallbackBox{f: b})
}
func (r *baseHandler) getSilentErrCallback() isSilentError {
	if v := r.silentErrCallback.Load(); v != nil {
		if box, ok := v.(silentErrCallbackBox); ok {
			return box.f
		}
	}
	return nil
}
func (r *baseHandler) setRegisterCollector(rc RegisterCollectorFunc) { registerMetric(rc) }
func (r *baseHandler) before(ctx context.Context, command Command) context.Context {
	return r.beforeWithKeys(ctx, command, nil)
}
func (r *baseHandler) beforeWithKeys(ctx context.Context, command Command, getKeys func() []string) context.Context {
	if r.v.GetDevelopment() {
		if skipCheck := ctx.Value(skipCheckContextKey); skipCheck == nil {
			// 需要检验命令是否在黑名单
			if command.Forbid() {
				e(fmt.Sprintf("[%s]: redis command are not allowed", command.String()))
			}
			// 需要检验版本是否支持该命令
			ver := r.getVersion()
			if ver != nil && ver.LessThan(mustNewSemVersion(command.RequireVersion())) {
				e(fmt.Sprintf("[%s]: redis command are not supported in version %q, available since %s", command, ver, command.RequireVersion()))
			}
			if r.cluster.Load() {
				// 需要检验所有的key是否均在同一槽位
				if err := checkMultipleKeySlots(command, getKeys); err != nil {
					e(err.Error())
				}
			}
			// 该命令是否有警告日志输出
			if ver != nil {
				if wv := command.WarnVersion(); len(wv) > 0 && mustNewSemVersion(wv).LessThan(*ver) {
					needWarning := false
					if command.WarningOnce() {
						cs := command.String()
						r.mx.Lock()
						if _, ok := r.warningOnceMapping[cs]; !ok {
							needWarning = true
							r.warningOnceMapping[cs] = struct{}{}
						}
						r.mx.Unlock()
					} else {
						needWarning = true
					}
					if needWarning {
						instead := command.Instead()
						etc := command.ETC()
						if len(instead) > 0 && len(etc) > 0 {
							warning(fmt.Sprintf("[%s]: %s \n\t\t use '%s' instead. \n\t\t %s, etc.", command.String(), command.Warning(), instead, etc))
						} else if len(instead) > 0 {
							warning(fmt.Sprintf("[%s]: %s \n\t\t use '%s' instead.", command.String(), command.Warning(), instead))
						} else if len(etc) > 0 {
							warning(fmt.Sprintf("[%s]: %s \n\t\t %s, etc.", command.String(), command.Warning(), etc))
						} else {
							warning(fmt.Sprintf("[%s]: %s", command.String(), command.Warning()))
						}
					}
				}
			}
		}
	}
	if r.v.GetEnableMonitor() {
		ctx = context.WithValue(ctx, startTimeContextKey, nowFunc())
		ctx = context.WithValue(ctx, commandContextKey, command.Class())
		ctx = context.WithValue(ctx, subCommandContextKey, command.String())
	}
	return ctx
}
func (r *baseHandler) isImplicitError(err error) bool {
	cb := r.getSilentErrCallback()
	if cb == nil {
		return false
	}
	return cb(err)
}
func (r *baseHandler) after(ctx context.Context, err error) {
	if !r.v.GetEnableMonitor() || r.metrics == nil {
		return
	}
	cmd := ctx.Value(commandContextKey).(string)
	subCmd := ctx.Value(subCommandContextKey).(string)
	if err != nil && !r.isImplicitError(err) {
		r.metrics.err.WithLabelValues(cmd, subCmd).Inc()
	} else {
		r.metrics.timing.WithLabelValues(cmd, subCmd).Observe(sinceFunc(ctx.Value(startTimeContextKey).(time.Time)).Seconds())
	}
}
func (r *baseHandler) cache(ctx context.Context, hit bool) {
	if !r.v.GetEnableMonitor() || r.metrics == nil {
		return
	}
	cmd := ctx.Value(commandContextKey).(string)
	subCmd := ctx.Value(subCommandContextKey).(string)
	if hit {
		r.metrics.hits.WithLabelValues(cmd, subCmd).Inc()
	} else {
		r.metrics.miss.WithLabelValues(cmd, subCmd).Inc()
	}
}
func (r *baseHandler) delayPollError(name string) {
	if r.v.GetEnableMonitor() && r.metrics != nil {
		r.metrics.delayPollError.WithLabelValues(name).Inc()
	}
}
func (r *baseHandler) delayReclaimError(name string) {
	if r.v.GetEnableMonitor() && r.metrics != nil {
		r.metrics.delayReclaimError.WithLabelValues(name).Inc()
	}
}
func (r *baseHandler) delayReclaim(name string, count int) {
	if r.v.GetEnableMonitor() && r.metrics != nil {
		r.metrics.delayReclaimCount.WithLabelValues(name).Add(float64(count))
	}
}
