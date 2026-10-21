package redisson

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/redis/rueidis"
)

// Two protocol-speaking nodes keep the original connection open when a slot
// moves. This exercises Rueidis' real cluster routing and SUNSUBSCRIBE handling.
type shardTestCluster struct {
	mu     sync.Mutex
	nodes  [2]*miniredis.Miniredis
	owner  int
	peers  [2]map[string]map[*server.Peer]struct{}
	starts [2]int
}

func newShardTestCluster(t *testing.T) *shardTestCluster {
	t.Helper()
	c := &shardTestCluster{}
	for i := range c.nodes {
		c.nodes[i] = miniredis.RunT(t)
		c.peers[i] = make(map[string]map[*server.Peer]struct{})
		index := i
		c.nodes[i].Server().SetPreHook(func(p *server.Peer, cmd string, args ...string) bool {
			return c.command(index, p, strings.ToUpper(cmd), args)
		})
	}
	return c
}

func (c *shardTestCluster) command(index int, p *server.Peer, cmd string, args []string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch cmd {
	case "HELLO":
		// Keep the simulated cluster on Redis 7's CLUSTER SLOTS protocol,
		// regardless of the Redis version advertised by miniredis itself.
		proto, _ := strconv.Atoi(args[0])
		p.Resp3 = proto == 3
		p.WriteMapLen(3)
		p.WriteBulk("proto")
		p.WriteInt(proto)
		p.WriteBulk("version")
		p.WriteBulk("7.0.0")
		p.WriteBulk("mode")
		p.WriteBulk("cluster")
	case "INFO":
		p.WriteBulk("redis_version:7.0.0\r\ncluster_enabled:1\r\n")
	case "CLUSTER":
		if len(args) == 0 || strings.ToUpper(args[0]) != "SLOTS" {
			p.WriteError("ERR unsupported CLUSTER command")
			return true
		}
		host, portString, _ := net.SplitHostPort(c.nodes[c.owner].Addr())
		port, _ := strconv.Atoi(portString)
		p.WriteLen(1)
		p.WriteLen(3)
		p.WriteInt(0)
		p.WriteInt(16383)
		p.WriteLen(3)
		p.WriteBulk(host)
		p.WriteInt(port)
		p.WriteBulk(fmt.Sprintf("node-%d", c.owner))
	case "SSUBSCRIBE":
		if index != c.owner {
			// The tests use foo, whose Redis hash slot is 12182.
			p.WriteError("MOVED 12182 " + c.nodes[c.owner].Addr())
			return true
		}
		for _, channel := range args {
			if c.peers[index][channel] == nil {
				c.peers[index][channel] = make(map[*server.Peer]struct{})
			}
			c.peers[index][channel][p] = struct{}{}
			c.starts[index]++
			shardTestPush(p, "ssubscribe", channel, "")
		}
	case "SUNSUBSCRIBE":
		for _, channel := range args {
			delete(c.peers[index][channel], p)
			shardTestPush(p, "sunsubscribe", channel, "")
		}
	case "SPUBLISH":
		if index != c.owner {
			p.WriteError("MOVED 12182 " + c.nodes[c.owner].Addr())
			return true
		}
		for peer := range c.peers[index][args[0]] {
			shardTestPush(peer, "smessage", args[0], args[1])
		}
		p.WriteInt(len(c.peers[index][args[0]]))
	default:
		return false
	}
	return true
}

func shardTestPush(p *server.Peer, kind, channel, message string) {
	p.Block(func(w *server.Writer) {
		w.WritePushLen(3)
		w.WriteBulk(kind)
		w.WriteBulk(channel)
		if kind == "smessage" {
			w.WriteBulk(message)
		} else {
			w.WriteInt(1)
		}
	})
	p.Flush()
}

func (c *shardTestCluster) move() {
	c.mu.Lock()
	defer c.mu.Unlock()
	old := c.owner
	c.owner = 1 - old
	for channel, peers := range c.peers[old] {
		for peer := range peers {
			shardTestPush(peer, "sunsubscribe", channel, "")
		}
	}
	c.peers[old] = make(map[string]map[*server.Peer]struct{})
}

func (c *shardTestCluster) subscribers(channel string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.peers[c.owner][channel])
}

func waitShardTest(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for !condition() {
		select {
		case <-deadline.C:
			t.Fatal("timed out waiting for shard subscription")
		case <-tick.C:
		}
	}
}

func shardTestClient(t *testing.T, cluster *shardTestCluster, resp2 bool) Cmdable {
	t.Helper()
	c, err := Connect(NewConf(
		WithAddrs(cluster.nodes[0].Addr()), WithEnableCache(false),
		WithAlwaysRESP2(resp2), WithWriteTimeout(time.Second),
	))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if mode := c.(*client).cmd.Mode(); mode != rueidis.ClientModeCluster {
		t.Fatalf("expected a cluster client, got %v", mode)
	}
	return c
}

func assertShardMessage(t *testing.T, c Cmdable, p PubSub, message string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.SPublish(ctx, "foo", message).Err(); err != nil {
		t.Fatal(err)
	}
	select {
	case m, ok := <-p.Channel():
		if !ok || m.Channel != "foo" || m.Message != message {
			t.Fatalf("unexpected message: %+v, open: %v", m, ok)
		}
	case <-ctx.Done():
		t.Fatal("did not receive shard message")
	}
}

func TestShardPubSubSlotMigration(t *testing.T) {
	for _, resp2 := range []bool{false, true} {
		t.Run(fmt.Sprintf("RESP2=%v", resp2), func(t *testing.T) {
			cluster := newShardTestCluster(t)
			c := shardTestClient(t, cluster, resp2)
			p := c.SSubscribe(context.Background(), "foo")
			t.Cleanup(func() { _ = p.Close() })
			waitShardTest(t, func() bool { return cluster.subscribers("foo") == 1 })
			assertShardMessage(t, c, p, "before migration")
			for i := 0; i < 2; i++ {
				cluster.move()
				waitShardTest(t, func() bool { return cluster.subscribers("foo") == 1 })
				assertShardMessage(t, c, p, fmt.Sprintf("after migration %d", i))
			}
			// Race explicit unsubscribe against a migration-triggered retry.
			cluster.move()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := p.SUnsubscribe(ctx, "foo"); err != nil {
				t.Fatal(err)
			}
			time.Sleep(150 * time.Millisecond)
			if n := cluster.subscribers("foo"); n != 0 {
				t.Fatalf("explicitly unsubscribed channel was restored: %d", n)
			}
			if err := p.SSubscribe(ctx, "foo"); err != nil {
				t.Fatal(err)
			}
			waitShardTest(t, func() bool { return cluster.subscribers("foo") == 1 })
			assertShardMessage(t, c, p, "subscribed again")
			if err := p.Close(); err != nil {
				t.Fatal(err)
			}
			waitShardTest(t, func() bool { return cluster.subscribers("foo") == 0 })
			select {
			case _, ok := <-p.Channel():
				if ok {
					t.Fatal("channel remained open after Close")
				}
			case <-ctx.Done():
				t.Fatal("Close did not close the message channel")
			}
		})
	}
}

func TestShardPubSubPartialUnsubscribe(t *testing.T) {
	cluster := newShardTestCluster(t)
	c := shardTestClient(t, cluster, false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	p := c.SSubscribe(ctx, "foo", "foo{other}")
	t.Cleanup(func() { _ = p.Close() })
	waitShardTest(t, func() bool {
		return cluster.subscribers("foo") == 1 && cluster.subscribers("foo{other}") == 1
	})
	if err := p.SUnsubscribe(ctx, "foo{other}"); err != nil {
		t.Fatal(err)
	}
	cluster.move()
	waitShardTest(t, func() bool { return cluster.subscribers("foo") == 1 })
	assertShardMessage(t, c, p, "remaining channel")
	if n := cluster.subscribers("foo{other}"); n != 0 {
		t.Fatal("removed channel was restored after migration")
	}
	if err := p.SUnsubscribe(ctx); err != nil {
		t.Fatal(err)
	}
	if n := cluster.subscribers("foo"); n != 0 {
		t.Fatal("unsubscribe-all left a server subscription")
	}
}

func TestShardPubSubCancellationDuringMigration(t *testing.T) {
	for _, closePubSub := range []bool{false, true} {
		t.Run(fmt.Sprintf("Close=%v", closePubSub), func(t *testing.T) {
			cluster := newShardTestCluster(t)
			c := shardTestClient(t, cluster, false)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := c.SSubscribe(ctx, "foo").(*pubSub)
			t.Cleanup(func() { _ = p.Close() })
			waitShardTest(t, func() bool { return cluster.subscribers("foo") == 1 })
			// Duplicate calls must not start an additional receiver.
			if err := p.SSubscribe(ctx, "foo"); err != nil {
				t.Fatal(err)
			}
			p.shardMu.Lock()
			done := p.shards["foo"].done
			p.shardMu.Unlock()
			cluster.mu.Lock()
			starts := cluster.starts[0]
			cluster.mu.Unlock()
			if starts != 1 {
				t.Fatalf("duplicate receivers started: %d", starts)
			}
			cluster.move()
			if closePubSub {
				_ = p.Close()
			} else {
				cancel()
			}
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("receiver did not exit after cancellation")
			}
			if n := cluster.subscribers("foo"); n != 0 {
				t.Fatalf("canceled subscription remains active: %d", n)
			}
			if err := p.SSubscribe(context.Background(), "foo"); err != context.Canceled {
				t.Fatalf("want context.Canceled after shutdown, got %v", err)
			}
		})
	}
}

func TestPubSubCloseMessageChannel(t *testing.T) {
	for _, pattern := range []bool{false, true} {
		t.Run(fmt.Sprintf("pattern=%v", pattern), func(t *testing.T) {
			// This miniredis version only permits Pub/Sub commands on subscribed
			// connections, so use RESP2's dedicated subscription connection.
			s := miniredis.RunT(t)
			s.Server().SetPreHook(func(p *server.Peer, cmd string, _ ...string) bool {
				if strings.EqualFold(cmd, "INFO") {
					p.WriteBulk("redis_version:7.0.0\r\ncluster_enabled:0\r\n")
					return true
				}
				return false
			})
			c, err := Connect(NewConf(WithAddrs(s.Addr()), WithDevelopment(false), WithEnableCache(false), WithAlwaysRESP2(true)))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var p PubSub
			if pattern {
				p = c.PSubscribe(ctx, "f*")
			} else {
				p = c.Subscribe(ctx, "foo")
			}
			defer p.Close()
			waitShardTest(t, func() bool { return c.Publish(ctx, "foo", "message").Val() == 1 })
			select {
			case m := <-p.Channel():
				if m.Message != "message" {
					t.Fatalf("unexpected message: %+v", m)
				}
			case <-ctx.Done():
				t.Fatal("did not receive message")
			}
			_ = p.Close()
			select {
			case _, ok := <-p.Channel():
				if ok {
					t.Fatal("message channel remained open after Close")
				}
			case <-ctx.Done():
				t.Fatal("message channel did not close")
			}
		})
	}
}
