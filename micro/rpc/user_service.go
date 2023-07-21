package rpc

import (
	"context"
	"orm/micro/proto/gen"
	"time"
)

type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)

	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
	Err error
	Msg string
}

func (u UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: u.Msg,
		},
	}, u.Err

}

func (u UserServiceServer) Name() string {
	return "user-service"
}

type UserServiceServerTimeout struct {
	Err   error
	Msg   string
	sleep time.Duration
}

func (u UserServiceServerTimeout) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	if _, ok := ctx.Deadline(); !ok {
		panic("必须设置超时时间")
	}
	time.Sleep(u.sleep)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u UserServiceServerTimeout) Name() string {
	return "user-service"
}
