package redisson

import "context"

type ClusterCmdable interface {
	// ClusterAddSlots
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of hash slot arguments
	// ACL categories: @admin @slow @dangerous
	ClusterAddSlots(ctx context.Context, slots ...int64) StatusCmd

	// ClusterAddSlotsRange
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of the slots between the start slot and end slot arguments.
	// ACL categories: @admin @slow @dangerous
	ClusterAddSlotsRange(ctx context.Context, min, max int64) StatusCmd

	// ClusterCountFailureReports
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the number of failure reports
	// ACL categories: @admin @slow @dangerous
	ClusterCountFailureReports(ctx context.Context, nodeID string) IntCmd

	// ClusterCountKeysInSlot
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @slow
	ClusterCountKeysInSlot(ctx context.Context, slot int64) IntCmd

	// ClusterDelSlots
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of hash slot arguments
	// ACL categories: @admin @slow @dangerous
	ClusterDelSlots(ctx context.Context, slots ...int64) StatusCmd

	// ClusterDelSlotsRange
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of the slots between the start slot and end slot arguments.
	// ACL categories: @admin @slow @dangerous
	ClusterDelSlotsRange(ctx context.Context, min, max int64) StatusCmd

	// ClusterFailover
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ClusterFailover(ctx context.Context) StatusCmd

	// ClusterForget
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ClusterForget(ctx context.Context, nodeID string) StatusCmd

	// ClusterGetKeysInSlot
	// Available since: 3.0.0
	// Time complexity: O(log(N)) where N is the number of requested keys
	// ACL categories: @slow
	ClusterGetKeysInSlot(ctx context.Context, slot int64, count int64) StringSliceCmd

	// ClusterInfo
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @slow
	ClusterInfo(ctx context.Context) StringCmd

	// ClusterKeySlot
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the number of bytes in the key
	// ACL categories: @slow
	ClusterKeySlot(ctx context.Context, key string) IntCmd

	// ClusterMeet
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ClusterMeet(ctx context.Context, host string, port int64) StatusCmd

	// ClusterNodes
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of Cluster nodes
	// ACL categories: @slow
	ClusterNodes(ctx context.Context) StringCmd

	// ClusterReplicate
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ClusterReplicate(ctx context.Context, nodeID string) StatusCmd

	// ClusterResetSoft
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the number of known nodes. The command may execute a FLUSHALL as a side effect.
	// ACL categories: @admin @slow @dangerous
	ClusterResetSoft(ctx context.Context) StatusCmd
	ClusterResetHard(ctx context.Context) StatusCmd

	// ClusterSaveConfig
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ClusterSaveConfig(ctx context.Context) StatusCmd

	// ClusterSlaves
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ClusterSlaves(ctx context.Context, nodeID string) StringSliceCmd

	// ClusterSlots
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of Cluster nodes
	// ACL categories: @slow
	ClusterSlots(ctx context.Context) ClusterSlotsCmd

	// ClusterShards
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of cluster nodes
	// ACL categories: @slow
	ClusterShards(ctx context.Context) ClusterShardsCmd

	// ReadOnly
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	ReadOnly(ctx context.Context) StatusCmd

	// ReadWrite
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	ReadWrite(ctx context.Context) StatusCmd
}

func (c *client) ClusterAddSlots(ctx context.Context, slots ...int64) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterAddSlots)
	r := c.adapter.ClusterAddSlots(ctx, slots...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterAddSlotsRange(ctx context.Context, min, max int64) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterAddSlotsRange)
	r := c.adapter.ClusterAddSlotsRange(ctx, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterCountFailureReports(ctx context.Context, nodeID string) IntCmd {
	ctx = c.handler.before(ctx, CommandClusterCountFailureReports)
	r := c.adapter.ClusterCountFailureReports(ctx, nodeID)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterCountKeysInSlot(ctx context.Context, slot int64) IntCmd {
	ctx = c.handler.before(ctx, CommandClusterCountKeysInSlot)
	r := c.adapter.ClusterCountKeysInSlot(ctx, slot)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterDelSlots(ctx context.Context, slots ...int64) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterDelSlots)
	r := c.adapter.ClusterDelSlots(ctx, slots...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterDelSlotsRange(ctx context.Context, min, max int64) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterDelSlotsRange)
	r := c.adapter.ClusterDelSlotsRange(ctx, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterFailover(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterFailover)
	r := c.adapter.ClusterFailover(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterForget(ctx context.Context, nodeID string) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterForget)
	r := c.adapter.ClusterForget(ctx, nodeID)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterGetKeysInSlot(ctx context.Context, slot int64, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandClusterGetKeysInSlot)
	r := c.adapter.ClusterGetKeysInSlot(ctx, slot, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterInfo(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandClusterInfo)
	r := c.adapter.ClusterInfo(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterKeySlot(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandClusterKeySlot)
	r := c.adapter.ClusterKeySlot(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterMeet(ctx context.Context, host string, port int64) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterMeet)
	r := c.adapter.ClusterMeet(ctx, host, port)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterNodes(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandClusterNodes)
	r := c.adapter.ClusterNodes(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterReplicate(ctx context.Context, nodeID string) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterReplicate)
	r := c.adapter.ClusterReplicate(ctx, nodeID)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterResetSoft(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterResetSoft)
	r := c.adapter.ClusterResetSoft(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterResetHard(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterResetHard)
	r := c.adapter.ClusterResetHard(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterSaveConfig(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandClusterSaveConfig)
	r := c.adapter.ClusterSaveConfig(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterSlaves(ctx context.Context, nodeID string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandClusterSlaves)
	r := c.adapter.ClusterSlaves(ctx, nodeID)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterSlots(ctx context.Context) ClusterSlotsCmd {
	ctx = c.handler.before(ctx, CommandClusterSlots)
	r := c.adapter.ClusterSlots(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClusterShards(ctx context.Context) ClusterShardsCmd {
	ctx = c.handler.before(ctx, CommandClusterShards)
	r := c.adapter.ClusterShards(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ReadOnly(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandReadOnly)
	r := c.adapter.ReadOnly(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ReadWrite(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandReadWrite)
	r := c.adapter.ReadWrite(ctx)
	c.handler.after(ctx, r.Err())
	return r
}
