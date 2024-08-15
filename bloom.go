package redisson

import (
	"context"
	"errors"
	"github.com/redis/rueidis/rueidisprob"
)

// BloomFilter based on Redis Bitmaps.
// BloomFilter uses 128-bit murmur3 hash function.
type BloomFilter interface {
	// Add adds an item to the Bloom filter.
	Add(ctx context.Context, key string) error

	// AddMulti adds one or more items to the Bloom filter.
	// NOTE: If keys are too many, it can block the Redis server for a long time.
	AddMulti(ctx context.Context, keys []string) error

	// Exists checks if an item is in the Bloom filter.
	Exists(ctx context.Context, key string) (bool, error)

	// ExistsMulti checks if one or more items are in the Bloom filter.
	// Returns a slice of bool values where each bool indicates whether the corresponding key was found.
	ExistsMulti(ctx context.Context, keys []string) ([]bool, error)

	// Reset resets the Bloom filter.
	Reset(ctx context.Context) error

	// Delete deletes the Bloom filter.
	Delete(ctx context.Context) error

	// Count returns count of items in Bloom filter.
	Count(ctx context.Context) (uint64, error)
}

const bloomFilterROVersion = "7.0.0"

var errEnableReadOperationInvalidVersion = errors.New("if enabled read operation, minimum redis version should be " + bloomFilterROVersion)

// newBloomFilter 新键一个布隆过滤器
func newBloomFilter(c *client, name string, expectedNumberOfItems uint, falsePositiveRate float64, opts ...BloomOption) (BloomFilter, error) {
	// 校验版本
	cc := newBloomOptions(opts...)
	if cc.GetEnableReadOperation() && c.version.LessThan(mustNewSemVersion(bloomFilterROVersion)) {
		return nil, errEnableReadOperationInvalidVersion
	}
	return rueidisprob.NewBloomFilter(c.cmd, name, expectedNumberOfItems, falsePositiveRate, rueidisprob.WithEnableReadOperation(cc.GetEnableReadOperation()))
}

// NewBloomFilter 新键一个布隆过滤器
func (c *client) NewBloomFilter(name string, expectedNumberOfItems uint, falsePositiveRate float64, opts ...BloomOption) (BloomFilter, error) {
	return newBloomFilter(c, name, expectedNumberOfItems, falsePositiveRate, opts...)
}
