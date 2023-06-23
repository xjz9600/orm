package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	errOverCapacity = errors.New("超过容量限制")
)

type MaxCntCache struct {
	BuildInMapCache
	cnt    int32
	maxCnt int32
}

type Option func(*MaxCntCache)

func NewMaxCntCache(c BuildInMapCache, maxCnt int32) *MaxCntCache {
	res := &MaxCntCache{
		BuildInMapCache: c,
		maxCnt:          maxCnt,
	}
	origin := c.onEvicted
	res.onEvicted = func(key string, val any) {
		atomic.AddInt32(&res.cnt, -1)
		if origin != nil {
			origin(key, val)
		}
	}
	return res
}

func (m *MaxCntCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.data[key]; !ok {
		if m.cnt+1 > m.maxCnt {
			return errOverCapacity
		}
		m.cnt++
	}
	return m.BuildInMapCache.set(key, val, expiration)
}
