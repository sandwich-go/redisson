package redisson

import (
	"context"
	"sync"
)

type PipelineCmdable interface {
	Pipeline() Pipeliner
}

type Pipeliner interface {
	builder() builder

	Cmd(...Completed)
	Exec(context.Context) ([]any, error)
}

type pipelineCommand struct{}

func (pipelineCommand) String() string         { return "PIPELINE" }
func (pipelineCommand) Class() string          { return "Pipeline" }
func (pipelineCommand) RequireVersion() string { return "0.0.0" }
func (pipelineCommand) Forbid() bool           { return false }
func (pipelineCommand) WarnVersion() string    { return "0.0.0" }
func (pipelineCommand) Warning() string        { return "" }

var pipelineCmd = &pipelineCommand{}

type pipeline struct {
	client   *client
	commands []Completed

	mx sync.RWMutex
}

func (c *client) Pipeline() Pipeliner { return &pipeline{client: c} }

func (p *pipeline) builder() builder { return p.client.builder }
func (p *pipeline) Cmd(cs ...Completed) {
	if len(cs) == 0 {
		return
	}
	p.mx.Lock()
	p.commands = append(p.commands, cs...)
	p.mx.Unlock()
	return
}

func (p *pipeline) Exec(ctx context.Context) ([]any, error) {
	ctx = p.client.handler.before(ctx, pipelineCmd)

	var cmds []Completed
	p.mx.RLock()
	cmds = p.commands
	p.mx.RUnlock()

	var firstError error
	defer func() {
		p.client.handler.after(ctx, firstError)
	}()

	if len(cmds) == 0 {
		return nil, nil
	}

	var result = make([]any, len(cmds))
	if len(cmds) == 1 {
		r, err := p.client.cmd.Do(ctx, cmds[0]).ToAny()
		if err != nil {
			firstError = err
			result[0] = err
		} else {
			result[0] = r
		}
		return result, err
	}

	resps := p.client.cmd.DoMulti(ctx, cmds...)
	for i, resp := range resps {
		r, err := resp.ToAny()
		if err != nil && firstError == nil {
			firstError = err
			result[i] = err
		} else {
			result[i] = r
		}
	}
	return result, firstError
}
