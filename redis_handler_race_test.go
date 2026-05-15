package redisson

import (
	"errors"
	"testing"

	"github.com/coreos/go-semver/semver"
)

// TestBaseHandlerConcurrentVersionReadWrite 回归 Bug #9：
// baseHandler.version / cluster / silentErrCallback 改 atomic 后
// 在并发读写下不应触发 -race 告警（go test -race 即可验证）。
func TestBaseHandlerConcurrentVersionReadWrite(t *testing.T) {
	h := newBaseHandler(NewConf()).(*baseHandler)
	v := semver.New("7.2.5")

	done := make(chan struct{})
	// writer
	go func() {
		for i := 0; i < 1000; i++ {
			h.setVersion(v)
		}
		done <- struct{}{}
	}()
	// reader
	go func() {
		for i := 0; i < 1000; i++ {
			_ = h.getVersion()
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}

func TestBaseHandlerConcurrentClusterReadWrite(t *testing.T) {
	h := newBaseHandler(NewConf()).(*baseHandler)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			h.setIsCluster(i%2 == 0)
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			_ = h.isCluster()
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}

func TestBaseHandlerConcurrentSilentErrCallback(t *testing.T) {
	h := newBaseHandler(NewConf()).(*baseHandler)
	cb1 := func(err error) bool { return err == nil }
	cb2 := func(err error) bool { return err != nil }

	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			h.setSilentErrCallback(cb1)
			h.setSilentErrCallback(cb2)
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			_ = h.isImplicitError(errors.New("x"))
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}
