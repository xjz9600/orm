package cache

import "context"

type BloomFilterCache struct {
	ReadThroughCache
}

func NewBloomFilterCache(cache Cache, bf BloomFilter, loadFunc func(ctx context.Context, key string) (any, error)) *BloomFilterCache {
	return &BloomFilterCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				if !bf.HashKey(ctx, key) {
					return nil, errKeyNotFound
				}
				return loadFunc(ctx, key)
			},
		},
	}
}

type BloomFilter interface {
	HashKey(ctx context.Context, key string) bool
}
