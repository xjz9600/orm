package cache

import (
	"context"
	"golang.org/x/sync/singleflight"
	"time"
)

type SignleflightCache struct {
	ReadThroughCache
}

func NewSingleflightCache(cache Cache, LoadFunc func(ctx context.Context, key string) (any, error), expiration time.Duration) *SignleflightCache {
	g := &singleflight.Group{}
	return &SignleflightCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				val, err, _ := g.Do(key, func() (interface{}, error) {
					return LoadFunc(ctx, key)
				})
				return val, err
			},
			Expiration: expiration,
		},
	}
}
