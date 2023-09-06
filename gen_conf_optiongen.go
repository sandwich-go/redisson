// Code generated by optiongen. DO NOT EDIT.
// optiongen: github.com/timestee/optiongen

package redisson

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// Conf should use NewConf to initialize it
type Conf struct {
	Resp              RESP          `xconf:"resp" usage:"RESP版本"`
	Name              string        `xconf:"name" usage:"Redis客户端名字"`
	MasterName        string        `xconf:"master_name" usage:"Redis Sentinel模式下，master名字"`
	EnableMonitor     bool          `xconf:"enable_monitor" usage:"是否开启监控"`
	Addrs             []string      `xconf:"addrs" usage:"Redis地址列表"`
	DB                int           `xconf:"db" usage:"Redis实例数据库编号，集群下只能用0"`
	Username          string        `xconf:"username" usage:"Redis用户名"`
	Password          string        `xconf:"password" usage:"Redis用户密码"`
	ReadTimeout       time.Duration `xconf:"read_timeout" usage:"Redis连接读取的超时时长"`
	WriteTimeout      time.Duration `xconf:"write_timeout" usage:"Redis连接写入的超时时长"`
	ConnPoolSize      int           `xconf:"conn_pool_size" usage:"Redis连接池大小，默认0，RESP2时，即非集群模式下为10*runtime.GOMAXPROCS，集群模式下为5*runtime.GOMAXPROCS。RESP3时，为Block连接池，默认1000"`
	MinIdleConns      int           `xconf:"min_idle_conns" usage:"Redis连接池最小空闲连接数量，RESP2时有效"`
	ConnMaxAge        time.Duration `xconf:"conn_max_age" usage:"Redis连接生命周期，RESP2时有效"`
	ConnPoolTimeout   time.Duration `xconf:"conn_pool_timeout" usage:"Redis获取连接超时时间，默认0s，表示socket_read_timeout+1s，RESP2时有效"`
	IdleConnTimeout   time.Duration `xconf:"idle_conn_timeout" usage:"Redis连接空闲超时时间，默认-1s，表示空闲连接不会被回收，RESP2时有效"`
	EnableCache       bool          `xconf:"enable_cache" usage:"是否开启客户端缓存，RESP3时有效"`
	CacheSizeEachConn int           `xconf:"cache_size_each_conn" usage:"开启客户端缓存时，单个连接缓存大小，默认128 MiB，RESP3时有效"`
	RingScaleEachConn int           `xconf:"ring_scale_each_conn" usage:"单个连接ring buffer大小，默认2 ^ RingScaleEachConn, RingScaleEachConn默认情况下为10，RESP3时有效"`
	Cluster           bool          `xconf:"cluster" usage:"是否为Redis集群，默认为false，集群需要设置为true"`
	Development       bool          `xconf:"development" usage:"是否为开发模式，开发模式下，使用部分接口会有警告日志输出，会校验多key是否为同一hash槽，会校验部分接口是否满足版本要求"`
	T                 Tester        `xconf:"t" usage:"如果设置该值，则启动mock"`
}

// NewConf new Conf
func NewConf(opts ...ConfOption) *Conf {
	cc := newDefaultConf()
	for _, opt := range opts {
		opt(cc)
	}
	if watchDogConf != nil {
		watchDogConf(cc)
	}
	return cc
}

// ApplyOption apply multiple new option and return the old ones
// sample:
// old := cc.ApplyOption(WithTimeout(time.Second))
// defer cc.ApplyOption(old...)
func (cc *Conf) ApplyOption(opts ...ConfOption) []ConfOption {
	var previous []ConfOption
	for _, opt := range opts {
		previous = append(previous, opt(cc))
	}
	return previous
}

// ConfOption option func
type ConfOption func(cc *Conf) ConfOption

// WithResp RESP版本
func WithResp(v RESP) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.Resp
		cc.Resp = v
		return WithResp(previous)
	}
}

// WithName Redis客户端名字
func WithName(v string) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.Name
		cc.Name = v
		return WithName(previous)
	}
}

// WithMasterName Redis Sentinel模式下，master名字
func WithMasterName(v string) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.MasterName
		cc.MasterName = v
		return WithMasterName(previous)
	}
}

// WithEnableMonitor 是否开启监控
func WithEnableMonitor(v bool) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.EnableMonitor
		cc.EnableMonitor = v
		return WithEnableMonitor(previous)
	}
}

// WithAddrs Redis地址列表
func WithAddrs(v ...string) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.Addrs
		cc.Addrs = v
		return WithAddrs(previous...)
	}
}

// WithDB Redis实例数据库编号，集群下只能用0
func WithDB(v int) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.DB
		cc.DB = v
		return WithDB(previous)
	}
}

// WithUsername Redis用户名
func WithUsername(v string) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.Username
		cc.Username = v
		return WithUsername(previous)
	}
}

// WithPassword Redis用户密码
func WithPassword(v string) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.Password
		cc.Password = v
		return WithPassword(previous)
	}
}

// WithReadTimeout Redis连接读取的超时时长
func WithReadTimeout(v time.Duration) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.ReadTimeout
		cc.ReadTimeout = v
		return WithReadTimeout(previous)
	}
}

// WithWriteTimeout Redis连接写入的超时时长
func WithWriteTimeout(v time.Duration) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.WriteTimeout
		cc.WriteTimeout = v
		return WithWriteTimeout(previous)
	}
}

// WithConnPoolSize Redis连接池大小，默认0，RESP2时，即非集群模式下为10*runtime.GOMAXPROCS，集群模式下为5*runtime.GOMAXPROCS。RESP3时，为Block连接池，默认1000
func WithConnPoolSize(v int) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.ConnPoolSize
		cc.ConnPoolSize = v
		return WithConnPoolSize(previous)
	}
}

// WithMinIdleConns Redis连接池最小空闲连接数量，RESP2时有效
func WithMinIdleConns(v int) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.MinIdleConns
		cc.MinIdleConns = v
		return WithMinIdleConns(previous)
	}
}

// WithConnMaxAge Redis连接生命周期，RESP2时有效
func WithConnMaxAge(v time.Duration) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.ConnMaxAge
		cc.ConnMaxAge = v
		return WithConnMaxAge(previous)
	}
}

// WithConnPoolTimeout Redis获取连接超时时间，默认0s，表示socket_read_timeout+1s，RESP2时有效
func WithConnPoolTimeout(v time.Duration) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.ConnPoolTimeout
		cc.ConnPoolTimeout = v
		return WithConnPoolTimeout(previous)
	}
}

// WithIdleConnTimeout Redis连接空闲超时时间，默认-1s，表示空闲连接不会被回收，RESP2时有效
func WithIdleConnTimeout(v time.Duration) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.IdleConnTimeout
		cc.IdleConnTimeout = v
		return WithIdleConnTimeout(previous)
	}
}

// WithEnableCache 是否开启客户端缓存，RESP3时有效
func WithEnableCache(v bool) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.EnableCache
		cc.EnableCache = v
		return WithEnableCache(previous)
	}
}

// WithCacheSizeEachConn 开启客户端缓存时，单个连接缓存大小，默认128 MiB，RESP3时有效
func WithCacheSizeEachConn(v int) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.CacheSizeEachConn
		cc.CacheSizeEachConn = v
		return WithCacheSizeEachConn(previous)
	}
}

// WithRingScaleEachConn 单个连接ring buffer大小，默认2 ^ RingScaleEachConn, RingScaleEachConn默认情况下为10，RESP3时有效
func WithRingScaleEachConn(v int) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.RingScaleEachConn
		cc.RingScaleEachConn = v
		return WithRingScaleEachConn(previous)
	}
}

// WithCluster 是否为Redis集群，默认为false，集群需要设置为true
func WithCluster(v bool) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.Cluster
		cc.Cluster = v
		return WithCluster(previous)
	}
}

// WithDevelopment 是否为开发模式，开发模式下，使用部分接口会有警告日志输出，会校验多key是否为同一hash槽，会校验部分接口是否满足版本要求
func WithDevelopment(v bool) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.Development
		cc.Development = v
		return WithDevelopment(previous)
	}
}

// WithT 如果设置该值，则启动mock
func WithT(v Tester) ConfOption {
	return func(cc *Conf) ConfOption {
		previous := cc.T
		cc.T = v
		return WithT(previous)
	}
}

// InstallConfWatchDog the installed func will called when NewConf  called
func InstallConfWatchDog(dog func(cc *Conf)) { watchDogConf = dog }

// watchDogConf global watch dog
var watchDogConf func(cc *Conf)

// newDefaultConf new default Conf
func newDefaultConf() *Conf {
	cc := &Conf{}

	for _, opt := range [...]ConfOption{
		WithResp(RESP3),
		WithName(""),
		WithMasterName(""),
		WithEnableMonitor(true),
		WithAddrs([]string{"127.0.0.1:6379"}...),
		WithDB(0),
		WithUsername(""),
		WithPassword(""),
		WithReadTimeout(10 * time.Second),
		WithWriteTimeout(10 * time.Second),
		WithConnPoolSize(0),
		WithMinIdleConns(0),
		WithConnMaxAge(4 * time.Hour),
		WithConnPoolTimeout(0),
		WithIdleConnTimeout(-1 * time.Second),
		WithEnableCache(true),
		WithCacheSizeEachConn(0),
		WithRingScaleEachConn(0),
		WithCluster(false),
		WithDevelopment(true),
		WithT(nil),
	} {
		opt(cc)
	}

	return cc
}

// AtomicSetFunc used for XConf
func (cc *Conf) AtomicSetFunc() func(interface{}) { return AtomicConfSet }

// atomicConf global *Conf holder
var atomicConf unsafe.Pointer

// onAtomicConfSet global call back when  AtomicConfSet called by XConf.
// use ConfInterface.ApplyOption to modify the updated cc
// if passed in cc not valid, then return false, cc will not set to atomicConf
var onAtomicConfSet func(cc ConfInterface) bool

// InstallCallbackOnAtomicConfSet install callback
func InstallCallbackOnAtomicConfSet(callback func(cc ConfInterface) bool) { onAtomicConfSet = callback }

// AtomicConfSet atomic setter for *Conf
func AtomicConfSet(update interface{}) {
	cc := update.(*Conf)
	if onAtomicConfSet != nil && !onAtomicConfSet(cc) {
		return
	}
	atomic.StorePointer(&atomicConf, (unsafe.Pointer)(cc))
}

// AtomicConf return atomic *ConfVisitor
func AtomicConf() ConfVisitor {
	current := (*Conf)(atomic.LoadPointer(&atomicConf))
	if current == nil {
		defaultOne := newDefaultConf()
		if watchDogConf != nil {
			watchDogConf(defaultOne)
		}
		atomic.CompareAndSwapPointer(&atomicConf, nil, (unsafe.Pointer)(defaultOne))
		return (*Conf)(atomic.LoadPointer(&atomicConf))
	}
	return current
}

// all getter func
func (cc *Conf) GetResp() RESP                     { return cc.Resp }
func (cc *Conf) GetName() string                   { return cc.Name }
func (cc *Conf) GetMasterName() string             { return cc.MasterName }
func (cc *Conf) GetEnableMonitor() bool            { return cc.EnableMonitor }
func (cc *Conf) GetAddrs() []string                { return cc.Addrs }
func (cc *Conf) GetDB() int                        { return cc.DB }
func (cc *Conf) GetUsername() string               { return cc.Username }
func (cc *Conf) GetPassword() string               { return cc.Password }
func (cc *Conf) GetReadTimeout() time.Duration     { return cc.ReadTimeout }
func (cc *Conf) GetWriteTimeout() time.Duration    { return cc.WriteTimeout }
func (cc *Conf) GetConnPoolSize() int              { return cc.ConnPoolSize }
func (cc *Conf) GetMinIdleConns() int              { return cc.MinIdleConns }
func (cc *Conf) GetConnMaxAge() time.Duration      { return cc.ConnMaxAge }
func (cc *Conf) GetConnPoolTimeout() time.Duration { return cc.ConnPoolTimeout }
func (cc *Conf) GetIdleConnTimeout() time.Duration { return cc.IdleConnTimeout }
func (cc *Conf) GetEnableCache() bool              { return cc.EnableCache }
func (cc *Conf) GetCacheSizeEachConn() int         { return cc.CacheSizeEachConn }
func (cc *Conf) GetRingScaleEachConn() int         { return cc.RingScaleEachConn }
func (cc *Conf) GetCluster() bool                  { return cc.Cluster }
func (cc *Conf) GetDevelopment() bool              { return cc.Development }
func (cc *Conf) GetT() Tester                      { return cc.T }

// ConfVisitor visitor interface for Conf
type ConfVisitor interface {
	GetResp() RESP
	GetName() string
	GetMasterName() string
	GetEnableMonitor() bool
	GetAddrs() []string
	GetDB() int
	GetUsername() string
	GetPassword() string
	GetReadTimeout() time.Duration
	GetWriteTimeout() time.Duration
	GetConnPoolSize() int
	GetMinIdleConns() int
	GetConnMaxAge() time.Duration
	GetConnPoolTimeout() time.Duration
	GetIdleConnTimeout() time.Duration
	GetEnableCache() bool
	GetCacheSizeEachConn() int
	GetRingScaleEachConn() int
	GetCluster() bool
	GetDevelopment() bool
	GetT() Tester
}

// ConfInterface visitor + ApplyOption interface for Conf
type ConfInterface interface {
	ConfVisitor
	ApplyOption(...ConfOption) []ConfOption
}
