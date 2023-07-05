package limit

import (
	"context"
	"time"
)

// CacheV1 屏蔽不同的缓存中间件的差异
type CacheV1 interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte,
		expiration time.Duration) error
	Delete(ctx context.Context, key string) error

	LoadAndDelete(ctx context.Context, key string) ([]byte, error)

	OnEvicted(func(key string, val []byte))
}
