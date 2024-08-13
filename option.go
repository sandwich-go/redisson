package redisson

import (
	"time"
)

type Tester interface {
	Fatalf(string, ...any)
	Cleanup(func())
	Logf(format string, args ...interface{})
}

//go:generate optiongen --new_func=NewConf --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func ConfOptionDeclareWithDefault() any {
	return map[string]any{
		"AlwaysRESP2":       bool(false),                     // @MethodComment(always uses RESP2, otherwise it will try using RESP3 first)
		"Name":              "",                              // @MethodComment(Redis客户端名字)
		"MasterName":        "",                              // @MethodComment(Redis Sentinel模式下，master名字)
		"EnableMonitor":     true,                            // @MethodComment(是否开启监控)
		"Addrs":             []string{"127.0.0.1:6379"},      // @MethodComment(Redis地址列表)
		"DB":                0,                               // @MethodComment(Redis实例数据库编号，集群下只能用0)
		"Username":          "",                              // @MethodComment(Redis用户名)
		"Password":          "",                              // @MethodComment(Redis用户密码)
		"WriteTimeout":      time.Duration(10 * time.Second), // @MethodComment(Redis连接写入的超时时长)
		"ConnPoolSize":      0,                               // @MethodComment(RedisBlock连接池，默认1000)
		"EnableCache":       true,                            // @MethodComment(是否开启客户端缓存)
		"CacheSizeEachConn": 0,                               // @MethodComment(开启客户端缓存时，单个连接缓存大小，默认128 MiB)
		"RingScaleEachConn": 0,                               // @MethodComment(单个连接ring buffer大小，默认2 ^ RingScaleEachConn, RingScaleEachConn默认情况下为10)
		"Development":       true,                            // @MethodComment(是否为开发模式，开发模式下，使用部分接口会有警告日志输出，会校验多key是否为同一hash槽，会校验部分接口是否满足版本要求)
		"T":                 (Tester)(nil),                   // @MethodComment(如果设置该值，则启动mock)
		"ForceSingleClient": false,                           // @MethodComment(ForceSingleClient force the usage of a single client connection, without letting the lib guessing)
	}
}
