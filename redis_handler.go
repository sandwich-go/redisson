package redisson

import (
	"context"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	timingMetric = "redis_exec_timing"
	errorMetric  = "redis_exec_error"
	hitsMetric   = "redis_cache_hits"
	missMetric   = "redis_cache_miss"
)

type isSilentError func(error) bool

type handler interface {
	setVersion(*semver.Version)
	setSilentErrCallback(isSilentError)
	setRegisterCollector(RegisterCollectorFunc)

	before(ctx context.Context, command Command) context.Context
	beforeWithKeys(ctx context.Context, command Command, getKeys func() []string) context.Context
	after(ctx context.Context, err error)
	cache(ctx context.Context, hit bool)
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

var labelKeys = []string{"command", "s_command"}

type baseHandler struct {
	metric                            *prometheus.SummaryVec
	errMetric, hitsMetric, missMetric *prometheus.CounterVec
	silentErrCallback                 isSilentError
	v                                 ConfVisitor
	version                           *semver.Version
}

func newBaseHandler(v ConfVisitor) handler {
	h := &baseHandler{v: v}
	h.errMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: errorMetric,
	}, labelKeys)
	h.hitsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: hitsMetric,
	}, labelKeys)
	h.missMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: missMetric,
	}, labelKeys)
	h.metric = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       timingMetric,
		Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.02, 0.99: 0.001, 1: 0},
		MaxAge:     time.Minute,
	}, labelKeys)
	return h
}

type (
	startTimeContextKeyType  struct{}
	commandContextKeyType    struct{}
	subCommandContextKeyType struct{}
)

func (*startTimeContextKeyType) String() string  { return "start_time" }
func (*commandContextKeyType) String() string    { return "command" }
func (*subCommandContextKeyType) String() string { return "sub_command" }

var (
	startTimeContextKey  = startTimeContextKeyType(struct{}{})
	commandContextKey    = commandContextKeyType(struct{}{})
	subCommandContextKey = subCommandContextKeyType(struct{}{})
)

func (r *baseHandler) setVersion(v *semver.Version)         { r.version = v }
func (r *baseHandler) setSilentErrCallback(b isSilentError) { r.silentErrCallback = b }
func (r *baseHandler) setRegisterCollector(rc RegisterCollectorFunc) {
	rc(r.errMetric)
	rc(r.hitsMetric)
	rc(r.missMetric)
	rc(r.metric)
}
func (r *baseHandler) before(ctx context.Context, command Command) context.Context {
	return r.beforeWithKeys(ctx, command, nil)
}
func (r *baseHandler) beforeWithKeys(ctx context.Context, command Command, getKeys func() []string) context.Context {
	if r.v.GetDevelopment() {
		// ????????????????????????????????????
		if command.Forbid() {
			panic(fmt.Errorf("[%s]: redis command are not allowed", command.String()))
		}
		// ???????????????????????????????????????
		if r.version.LessThan(mustNewSemVersion(command.RequireVersion())) {
			panic(fmt.Errorf("[%s]: redis command are not supported in version %q, available since %s", command, r.version, command.RequireVersion()))
		}
		// ?????????????????????key????????????????????????
		panicIfUseMultipleKeySlots(command, getKeys)
		// ????????????????????????????????????
		if len(command.WarnVersion()) > 0 && mustNewSemVersion(command.WarnVersion()).LessThan(*r.version) {
			warning(fmt.Sprintf("[%s]: %s", command.String(), command.Warning()))
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
	if r.silentErrCallback == nil {
		return false
	}
	return r.silentErrCallback(err)
}
func (r *baseHandler) after(ctx context.Context, err error) {
	if r.v.GetEnableMonitor() {
		if err != nil && !r.isImplicitError(err) {
			r.errMetric.WithLabelValues(ctx.Value(commandContextKey).(string), ctx.Value(subCommandContextKey).(string)).Inc()
		} else {
			r.metric.WithLabelValues(ctx.Value(commandContextKey).(string), ctx.Value(subCommandContextKey).(string)).
				Observe(sinceFunc(ctx.Value(startTimeContextKey).(time.Time)).Seconds())
		}
	}
}
func (r *baseHandler) cache(ctx context.Context, hit bool) {
	if r.v.GetEnableMonitor() {
		if hit {
			r.hitsMetric.WithLabelValues(ctx.Value(commandContextKey).(string), ctx.Value(subCommandContextKey).(string)).Inc()
		} else {
			r.missMetric.WithLabelValues(ctx.Value(commandContextKey).(string), ctx.Value(subCommandContextKey).(string)).Inc()
		}
	}
}
