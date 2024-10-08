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
	cmd(Completed, BaseCmd)

	// Exec 执行，返回结果集
	// 如果有错误，则 error 不为 nil，除 rueidis.Nil 错误外
	// 返回的结果集为数组，与执行的命令顺序有关，如果有 error，则数组中也包含 error，除 rueidis.Nil 错误外
	Exec(context.Context) ([]any, error)
	// ExecCmds 执行，返回命令结果集
	// 如果有错误，则 error 不为 nil，除 rueidis.Nil 错误外
	// 命令结果集的类型，可以通过命令的 PR 函数获得，例如 CommandAppend.PR()
	ExecCmds(context.Context) ([]BaseCmd, error)
}

type pipelineCommand struct{}

func (pipelineCommand) String() string         { return "PIPELINE" }
func (pipelineCommand) Class() string          { return "Pipeline" }
func (pipelineCommand) RequireVersion() string { return "0.0.0" }
func (pipelineCommand) Forbid() bool           { return false }
func (pipelineCommand) WarnVersion() string    { return "" }
func (pipelineCommand) Warning() string        { return "" }
func (pipelineCommand) WarningOnce() bool      { return false }
func (pipelineCommand) Instead() string        { return "" }
func (pipelineCommand) ETC() string            { return "" }

var pipelineCmd = &pipelineCommand{}

type pipeline struct {
	client   *client
	commands []Completed
	rets     []BaseCmd

	mx sync.RWMutex
}

func (c *client) Pipeline() Pipeliner { return &pipeline{client: c} }

func (p *pipeline) builder() builder { return p.client.builder }
func (p *pipeline) cmd(cs Completed, ret BaseCmd) {
	p.mx.Lock()
	p.commands = append(p.commands, cs)
	p.rets = append(p.rets, ret)
	p.mx.Unlock()
	return
}

func (p *pipeline) exec(ctx context.Context, f func([]Completed, []BaseCmd) error) {
	ctx = p.client.handler.before(ctx, pipelineCmd)

	var cmds []Completed
	var rets []BaseCmd
	p.mx.RLock()
	cmds = p.commands
	rets = p.rets
	p.mx.RUnlock()

	var firstError error
	defer func() {
		p.client.handler.after(ctx, firstError)
	}()

	if len(cmds) == 0 {
		return
	}
	firstError = f(cmds, rets)
	return
}

func (p *pipeline) Exec(ctx context.Context) (result []any, err error) {
	p.exec(ctx, func(cmds []Completed, _ []BaseCmd) error {
		result = make([]any, len(cmds))
		if len(cmds) == 1 {
			result[0], err = p.client.cmd.Do(ctx, cmds[0]).ToAny()
			if err != nil && result[0] == nil {
				result[0] = err
			}
		} else {
			for i, resp := range p.client.cmd.DoMulti(ctx, cmds...) {
				var err0 error
				result[i], err0 = resp.ToAny()
				if err0 == nil {
					continue
				}
				if err == nil {
					err = err0
				}
				if result[i] == nil {
					result[i] = err0
				}
			}
		}
		return err
	})
	return
}

func (p *pipeline) ExecCmds(ctx context.Context) (rets []BaseCmd, err error) {
	p.exec(ctx, func(cmds []Completed, in []BaseCmd) error {
		rets = in
		if len(cmds) == 1 {
			resp := p.client.cmd.Do(ctx, cmds[0])
			err = resp.NonRedisError()
			rets[0].(fromRedisResult).from(resp)
		} else {
			for i, resp := range p.client.cmd.DoMulti(ctx, cmds...) {
				if err0 := resp.NonRedisError(); err0 != nil && err == nil {
					err = err0
				}
				rets[i].(fromRedisResult).from(resp)
			}
		}
		return err
	})
	return
}
