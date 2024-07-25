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
	var slot2Keys = make(map[uint16][]string)
	var keyIndexes = make(map[string]int)
	for i, key := range keys {
		keySlot := slot(key)
		slot2Keys[keySlot] = append(slot2Keys[keySlot], key)
		keyIndexes[key] = i
	}
	if len(slot2Keys) == 1 {
		return c.MGet(ctx, keys...)
	}
	var wg sync.WaitGroup
	var mx sync.Mutex
	var scs = make(map[uint16]SliceCmd)
	wg.Add(len(slot2Keys))
	for i, sameSlotKeys := range slot2Keys {
		go func(_i uint16, _keys []string) {
			ret := c.MGet(WithSkipCheck(context.Background()), _keys...)
			mx.Lock()
			scs[_i] = ret
			mx.Unlock()
			wg.Done()
		}(i, sameSlotKeys)
	}
	wg.Wait()

	var res = make([]any, len(keys))
	for i, ret := range scs {
		if err := ret.Err(); err != nil {
			return newSliceCmdFromSlice(nil, err, keys...)
		}
		_values := ret.Val()
		for _i, _key := range slot2Keys[i] {
			res[keyIndexes[_key]] = _values[_i]
		}
	}
	return newSliceCmdFromSlice(res, nil, keys...)
}
