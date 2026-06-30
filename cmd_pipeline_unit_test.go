package redisson

import "testing"

// TestPipeline_DefaultCap 锁定 Pipeline() 默认预分配 cap = defaultPipelineCap.
// 防止后续重构无意中改回 nil slice 导致 hot path 多次 grow alloc.
func TestPipeline_DefaultCap(t *testing.T) {
	p := newPipeline(nil, 0) // hint=0 退化默认
	if cap(p.commands) != defaultPipelineCap {
		t.Errorf("Pipeline() commands cap = %d, want %d", cap(p.commands), defaultPipelineCap)
	}
	if cap(p.rets) != defaultPipelineCap {
		t.Errorf("Pipeline() rets cap = %d, want %d", cap(p.rets), defaultPipelineCap)
	}
	if len(p.commands) != 0 {
		t.Errorf("Pipeline() commands len = %d, want 0", len(p.commands))
	}
}

// TestPipeline_WithCap 验证 caller 显式 hint 透传到底层切片.
func TestPipeline_WithCap(t *testing.T) {
	for _, hint := range []int{1, 4, 16, 100} {
		p := newPipeline(nil, hint)
		if cap(p.commands) != hint {
			t.Errorf("newPipeline(_, %d) commands cap = %d, want %d", hint, cap(p.commands), hint)
		}
		if cap(p.rets) != hint {
			t.Errorf("newPipeline(_, %d) rets cap = %d, want %d", hint, cap(p.rets), hint)
		}
	}
}

// TestPipeline_NegativeHintFallsBack 负数/0 hint MUST 退化默认, 不触发 panic.
func TestPipeline_NegativeHintFallsBack(t *testing.T) {
	for _, hint := range []int{-1, -100, 0} {
		p := newPipeline(nil, hint)
		if cap(p.commands) != defaultPipelineCap {
			t.Errorf("newPipeline(_, %d) should fall back to default cap %d, got %d",
				hint, defaultPipelineCap, cap(p.commands))
		}
	}
}

// BenchmarkPipeline_DefaultCap_8Cmds 对比 Pipeline() 预分配 vs nil-slice
// append (旧实现) 在 8 条 cmd 场景下的 alloc 差异. Completed 零值合法 (alias
// to rueidis.Completed, struct type).
func BenchmarkPipeline_DefaultCap_8Cmds(b *testing.B) {
	var zeroCmd Completed
	for i := 0; i < b.N; i++ {
		p := newPipeline(nil, 0)
		for j := 0; j < 8; j++ {
			p.commands = append(p.commands, zeroCmd)
			p.rets = append(p.rets, nil)
		}
		_ = p
	}
}

func BenchmarkPipeline_NilSlice_8Cmds(b *testing.B) {
	var zeroCmd Completed
	for i := 0; i < b.N; i++ {
		var commands []Completed
		var rets []BaseCmd
		for j := 0; j < 8; j++ {
			commands = append(commands, zeroCmd)
			rets = append(rets, nil)
		}
		_ = commands
		_ = rets
	}
}

func BenchmarkPipeline_WithCap_8Cmds(b *testing.B) {
	var zeroCmd Completed
	for i := 0; i < b.N; i++ {
		p := newPipeline(nil, 8)
		for j := 0; j < 8; j++ {
			p.commands = append(p.commands, zeroCmd)
			p.rets = append(p.rets, nil)
		}
		_ = p
	}
}
