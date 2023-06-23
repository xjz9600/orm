package cache

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"orm/cache/mocks"
	"testing"
	"time"
)

func TestRedisCache_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testCases := []struct {
		name       string
		key        string
		mock       func() redis.Cmdable
		value      any
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "set value",
			mock: func() redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStatusCmd(context.Background())
				status.SetVal("OK")
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
		},
		{
			name: "timeout",
			mock: func() redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStatusCmd(context.Background())
				status.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "unexpected msg",
			mock: func() redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStatusCmd(context.Background())
				status.SetVal("NO OK")
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
			wantErr:    errFailedToSetCache,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewRedisCache(tc.mock())
			err := c.Set(context.Background(), tc.key, tc.value, tc.expiration)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testCases := []struct {
		name       string
		key        string
		mock       func() redis.Cmdable
		expiration time.Duration
		wantErr    error
		wantVal    string
	}{
		{
			name: "get value",
			mock: func() redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStringCmd(context.Background())
				status.SetVal("value1")
				cmd.EXPECT().Get(context.Background(), "key1").Return(status)
				return cmd
			},
			key:     "key1",
			wantVal: "value1",
		},
		{
			name: "timeout ",
			mock: func() redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStringCmd(context.Background())
				status.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Get(context.Background(), "key1").Return(status)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewRedisCache(tc.mock())
			val, err := c.Get(context.Background(), tc.key)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, val, tc.wantVal)
		})
	}
}
