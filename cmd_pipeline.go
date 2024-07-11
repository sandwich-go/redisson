package redisson

import (
	"context"
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

func (p pipeCommand) exec(ctx context.Context, c *client) (v interface{}, err error) {
	return c.cmd.Do(ctx, c.cmd.B().Arbitrary(p.cmd...).Keys(p.keys...).Args(argsToSlice(p.args)...).Build()).ToAny()
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
		return result, firstError
	}

	var wg sync.WaitGroup
	wg.Add(len(cmds))
	for i, cmd := range cmds {
		go func(_i int, _cmd pipeCommand) {
			r, err := _cmd.exec(ctx, p.client)
			if err != nil {
				if firstError == nil {
					firstError = err
				}
				result[_i] = err
			} else {
				result[_i] = r
			}
			wg.Done()
		}(i, cmd)
	}
	wg.Wait()
	return result, firstError
}
