package redisson

import (
	"context"
	"github.com/redis/rueidis"
	"sync"
)

type PipelineCmdable interface {
	Pipeline() Pipeliner
}

type Pipeliner interface {
	Put(ctx context.Context, cmd Command, keys []string, args ...interface{}) error
	Exec(ctx context.Context) ([]interface{}, error)
}

type pipelineCommand struct{}

func (pipelineCommand) String() string         { return "PIPELINE" }
func (pipelineCommand) Class() string          { return "Pipeline" }
func (pipelineCommand) RequireVersion() string { return "0.0.0" }
func (pipelineCommand) Forbid() bool           { return false }
func (pipelineCommand) WarnVersion() string    { return "" }
func (pipelineCommand) Warning() string        { return "" }
func (pipelineCommand) Cmd() []string          { return nil }

var pipelineCmd = &pipelineCommand{}

type pipeCommand struct {
	cmd  []string
	keys []string
	args []interface{}
}

func (p pipeCommand) getCompleted(c *client) rueidis.Completed {
	return c.cmd.B().Arbitrary(p.cmd...).Keys(p.keys...).Args(argsToSlice(p.args)...).Build()
}

func (p pipeCommand) exec(ctx context.Context, c *client) (v interface{}, err error) {
	return c.cmd.Do(ctx, p.getCompleted(c)).ToAny()
}

type pipeline struct {
	client   *client
	commands []pipeCommand
	mx       sync.RWMutex
}

func (c *client) Pipeline() Pipeliner { return &pipeline{client: c} }

func (p *pipeline) Put(_ context.Context, cmd Command, keys []string, args ...interface{}) (err error) {
	p.mx.Lock()
	p.commands = append(p.commands, pipeCommand{cmd: cmd.Cmd(), keys: keys, args: args})
	p.mx.Unlock()
	return
}

func (p *pipeline) Exec(ctx context.Context) ([]interface{}, error) {
	ctx = p.client.handler.before(ctx, pipelineCmd)

	var cmds []pipeCommand
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

	var result = make([]interface{}, len(cmds))
	if len(cmds) == 1 {
		r, err := cmds[0].exec(ctx, p.client)
		if err != nil {
			firstError = err
			result[0] = err
		} else {
			result[0] = r
		}
		return result, err
	}

	var cs = make([]rueidis.Completed, 0, len(cmds))
	for _, cmd := range cmds {
		cs = append(cs, cmd.getCompleted(p.client))
	}
	resps := p.client.cmd.DoMulti(ctx, cs...)
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
