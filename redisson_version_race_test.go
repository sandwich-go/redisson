package redisson

import (
	"sync"
	"testing"

	"github.com/coreos/go-semver/semver"
)

// TestClientVersion_ConcurrentReadWrite 回归 client.version 改 atomic.Pointer 后
// 在并发读写下不应触发 -race 告警。
//
// 旧实现 c.version 是 plain semver.Version 字段，reconnect 路径写入与
// 命令读形成 data race。改 atomic.Pointer[semver.Version] 后消除。
func TestClientVersion_ConcurrentReadWrite(t *testing.T) {
	c := &client{}
	v1 := semver.New("7.0.0")
	v2 := semver.New("7.2.5")
	c.version.Store(v1)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if i%2 == 0 {
				c.version.Store(v1)
			} else {
				c.version.Store(v2)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = c.Version()
		}
	}()
	wg.Wait()
}

// TestClientIsCluster_ConcurrentReadWrite 同上：isCluster 改 atomic.Bool。
func TestClientIsCluster_ConcurrentReadWrite(t *testing.T) {
	c := &client{}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			c.isCluster.Store(i%2 == 0)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = c.IsCluster()
		}
	}()
	wg.Wait()
}

// TestCloneClientFrom 验证派生 client 时 atomic 字段被正确复制。
func TestCloneClientFrom(t *testing.T) {
	src := &client{v: NewConf()}
	v := semver.New("7.2.5")
	src.version.Store(v)
	src.isCluster.Store(true)

	dst := cloneClientFrom(src)
	if got := dst.Version(); got == nil || got.String() != "7.2.5" {
		t.Fatalf("clone version: got %v, want 7.2.5", got)
	}
	if !dst.IsCluster() {
		t.Fatalf("clone isCluster: got false, want true")
	}

	// 修改 src 不影响 dst（atomic.Pointer 存的是指针，但 cloneClientFrom 调 Store
	// 把 dst.version 指向同一个 *Version；这里只验证 Load 出来的值仍正确）。
	src.isCluster.Store(false)
	if !dst.IsCluster() {
		t.Fatalf("dst.isCluster should remain true after src changed")
	}
}
