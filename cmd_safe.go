package redisson

import (
	"context"
	"sync"
)

type SafeCmdable interface {
	// SafeMGet
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys to retrieve.
	// ACL categories: @read @string @fast
	// Like MGet, but safe in cluster mode.
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of values at the specified keys.
	SafeMGet(ctx context.Context, keys ...string) SliceCmd
}

func (c *client) SafeMGet(ctx context.Context, keys ...string) SliceCmd {
	ctx = WithSkipCheck(ctx)
	if len(keys) <= 1 {
		return c.MGet(ctx, keys...)
	}
	// slot2Keys：每个 slot 对应去重后的 key 列表（保持首次出现顺序）。
	// keyPos：每个 key 在原 keys 中所有出现位置；分发结果时对每个位置都赋值，
	// 避免重复 key 导致结果遗漏（旧实现以 map[key]int 仅记最后位置，res[0]=nil 等错误）。
	var slot2Keys = make(map[uint16][]string)
	var keyPos = make(map[string][]int)
	for i, key := range keys {
		if _, seen := keyPos[key]; !seen {
			slot2Keys[slot(key)] = append(slot2Keys[slot(key)], key)
		}
		keyPos[key] = append(keyPos[key], i)
	}
	if len(slot2Keys) == 1 {
		return c.MGet(ctx, keys...)
	}

	var mx sync.Mutex
	var scs = make(map[uint16]SliceCmd)

	parallelK(c.maxp, slot2Keys, func(k uint16) {
		// 透传外层 ctx，保留用户的取消/超时语义
		ret := c.MGet(WithSkipCheck(ctx), slot2Keys[k]...)
		mx.Lock()
		scs[k] = ret
		mx.Unlock()
	})

	var res = make([]any, len(keys))
	for s, ret := range scs {
		if err := ret.Err(); err != nil {
			return newSliceCmdFromSlice(nil, err, keys...)
		}
		_values := ret.Val()
		for _i, _key := range slot2Keys[s] {
			// 每个 key 可能在原 keys 中出现多次，全部位置都填充同一值。
			for _, p := range keyPos[_key] {
				res[p] = _values[_i]
			}
		}
	}
	return newSliceCmdFromSlice(res, nil, keys...)
}
