package redisson

import "context"

type ClusterCmdable interface {
	// ClusterAddSlots
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of hash slot arguments
	// ACL categories: @admin @slow @dangerous
	//
	// This command is useful in order to modify a node's view of the cluster configuration.
	// Specifically it assigns a set of hash slots to the node receiving the command. If the command is successful,
	// the node will map the specified hash slots to itself, and will start broadcasting the new configuration.
	//
	// However note that:
	// 	1. The command only works if all the specified slots are, from the point of view of the node receiving the command,
	//		currently not assigned. A node will refuse to take ownership for slots that already belong to some other node (including itself).
	// 	2. The command fails if the same slot is specified multiple times.
	// 	3. As a side effect of the command execution, if a slot among the ones specified as argument is set as importing,
	//		this state gets cleared once the node assigns the (previously unbound) slot to itself.
	//
	// RESP2/RESP3 Reply
	//	Simple string reply: OK if the command was successful. Otherwise an error is returned.
	ClusterAddSlots(ctx context.Context, slots ...int64) StatusCmd

	// ClusterAddSlotsRange
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of the slots between the start slot and end slot arguments.
	// ACL categories: @admin @slow @dangerous
	//
	// The CLUSTER ADDSLOTSRANGE is similar to the CLUSTER ADDSLOTS command in that they both assign hash slots to nodes.
	//
	// The difference between the two commands is that ADDSLOTS takes a list of slots to assign to the node,
	// while ADDSLOTSRANGE takes a list of slot ranges (specified by start and end slots) to assign to the node.
	//
	// RESP2/RESP3 Reply
	//	Simple string reply: OK if the command was successful. Otherwise an error is returned.
	ClusterAddSlotsRange(ctx context.Context, min, max int64) StatusCmd

	// ClusterCountFailureReports
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the number of failure reports
	// ACL categories: @admin @slow @dangerous
	//
	// The command returns the number of failure reports for the specified node.
	// Failure reports are the way Redis Cluster uses in order to promote a PFAIL state,
	// that means a node is not reachable, to a FAIL state, that means that the majority of masters in the cluster
	// agreed within a window of time that the node is not reachable.
	//
	// A few more details:
	//	- A node flags another node with PFAIL when the node is not reachable for a time greater than the configured node timeout,
	// 		which is a fundamental configuration parameter of a Redis Cluster.
	//	- Nodes in PFAIL state are provided in gossip sections of heartbeat packets.
	//	- Every time a node processes gossip packets from other nodes, it creates (and refreshes the TTL if needed) failure reports,
	//		remembering that a given node said another given node is in PFAIL condition.
	//	- Each failure report has a time to live of two times the node timeout time.
	// 	- If at a given time a node has another node flagged with PFAIL, and at the same time collected the majority of other master
	//		nodes failure reports about this node (including itself if it is a master), then it elevates the failure state of the node
	//		from PFAIL to FAIL, and broadcasts a message forcing all the nodes that can be reached to flag the node as FAIL.
	//
	// This command returns the number of failure reports for the current node which are currently not expired (so received within two times the node timeout time).
	// The count does not include what the node we are asking this count believes about the node ID we pass as argument, the count only includes the failure reports
	// the node received from other nodes.
	//
	// This command is mainly useful for debugging, when the failure detector of Redis Cluster is not operating as we believe it should.
	//
	// RESP2/RESP3 Reply
	//	Integer reply: the number of active failure reports for the node.
	ClusterCountFailureReports(ctx context.Context, nodeID string) IntCmd

	// ClusterCountKeysInSlot
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @slow
	//
	// Returns the number of keys in the specified Redis Cluster hash slot. The command only queries the local data set,
	// so contacting a node that is not serving the specified hash slot will always result in a count of zero being returned.
	//
	// RESP2/RESP3 Reply
	//	Integer reply: The number of keys in the specified hash slot, or an error if the hash slot is invalid.
	ClusterCountKeysInSlot(ctx context.Context, slot int64) IntCmd

	// ClusterDelSlots
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of hash slot arguments
	// ACL categories: @admin @slow @dangerous
	//
	// In Redis Cluster, each node keeps track of which master is serving a particular hash slot.
	//
	// The CLUSTER DELSLOTS command asks a particular Redis Cluster node to forget which master is serving the hash slots specified as arguments.
	//
	// In the context of a node that has received a CLUSTER DELSLOTS command and has consequently removed the associations for the passed hash slots,
	// we say those hash slots are unbound. Note that the existence of unbound hash slots occurs naturally when a node has not been configured to handle
	// them (something that can be done with the CLUSTER ADDSLOTS command) and if it has not received any information about who owns those hash slots
	// (something that it can learn from heartbeat or update messages).
	//
	// If a node with unbound hash slots receives a heartbeat packet from another node that claims to be the owner of some of those hash slots, the association
	// is established instantly. Moreover, if a heartbeat or update message is received with a configuration epoch greater than the node's own, the association is re-established.
	//
	// However, note that:
	//	1. The command only works if all the specified slots are already associated with some node.
	//	2. The command fails if the same slot is specified multiple times.
	//	3. As a side effect of the command execution, the node may go into down state because not all hash slots are covered.
	//
	// RESP2/RESP3 Reply
	//	Simple string reply: OK if the command was successful. Otherwise an error is returned.
	ClusterDelSlots(ctx context.Context, slots ...int64) StatusCmd

	// ClusterDelSlotsRange
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of the slots between the start slot and end slot arguments.
	// ACL categories: @admin @slow @dangerous
	//
	// The CLUSTER DELSLOTSRANGE command is similar to the CLUSTER DELSLOTS command in that they both remove hash slots from the node.
	// The difference is that CLUSTER DELSLOTS takes a list of hash slots to remove from the node, while CLUSTER DELSLOTSRANGE takes a
	// list of slot ranges (specified by start and end slots) to remove from the node.
	//
	// RESP2/RESP3 Reply
	//	Simple string reply: OK if the command was successful. Otherwise an error is returned.
	ClusterDelSlotsRange(ctx context.Context, min, max int64) StatusCmd

	// ClusterFailover
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	//
	// This command, that can only be sent to a Redis Cluster replica node, forces the replica to start a manual failover of its master instance.
	//
	// A manual failover is a special kind of failover that is usually executed when there are no actual failures, but we wish to swap the current
	// master with one of its replicas (which is the node we send the command to), in a safe way, without any window for data loss. It works in the following way:
	//	1. The replica tells the master to stop processing queries from clients.
	//	2. The master replies to the replica with the current replication offset.
	//	3. The replica waits for the replication offset to match on its side, to make sure it processed all the data from the master before it continues.
	//	4. The replica starts a failover, obtains a new configuration epoch from the majority of the masters, and broadcasts the new configuration.
	//	5. The old master receives the configuration update: unblocks its clients and starts replying with redirection messages so that they'll continue the chat with the new master.
	//
	// This way clients are moved away from the old master to the new master atomically and only when the replica that is turning into the new master has processed all of the replication stream from the old master.
	//
	// RESP2/RESP3 Reply
	//	Simple string reply: OK if the command was accepted and a manual failover is going to be attempted. An error if the operation cannot be executed, for example if the client is connected to a node that is already a master.
	ClusterFailover(ctx context.Context) StatusCmd

	// ClusterForget
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	//
	// The command is used in order to remove a node, specified via its node ID, from the set of known nodes of the Redis Cluster node receiving the command.
	// In other words the specified node is removed from the nodes table of the node receiving the command.
	//
	// Because when a given node is part of the cluster, all the other nodes participating in the cluster knows about it, in order for a node to be completely
	// removed from a cluster, the CLUSTER FORGET command must be sent to all the remaining nodes, regardless of the fact they are masters or replicas.
	//
	// However the command cannot simply drop the node from the internal node table of the node receiving the command, it also implements a ban-list, not allowing
	// the same node to be added again as a side effect of processing the gossip section of the heartbeat packets received from other nodes.
	//
	// RESP2/RESP3 Reply
	//	Simple string reply: OK if the command was executed successfully, otherwise an error is returned.
	ClusterForget(ctx context.Context, nodeID string) StatusCmd

	// ClusterGetKeysInSlot
	// Available since: 3.0.0
	// Time complexity: O(log(N)) where N is the number of requested keys
	// ACL categories: @slow
	//
	// The command returns an array of keys names stored in the contacted node and hashing to the specified hash slot.
	// The maximum number of keys to return is specified via the count argument, so that it is possible for the user of this API to batch-processing keys.
	//
	// The main usage of this command is during rehashing of cluster slots from one node to another. The way the rehashing is performed is exposed in the Redis Cluster specification,
	// or in a more simple to digest form, as an appendix of the CLUSTER SETSLOT command documentation.
	//
	// RESP2/RESP3 Reply
	//	Array reply: From 0 to count key names in a Redis array reply.
	ClusterGetKeysInSlot(ctx context.Context, slot int64, count int64) StringSliceCmd

	// ClusterInfo
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @slow
	//
	// CLUSTER INFO provides INFO style information about Redis Cluster vital parameters. The following fields are always present in the reply:
	//  - cluster_state: State is ok if the node is able to receive queries. fail if there is at least one hash slot which is unbound (no node associated), in error state (node serving it is flagged with FAIL flag), or if the majority of masters can't be reached by this node.
	//	- cluster_slots_assigned: Number of slots which are associated to some node (not unbound). This number should be 16384 for the node to work properly, which means that each hash slot should be mapped to a node.
	//	- cluster_slots_ok: Number of hash slots mapping to a node not in FAIL or PFAIL state.
	//	- cluster_slots_pfail: Number of hash slots mapping to a node in PFAIL state. Note that those hash slots still work correctly, as long as the PFAIL state is not promoted to FAIL by the failure detection algorithm. PFAIL only means that we are currently not able to talk with the node, but may be just a transient error.
	//	- cluster_slots_fail: Number of hash slots mapping to a node in FAIL state. If this number is not zero the node is not able to serve queries unless cluster-require-full-coverage is set to no in the configuration.
	//	- cluster_known_nodes: The total number of known nodes in the cluster, including nodes in HANDSHAKE state that may not currently be proper members of the cluster.
	//	- cluster_size: The number of master nodes serving at least one hash slot in the cluster.
	//	- cluster_current_epoch: The local Current Epoch variable. This is used in order to create unique increasing version numbers during fail overs.
	//	- cluster_my_epoch: The Config Epoch of the node we are talking with. This is the current configuration version assigned to this node.
	//	- cluster_stats_messages_sent: Number of messages sent via the cluster node-to-node binary bus.
	//	- cluster_stats_messages_received: Number of messages received via the cluster node-to-node binary bus.
	//	- total_cluster_links_buffer_limit_exceeded: Accumulated count of cluster links freed due to exceeding the cluster-link-sendbuf-limit configuration.
	//
	// The following message-related fields may be included in the reply if the value is not 0: Each message type includes statistics on the number of messages sent and received. Here are the explanation of these fields:
	//
	//	- cluster_stats_messages_ping_sent and cluster_stats_messages_ping_received: Cluster bus PING (not to be confused with the client command PING).
	//	- cluster_stats_messages_pong_sent and cluster_stats_messages_pong_received: PONG (reply to PING).
	//	- cluster_stats_messages_meet_sent and cluster_stats_messages_meet_received: Handshake message sent to a new node, either through gossip or CLUSTER MEET.
	//	- cluster_stats_messages_fail_sent and cluster_stats_messages_fail_received: Mark node xxx as failing.
	//	- cluster_stats_messages_publish_sent and cluster_stats_messages_publish_received: Pub/Sub Publish propagation, see Pubsub.
	//	- cluster_stats_messages_auth-req_sent and cluster_stats_messages_auth-req_received: Replica initiated leader election to replace its master.
	//	- cluster_stats_messages_auth-ack_sent and cluster_stats_messages_auth-ack_received: Message indicating a vote during leader election.
	//	- cluster_stats_messages_update_sent and cluster_stats_messages_update_received: Another node slots configuration.
	//	- cluster_stats_messages_mfstart_sent and cluster_stats_messages_mfstart_received: Pause clients for manual failover.
	//	- cluster_stats_messages_module_sent and cluster_stats_messages_module_received: Module cluster API message.
	//	- cluster_stats_messages_publishshard_sent and cluster_stats_messages_publishshard_received: Pub/Sub Publish shard propagation, see Sharded Pubsub.
	//
	// RESP2 Reply
	// 	Bulk string reply: A map between named fields and values in the form of <field>:<value> lines separated by newlines composed by the two bytes CRLF.
	//
	// RESP3 Reply
	//  Bulk string reply: A map between named fields and values in the form of : lines separated by newlines composed by the two bytes CRLF
	ClusterInfo(ctx context.Context) StringCmd

	// ClusterKeySlot
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the number of bytes in the key
	// ACL categories: @slow
	//
	// Returns an integer identifying the hash slot the specified key hashes to. This command is mainly useful for debugging and testing,
	// since it exposes via an API the underlying Redis implementation of the hashing algorithm. Example use cases for this command:
	//	1. Client libraries may use Redis in order to test their own hashing algorithm, generating random keys and hashing them with both their local implementation and using Redis CLUSTER KEYSLOT command, then checking if the result is the same.
	//	2. Humans may use this command in order to check what is the hash slot, and then the associated Redis Cluster node, responsible for a given key.
	//
	// RESP2/RESP3 Reply
	//	Integer reply: The hash slot number.
	ClusterKeySlot(ctx context.Context, key string) IntCmd

	// ClusterMeet
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	//
	// CLUSTER MEET is used in order to connect different Redis nodes with cluster support enabled, into a working cluster.
	//
	// The basic idea is that nodes by default don't trust each other, and are considered unknown, so that it is unlikely that different cluster
	// nodes will mix into a single one because of system administration errors or network addresses modifications.
	//
	// So in order for a given node to accept another one into the list of nodes composing a Redis Cluster, there are only two ways:
	// 	1. The system administrator sends a CLUSTER MEET command to force a node to meet another one.
	// 	2. An already known node sends a list of nodes in the gossip section that we are not aware of. If the receiving node trusts the sending node as a known node, it will process the gossip section and send a handshake to the nodes that are still not known.
	//
	// Note that Redis Cluster needs to form a full mesh (each node is connected with each other node), but in order to create a cluster, there is no need to send all
	// the CLUSTER MEET commands needed to form the full mesh. What matter is to send enough CLUSTER MEET messages so that each node can reach each other node through
	// a chain of known nodes. Thanks to the exchange of gossip information in heartbeat packets, the missing links will be created.
	//
	// RESP2/RESP3 Reply
	//	Simple string reply: OK if the command was successful. If the address or port specified are invalid an error is returned.
	//
	// History
	//	Starting with Redis version 4.0.0: Added the optional cluster_bus_port argument.
	ClusterMeet(ctx context.Context, host string, port int64) StatusCmd

	// ClusterNodes
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of Cluster nodes
	// ACL categories: @slow
	// See https://redis.io/commands/cluster-nodes/
	ClusterNodes(ctx context.Context) StringCmd

	// ClusterReplicate
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// The command reconfigures a node as a replica of the specified master. If the node receiving the command is an empty master, as a side effect of the command, the node role is changed from master to replica.
	// Once a node is turned into the replica of another master node, there is no need to inform the other cluster nodes about the change: heartbeat packets exchanged between nodes will propagate the new configuration automatically.
	// A replica will always accept the command, assuming that:
	//	The specified node ID exists in its nodes table.
	//	The specified node ID does not identify the instance we are sending the command to.
	//	The specified node ID is a master.
	// If the node receiving the command is not already a replica, but is a master, the command will only succeed, and the node will be converted into a replica, only if the following additional conditions are met:
	//	The node is not serving any hash slots.
	//	The node is empty, no keys are stored at all in the key space.
	// If the command succeeds the new replica will immediately try to contact its master in order to replicate from it.
	// Return:
	//	Simple string reply: OK if the command was executed successfully, otherwise an error is returned.
	ClusterReplicate(ctx context.Context, nodeID string) StatusCmd

	// ClusterResetSoft
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the number of known nodes. The command may execute a FLUSHALL as a side effect.
	// ACL categories: @admin @slow @dangerous
	// Reset a Redis Cluster node, in a more or less drastic way depending on the reset type, that can be hard or soft. Note that this command does not work for masters if they hold one or more keys, in that case to completely reset a master node keys must be removed first, e.g. by using FLUSHALL first, and then CLUSTER RESET.
	// Effects on the node:
	//	All the other nodes in the cluster are forgotten.
	//	All the assigned / open slots are reset, so the slots-to-nodes mapping is totally cleared.
	//	If the node is a replica it is turned into an (empty) master. Its dataset is flushed, so at the end the node will be an empty master.
	//	Hard reset only: a new Node ID is generated.
	//	Hard reset only: currentEpoch and configEpoch vars are set to 0.
	//	The new configuration is persisted on disk in the node cluster configuration file.
	// This command is mainly useful to re-provision a Redis Cluster node in order to be used in the context of a new, different cluster. The command is also extensively used by the Redis Cluster testing framework in order to reset the state of the cluster every time a new test unit is executed.
	// If no reset type is specified, the default is soft.
	// Return:
	//	Simple string reply: OK if the command was successful. Otherwise an error is returned.
	ClusterResetSoft(ctx context.Context) StatusCmd

	// ClusterResetHard
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the number of known nodes. The command may execute a FLUSHALL as a side effect.
	// ACL categories: @admin @slow @dangerous
	// See ClusterResetSoft
	ClusterResetHard(ctx context.Context) StatusCmd

	// ClusterSaveConfig
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// Forces a node to save the nodes.conf configuration on disk. Before to return the command calls fsync(2) in order to make sure the configuration is flushed on the computer disk.
	// This command is mainly used in the event a nodes.conf node state file gets lost / deleted for some reason, and we want to generate it again from scratch. It can also be useful in case of mundane alterations of a node cluster configuration via the CLUSTER command in order to ensure the new configuration is persisted on disk, however all the commands should normally be able to auto schedule to persist the configuration on disk when it is important to do so for the correctness of the system in the event of a restart.
	// Return:
	//	Simple string reply: OK or an error if the operation fails.
	ClusterSaveConfig(ctx context.Context) StatusCmd

	// ClusterSlaves
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// As of Redis version 5.0.0, this command is regarded as deprecated.
	// It can be replaced by CLUSTER REPLICAS when migrating or writing new code.
	// A note about the word slave used in this man page and command name: starting with Redis version 5, if not for backward compatibility, the Redis project no longer uses the word slave. Please use the new command CLUSTER REPLICAS. The command CLUSTER SLAVES will continue to work for backward compatibility.
	// The command provides a list of replica nodes replicating from the specified master node. The list is provided in the same format used by CLUSTER NODES (please refer to its documentation for the specification of the format).
	// The command will fail if the specified node is not known or if it is not a master according to the node table of the node receiving the command.
	// Note that if a replica is added, moved, or removed from a given master node, and we ask CLUSTER SLAVES to a node that has not yet received the configuration update, it may show stale information. However eventually (in a matter of seconds if there are no network partitions) all the nodes will agree about the set of nodes associated with a given master.
	// Return:
	//	The command returns data in the same format as CLUSTER NODES.
	ClusterSlaves(ctx context.Context, nodeID string) StringSliceCmd

	// ClusterSlots
	// Available since: 3.0.0
	// Time complexity: O(N) where N is the total number of Cluster nodes
	// ACL categories: @slow
	// As of Redis version 7.0.0, this command is regarded as deprecated.
	// It can be replaced by CLUSTER SHARDS when migrating or writing new code.
	// CLUSTER SLOTS returns details about which cluster slots map to which Redis instances. The command is suitable to be used by Redis Cluster client libraries implementations in order to retrieve (or update when a redirection is received) the map associating cluster hash slots with actual nodes network information, so that when a command is received, it can be sent to what is likely the right instance for the keys specified in the command.
	// The networking information for each node is an array containing the following elements:
	//	Preferred endpoint (Either an IP address, hostname, or NULL)
	//	Port number
	//	The node ID
	//	A map of additional networking metadata
	// The preferred endpoint, along with the port, defines the location that clients should use to send requests for a given slot. A NULL value for the endpoint indicates the node has an unknown endpoint and the client should connect to the same endpoint it used to send the CLUSTER SLOTS command but with the port returned from the command. This unknown endpoint configuration is useful when the Redis nodes are behind a load balancer that Redis doesn't know the endpoint of. Which endpoint is set as preferred is determined by the cluster-preferred-endpoint-type config.
	// Additional networking metadata is provided as a map on the fourth argument for each node. The following networking metadata may be returned:
	//	IP: When the preferred endpoint is not set to IP.
	//	Hostname: When a node has an announced hostname but the primary endpoint is not set to hostname.
	// Nested Result Array
	// Each nested result is:
	//	Start slot range
	//	End slot range
	//	Master for slot range represented as nested networking information
	//	First replica of master for slot range
	//	Second replica
	//	...continues until all replicas for this master are returned.
	// Each result includes all active replicas of the master instance for the listed slot range. Failed replicas are not returned.
	// The third nested reply is guaranteed to be the networking information of the master instance for the slot range. All networking information after the third nested reply are replicas of the master.
	// If a cluster instance has non-contiguous slots (e.g. 1-400,900,1800-6000) then master and replica networking information results will be duplicated for each top-level slot range reply.
	// Return:
	//	Array reply: nested list of slot ranges with networking information.
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
	// Enables read queries for a connection to a Redis Cluster replica node.
	// Normally replica nodes will redirect clients to the authoritative master for the hash slot involved in a given command, however clients can use replicas in order to scale reads using the READONLY command.
	// READONLY tells a Redis Cluster replica node that the client is willing to read possibly stale data and is not interested in running write queries.
	// When the connection is in readonly mode, the cluster will send a redirection to the client only if the operation involves keys not served by the replica's master node. This may happen because:
	// 	The client sent a command about hash slots never served by the master of this replica.
	// 	The cluster was reconfigured (for example resharded) and the replica is no longer able to serve commands for a given hash slot.
	// Return:
	//	Simple string reply
	ReadOnly(ctx context.Context) StatusCmd

	// ReadWrite
	// Available since: 3.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// Disables read queries for a connection to a Redis Cluster replica node.
	// Read queries against a Redis Cluster replica node are disabled by default, but you can use the READONLY command to change this behavior on a per- connection basis. The READWRITE command resets the readonly mode flag of a connection back to readwrite.
	// Return:
	// 	Simple string reply
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
