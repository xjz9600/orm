package rpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"orm/micro/proto/gen"
	"orm/micro/rpc/compress/gz"
	"orm/micro/rpc/serialize/proto"
	"testing"
	"time"
)

func TestInitServiceJson(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8085")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}
	client, err := NewClient("localhost:8085")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello,world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello,world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = ""
			},
			wantErr:  errors.New("mock error"),
			wantResp: &GetByIdResp{},
		},
		{
			name: "both",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello,world"
			},
			wantErr: errors.New("mock error"),
			wantResp: &GetByIdResp{
				Msg: "hello,world",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, err := usClient.GetById(context.Background(), &GetByIdReq{Id: 234})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestInitServiceProto(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerialize(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8086")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}
	client, err := NewClient("localhost:8086", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello,world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello,world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = ""
			},
			wantErr:  errors.New("mock error"),
			wantResp: &GetByIdResp{},
		},
		{
			name: "both",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello,world"
			},
			wantErr: errors.New("mock error"),
			wantResp: &GetByIdResp{
				Msg: "hello,world",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, err := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
		})
	}
}

func TestOneway(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerialize(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8087")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}
	client, err := NewClient("localhost:8087", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *gen.GetByIdResp
	}{
		{
			name: "one way",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello,world"
			},
			wantErr: errors.New("micro：这时一个 oneway 调用，你不需要处理任何结果"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, err := usClient.GetByIdProto(CtxWithOneWay(context.Background()), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestTimeout(t *testing.T) {
	server := NewServer()
	service := &UserServiceServerTimeout{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8088")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}
	client, err := NewClient("localhost:8088")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		mock     func() context.Context
		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "timeout",
			mock: func() context.Context {
				service.Err = errors.New("mock error")
				service.Msg = "hello,world"
				service.sleep = time.Second * 2
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				return ctx
			},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := usClient.GetById(tc.mock(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestInitServiceGzipCompress(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerialize(&proto.Serializer{})
	server.RegisterCompress(&gz.GzipCompress{})
	go func() {
		err := server.Start("tcp", ":8089")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}
	client, err := NewClient("localhost:8089", ClientWithSerializer(&proto.Serializer{}), ClientWithCompress(&gz.GzipCompress{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello,world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello,world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = ""
			},
			wantErr:  errors.New("mock error"),
			wantResp: &GetByIdResp{},
		},
		{
			name: "both",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello,world"
			},
			wantErr: errors.New("mock error"),
			wantResp: &GetByIdResp{
				Msg: "hello,world",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, err := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
		})
	}
}
