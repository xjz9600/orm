package cache

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBuildInMapCache_Get(t *testing.T) {
	testCase := []struct {
		key     string
		name    string
		wantVal any
		wantErr error
		cache   *BuildInMapCache
	}{
		{
			name: "key not found",
			key:  "not exist key",
			cache: func() *BuildInMapCache {
				return NewBuildInMapCache(10 * time.Second)
			}(),
			wantErr: errKeyNotFound,
		},
		{
			name: "get value",
			key:  "key1",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "key1", 123, 10*time.Second)
				require.NoError(t, err)
				return res
			}(),
			wantVal: 123,
		},
		{
			name: "expired",
			key:  "expired key1",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "expired key1", 123, time.Second)
				require.NoError(t, err)
				time.Sleep(2 * time.Second)
				return res
			}(),
			wantErr: errKeyNotFound,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.cache.Get(context.Background(), tc.key)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, val, tc.wantVal)

		})
	}
}

func TestBuildInMapCache_Loop(t *testing.T) {
	var cnt int
	c := NewBuildInMapCache(time.Second, WithEvictedCallback(func(key string, val any) {
		cnt++
	}))
	err := c.Set(context.Background(), "key1", 123, time.Second)
	require.NoError(t, err)
	time.Sleep(3 * time.Second)
	c.mutex.RLock()
	_, ok := c.data["key1"]
	require.False(t, ok)
	require.Equal(t, 1, cnt)
}
