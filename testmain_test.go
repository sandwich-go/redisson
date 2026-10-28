package redisson

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"
)

// TestMain 为会清库的集成测试启动独立 Redis，禁止默认连接开发者已有的实例。
func TestMain(m *testing.M) {
	defaults := NewConf()
	defaultAddresses := slices.Clone(defaults.Addrs)
	defaultNetwork := defaults.Net
	address, stop, err := startIsolatedTestRedis()
	if err != nil {
		fmt.Fprintln(os.Stderr, "start isolated Redis:", err)
		os.Exit(1)
	}
	previous := watchDogConf
	InstallConfWatchDog(func(conf *Conf) {
		if previous != nil {
			previous(conf)
		}
		if conf.T != nil || conf.Net != defaultNetwork || !slices.Equal(conf.Addrs, defaultAddresses) {
			return
		}
		conf.ApplyOption(WithNet("unix"), WithAddrs(address))
	})
	code := m.Run()
	stop()
	os.Exit(code)
}

func startIsolatedTestRedis() (string, func(), error) {
	binary, err := exec.LookPath("redis-server")
	if err != nil {
		return "", nil, fmt.Errorf("redis-server is required for isolated integration tests: %w", err)
	}
	directory, err := os.MkdirTemp("", "redisson-test-")
	if err != nil {
		return "", nil, err
	}
	socket := filepath.Join(directory, "r.sock")
	log, err := os.Create(filepath.Join(directory, "redis.log"))
	if err != nil {
		_ = os.RemoveAll(directory)
		return "", nil, err
	}
	cmd := exec.Command(binary, "--port", "0", "--unixsocket", socket, "--unixsocketperm", "700", "--save", "", "--appendonly", "no", "--dir", directory)
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Start(); err != nil {
		_ = log.Close()
		_ = os.RemoveAll(directory)
		return "", nil, err
	}
	done := make(chan struct{})
	var waitError error
	go func() { waitError = cmd.Wait(); close(done) }()
	var once sync.Once
	stop := func() {
		once.Do(func() {
			_ = cmd.Process.Signal(os.Interrupt)
			timer := time.NewTimer(3 * time.Second)
			defer timer.Stop()
			select {
			case <-done:
			case <-timer.C:
				_ = cmd.Process.Kill()
				<-done
			}
			_ = log.Close()
			_ = os.RemoveAll(directory)
		})
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		connection, err := net.DialTimeout("unix", socket, 50*time.Millisecond)
		if err == nil {
			_ = connection.Close()
			return socket, stop, nil
		}
		select {
		case <-done:
			stop()
			return "", nil, fmt.Errorf("isolated Redis exited before becoming ready: %v", waitError)
		case <-deadline.C:
			stop()
			return "", nil, fmt.Errorf("isolated Redis startup timed out")
		case <-tick.C:
		}
	}
}
