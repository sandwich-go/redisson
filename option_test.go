package redisson

import (
	"testing"
	"time"
)

// TestRevise_FillsDefaults 零值字段会被 revise 填回默认值。
func TestRevise_FillsDefaults(t *testing.T) {
	conf := NewConf(
		WithWriteTimeout(0),
		WithPubSubChanSize(0),
		WithBootstrapTimeout(0),
	)
	revise(conf)
	if conf.GetWriteTimeout() != defaultWriteTimeout {
		t.Errorf("WriteTimeout=%v, want %v", conf.GetWriteTimeout(), defaultWriteTimeout)
	}
	if conf.GetPubSubChanSize() != defaultPubSubChanSize {
		t.Errorf("PubSubChanSize=%d, want %d", conf.GetPubSubChanSize(), defaultPubSubChanSize)
	}
	if conf.GetBootstrapTimeout() != defaultBootstrapTimeout {
		t.Errorf("BootstrapTimeout=%v, want %v", conf.GetBootstrapTimeout(), defaultBootstrapTimeout)
	}
}

// TestRevise_NonZeroPreserved 已显式设置的非零值不会被覆盖。
func TestRevise_NonZeroPreserved(t *testing.T) {
	conf := NewConf(
		WithWriteTimeout(7*time.Second),
		WithPubSubChanSize(42),
		WithBootstrapTimeout(15*time.Second),
	)
	revise(conf)
	if conf.GetWriteTimeout() != 7*time.Second {
		t.Errorf("WriteTimeout overwritten: %v", conf.GetWriteTimeout())
	}
	if conf.GetPubSubChanSize() != 42 {
		t.Errorf("PubSubChanSize overwritten: %d", conf.GetPubSubChanSize())
	}
	if conf.GetBootstrapTimeout() != 15*time.Second {
		t.Errorf("BootstrapTimeout overwritten: %v", conf.GetBootstrapTimeout())
	}
}

// TestNewConf_Defaults NewConf 默认值校验。
func TestNewConf_Defaults(t *testing.T) {
	conf := NewConf()
	if conf.GetNet() != "tcp" {
		t.Errorf("Net=%q", conf.GetNet())
	}
	if !conf.GetEnableMonitor() {
		t.Errorf("EnableMonitor default should be true")
	}
	if !conf.GetEnableCache() {
		t.Errorf("EnableCache default should be true")
	}
	if conf.GetDevelopment() {
		t.Errorf("Development default should be false")
	}
	if len(conf.GetAddrs()) != 1 || conf.GetAddrs()[0] != "127.0.0.1:6379" {
		t.Errorf("Addrs default=%v", conf.GetAddrs())
	}
}

// TestApplyOption_RoundTrip ApplyOption 返回的旧值可以再 ApplyOption 还原。
func TestApplyOption_RoundTrip(t *testing.T) {
	conf := NewConf(WithName("orig"))
	prev := conf.ApplyOption(WithName("new"))
	if conf.GetName() != "new" {
		t.Errorf("after ApplyOption Name=%q, want new", conf.GetName())
	}
	conf.ApplyOption(prev...)
	if conf.GetName() != "orig" {
		t.Errorf("after revert Name=%q, want orig", conf.GetName())
	}
}

// TestNewConf_OverrideAddrs 多 addrs 列表生效。
func TestNewConf_OverrideAddrs(t *testing.T) {
	conf := NewConf(WithAddrs("a:1", "b:2", "c:3"))
	if len(conf.GetAddrs()) != 3 {
		t.Errorf("Addrs=%v", conf.GetAddrs())
	}
}
