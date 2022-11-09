package redisson

import (
	"context"
)

type PipelineCmdable interface {
	Pipeline() Pipeliner
}

type Pipeliner interface {
	Put(ctx context.Context, cmd Command, keys []string, args ...interface{}) error
	Exec(ctx context.Context) ([]interface{}, error)
}

func (c *client) Pipeline() Pipeliner     { return c.cmdable.Pipeline() }
func (c *client) RawCmdable() interface{} { return c.cmdable.RawCmdable() }
