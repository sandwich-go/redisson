package redisson

import (
	"time"
)

type Tester interface {
	Fatalf(string, ...interface{})
	Cleanup(func())
}

var defaultWriteTimeout = 10 * time.Second

//go:generate optiongen --new_func=NewConf --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func ConfOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		"Net":               "tcp",                              // @MethodComment(网络连接类型，tcp/unix)
		"EnableInit":        true,                               // @MethodComment(是否需要进行初始化)
		"Resp":              RESP(RESP3),                        // @MethodComment(RESP版本)
		"AlwaysRESP2":       bool(false),                        // @MethodComment(always uses RESP2, otherwise it will try using RESP3 first)
		"Name":              "",                                 // @MethodComment(Redis客户端名字)
		"MasterName":        "",                                 // @MethodComment(Redis Sentinel模式下，master名字)
		"EnableMonitor":     true,                               // @MethodComment(是否开启监控)
		"Addrs":             []string{"127.0.0.1:6379"},         // @MethodComment(Redis地址列表)
		"DB":                0,                                  // @MethodComment(Redis实例数据库编号，集群下只能用0)
		"Username":          "",                                 // @MethodComment(Redis用户名)
		"Password":          "",                                 // @MethodComment(Redis用户密码)
		"ReadTimeout":       time.Duration(defaultWriteTimeout), // @MethodComment(Redis连接读取的超时时长)
		"WriteTimeout":      time.Duration(defaultWriteTimeout), // @MethodComment(Redis连接写入的超时时长)
		"ConnPoolSize":      0,                                  // @MethodComment(Redis连接池大小，默认0，RESP2时，即非集群模式下为10*runtime.GOMAXPROCS，集群模式下为5*runtime.GOMAXPROCS。RESP3时，为Block连接池，默认1000)
		"MinIdleConns":      0,                                  // @MethodComment(Redis连接池最小空闲连接数量，RESP2时有效)
		"ConnMaxAge":        time.Duration(4 * time.Hour),       // @MethodComment(Redis连接生命周期，RESP2时有效)
		"ConnPoolTimeout":   time.Duration(0),                   // @MethodComment(Redis获取连接超时时间，默认0s，表示socket_read_timeout+1s，RESP2时有效)
		"IdleConnTimeout":   time.Duration(-1 * time.Second),    // @MethodComment(Redis连接空闲超时时间，默认-1s，表示空闲连接不会被回收，RESP2时有效)
		"EnableCache":       true,                               // @MethodComment(是否开启客户端缓存，RESP3时有效)
		"CacheSizeEachConn": 0,                                  // @MethodComment(开启客户端缓存时，单个连接缓存大小，默认128 MiB，RESP3时有效)
		"RingScaleEachConn": 0,                                  // @MethodComment(单个连接ring buffer大小，默认2 ^ RingScaleEachConn, RingScaleEachConn默认情况下为10，RESP3时有效)
		"Cluster":           false,                              // @MethodComment(是否为Redis集群，默认为false，集群需要设置为true)
		"Development":       true,                               // @MethodComment(是否为开发模式，开发模式下，使用部分接口会有警告日志输出，会校验多key是否为同一hash槽，会校验部分接口是否满足版本要求)
		"T":                 (Tester)(nil),                      // @MethodComment(如果设置该值，则启动mock)
		"ForceSingleClient": false,                              // @MethodComment(ForceSingleClient force the usage of a single client connection, without letting the lib guessing)
	}
}

func revise(v ConfInterface) {
	if v.GetWriteTimeout() == 0 {
		v.ApplyOption(WithWriteTimeout(defaultWriteTimeout))
	}
	if v.GetReadTimeout() == 0 {
		v.ApplyOption(WithReadTimeout(defaultWriteTimeout))
	}
}
