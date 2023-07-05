//go:build e2e

package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestClient_e2e_TryLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	testCases := []struct {
		name       string
		before     func(t *testing.T)
		key        string
		expiration time.Duration
		wantErr    error
		wantLock   *Lock
		after      func(t *testing.T)
	}{
		{
			name: "key exit",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "values1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, res, "values1")
			},
			key:        "key1",
			expiration: time.Minute,
			wantErr:    ErrFailedToPreemptLock,
		},
		{
			name: "locked",
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.NotEmpty(t, res)
			},
			before: func(t *testing.T) {

			},
			key:        "key1",
			expiration: time.Minute,
			wantLock: &Lock{
				key: "key1",
			},
		},
	}
	client := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			lock, err := client.TryLock(ctx, tc.key, tc.expiration)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.NotEmpty(t, lock.value)
			assert.NotNil(t, lock.client)
		})
	}
}

func TestLock_e2e_Unlock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		wantErr  error
		wantLock *Lock
		after    func(t *testing.T)
	}{
		{
			name: "lock not hold",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			wantLock: &Lock{
				key:    "key1",
				value:  "123",
				client: rdb,
			},
			wantErr: ErrLockNotExist,
		},
		{
			name: "lock hold by others",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "values1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, res, "values1")
			},
			wantLock: &Lock{
				key:    "key1",
				value:  "123",
				client: rdb,
			},
			wantErr: ErrLockNotExist,
		},
		{
			name: "unlocked",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "values1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Exists(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, res, int64(0))
			},
			wantLock: &Lock{
				key:    "key1",
				value:  "values1",
				client: rdb,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			err := tc.wantLock.Unlock(ctx)
			assert.Equal(t, err, tc.wantErr)
		})
	}
}

func TestLock_e2e_Refresh(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		wantErr  error
		wantLock *Lock
		after    func(t *testing.T)
	}{
		{
			name: "lock not hold",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			wantLock: &Lock{
				key:        "key1",
				value:      "123",
				client:     rdb,
				expiration: time.Minute,
			},
			wantErr: ErrLockNotExist,
		},
		{
			name: "lock hold by others",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "values1", time.Second*10).Result()
				require.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.TTL(ctx, "key1").Result()
				require.NoError(t, err)
				require.True(t, res <= time.Second*10)
				num, err := rdb.Del(ctx, "key1").Result()
				require.NoError(t, err)
				require.Equal(t, num, int64(1))
			},
			wantLock: &Lock{
				key:        "key1",
				value:      "123",
				client:     rdb,
				expiration: time.Minute,
			},
			wantErr: ErrLockNotExist,
		},
		{
			name: "unlocked",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "values1", time.Second*10).Result()
				require.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.TTL(ctx, "key1").Result()
				require.NoError(t, err)
				require.True(t, res > time.Second*10)
				num, err := rdb.Del(ctx, "key1").Result()
				require.NoError(t, err)
				require.Equal(t, num, int64(1))
			},
			wantLock: &Lock{
				key:        "key1",
				value:      "values1",
				client:     rdb,
				expiration: time.Minute,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			err := tc.wantLock.Refresh(ctx)
			assert.Equal(t, err, tc.wantErr)
		})
	}
}

func TestLock_e2e_lock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	testCases := []struct {
		name       string
		before     func(t *testing.T)
		key        string
		expiration time.Duration
		wantErr    error
		wantLock   *Lock
		timeout    time.Duration
		retry      RetryStrategy
		after      func(t *testing.T)
	}{
		{
			name: "locked",
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.TTL(ctx, "key1").Result()
				require.NoError(t, err)
				require.True(t, res >= time.Second*50)
				num, err := rdb.Del(ctx, "key1").Result()
				require.NoError(t, err)
				require.Equal(t, num, int64(1))
			},
			before: func(t *testing.T) {

			},
			key:        "key1",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &FixedIntervalRetrySrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wantLock: &Lock{
				key:        "key1",
				expiration: time.Minute,
			},
		},
		{
			name: "others hold lock",
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, res, "values1")
			},
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "values1", time.Second*10).Result()
				require.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			key:        "key1",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &FixedIntervalRetrySrategy{
				Interval: time.Second,
				MaxCnt:   3,
			},
			wantErr: fmt.Errorf("redis-lock: 超出重试限制，%w", ErrFailedToPreemptLock),
		},
		{
			name: "retry and locked",
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.TTL(ctx, "key1").Result()
				require.NoError(t, err)
				require.True(t, res >= time.Second*50)
				num, err := rdb.Del(ctx, "key1").Result()
				require.NoError(t, err)
				require.Equal(t, num, int64(1))
			},
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "values1", time.Second*3).Result()
				require.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			key:        "key1",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &FixedIntervalRetrySrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wantLock: &Lock{
				key:        "key1",
				expiration: time.Minute,
			},
		},
	}
	client := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			lock, err := client.Lock(ctx, tc.key, tc.expiration, tc.timeout, tc.retry)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.NotEmpty(t, lock.value)
			assert.NotNil(t, lock.client)
		})
	}
}
