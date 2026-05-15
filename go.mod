module github.com/sandwich-go/redisson

go 1.25.0

// 当用户机器只有 1.25.x 而 patch 低于 1.25.0 时，Go 工具链会自动下载满足要求的版本。
// 也兼容 Go 1.24.x 的用户：toolchain 指令会让 go 命令自动拉取 go1.25.10 来满足 go 1.25.0 要求。
toolchain go1.25.10

require (
	github.com/alicebob/miniredis/v2 v2.38.0
	github.com/coreos/go-semver v0.3.1
	github.com/modern-go/reflect2 v1.0.2
	github.com/prometheus/client_golang v1.23.2
	github.com/redis/rueidis v1.0.75
	github.com/redis/rueidis/rueidiscompat v1.0.75
	github.com/redis/rueidis/rueidislimiter v1.0.75
	github.com/redis/rueidis/rueidisprob v1.0.75
	github.com/smartystreets/goconvey v1.8.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/smarty/assertions v1.15.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.43.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)
