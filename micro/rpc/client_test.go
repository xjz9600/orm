package rpc

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"orm/micro/rpc/compress"
	"orm/micro/rpc/message"
	"orm/micro/rpc/serialize/json"
	"testing"
)

func TestSetFuncField(t *testing.T) {
	testCases := []struct {
		name    string
		service Service
		mock    func(ctrl *gomock.Controller) Proxy
		wantErr error
	}{
		{
			name:    "nil",
			service: nil,
			wantErr: errors.New("rpc：不支持 nil"),
			mock: func(ctrl *gomock.Controller) Proxy {
				return NewMockProxy(ctrl)
			},
		},
		{
			name:    "no pointer",
			service: UserService{},
			wantErr: errors.New("rpc：只支持指向结构体的一级指针"),
			mock: func(ctrl *gomock.Controller) Proxy {
				return NewMockProxy(ctrl)
			},
		},
		{
			name:    "user service",
			service: &UserService{},
			mock: func(ctrl *gomock.Controller) Proxy {
				p := NewMockProxy(ctrl)
				p.EXPECT().Invoke(gomock.Any(), gomock.Any()).Return(&message.Response{}, nil)
				return p
			},
		},
	}
	s := &json.Serializer{}
	c := &compress.NoCompress{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			err := setFuncField(tc.service, tc.mock(ctrl), s, c)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			resp, err := tc.service.(*UserService).GetById(context.Background(), &GetByIdReq{Id: 123})
			assert.NoError(t, err)
			t.Log(resp)
		})
	}
}
