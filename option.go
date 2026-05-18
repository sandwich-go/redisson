package redisson

import (
	"time"
)

type Tester interface {
	Fatalf(string, ...any)
	Cleanup(func())
	Logf(format string, args ...interface{})
}

var (
	defaultWriteTimeout     = 10 * time.Second
	defaultPubSubChanSize   = 1024
	defaultBootstrapTimeout = 30 * time.Second
)

//go:generate optiongen --new_func=NewConf --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func ConfOptionDeclareWithDefault() any {
	return map[string]any{
		"Net":               "tcp",                      // @MethodComment(网络类型，tcp/unix)
		"AlwaysRESP2":       bool(false),                // @MethodComment(always uses RESP2, otherwise it will try using RESP3 first)
		"Name":              "",                         // @MethodComment(Redis客户端名字)
		"MasterName":        "",                         // @MethodComment(Redis Sentinel模式下，master名字)
		"EnableMonitor":     true,                       // @MethodComment(是否开启监控)
		"Addrs":             []string{"127.0.0.1:6379"}, // @MethodComment(Redis地址列表)
		"DB":                0,                          // @MethodComment(Redis实例数据库编号，集群下只能用0)
		"Username":          "",                         // @MethodComment(Redis用户名)
		"Password":          "",                         // @MethodComment(Redis用户密码)
		"WriteTimeout":      defaultWriteTimeout,        // @MethodComment(Redis连接写入的超时时长)
		"BootstrapTimeout":  defaultBootstrapTimeout,    // @MethodComment(连接建立时执行 INFO 等启动期探测的总超时；集群+多副本下默认 30s)
		"ConnPoolSize":      0,                          // @MethodComment(Redis 阻塞命令(BLPOP 等)使用的连接池大小；0 表示使用底层 rueidis 默认值（约 1000）)
		"EnableCache":       true,                       // @MethodComment(是否开启客户端缓存)
		"CacheSizeEachConn": 0,                          // @MethodComment(开启客户端缓存时，单个连接缓存大小（字节）；0 表示使用底层 rueidis 默认值（约 128 MiB）)
		"RingScaleEachConn": 0,                          // @MethodComment(单个连接 ring buffer 容量为 2^RingScaleEachConn；0 表示使用底层 rueidis 默认值（10，即 1024）)
		"Development":       false,                      // @MethodComment(是否为开发模式，开发模式下，使用部分接口会有警告日志输出，会校验多key是否为同一hash槽，会校验部分接口是否满足版本要求；生产环境请保持 false)
		"T":                 (Tester)(nil),              // @MethodComment(如果设置该值，则启动mock)
		"ForceSingleClient": false,                      // @MethodComment(ForceSingleClient force the usage of a single client connection, without letting the lib guessing)
		"PubSubChanSize":    defaultPubSubChanSize,      // @MethodComment(pubsub chan 大小)
	}
}

func revise(v ConfInterface) {
	if v.GetWriteTimeout() == 0 {
		v.ApplyOption(WithWriteTimeout(defaultWriteTimeout))
	}
	if v.GetPubSubChanSize() == 0 {
		v.ApplyOption(WithPubSubChanSize(defaultPubSubChanSize))
	}
	if v.GetBootstrapTimeout() == 0 {
		v.ApplyOption(WithBootstrapTimeout(defaultBootstrapTimeout))
	}
}
