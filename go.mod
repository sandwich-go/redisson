module github.com/sandwich-go/redisson

go 1.22

toolchain go1.23.3

require (
	github.com/alicebob/miniredis/v2 v2.30.5
	github.com/coreos/go-semver v0.3.1
	github.com/modern-go/reflect2 v1.0.2
	github.com/prometheus/client_golang v1.14.0
	github.com/redis/rueidis v1.0.50-alpha.3
	github.com/redis/rueidis/rueidiscompat v1.0.49
	github.com/redis/rueidis/rueidisprob v1.0.49
	github.com/sandwich-go/funnel v0.0.1
	github.com/smartystreets/goconvey v1.7.2
)

replace (
	github.com/redis/rueidis => github.com/sandwich-go/rueidis v1.0.50-0.20241203070424-17992444f236
	github.com/redis/rueidis/rueidiscompat => github.com/sandwich-go/rueidis/rueidiscompat v1.0.50-0.20241203070424-17992444f236
)

//replace (
//	github.com/redis/rueidis => ../rueidis
//	github.com/redis/rueidis/rueidiscompat => ../rueidis/rueidiscompat
//)

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
