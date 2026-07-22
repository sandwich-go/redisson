package redisson

import "testing"

func TestConfVisitor2ClientOptionPipelineMultiplex(t *testing.T) {
	const want = 3
	cfg := NewConf(WithPipelineMultiplex(want))

	got := confVisitor2ClientOption(cfg).PipelineMultiplex
	if got != want {
		t.Fatalf("PipelineMultiplex = %d, want %d", got, want)
	}
}
