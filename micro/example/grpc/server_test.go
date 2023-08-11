package grpc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"net"
	"orm/micro/proto/gen"
	"testing"
)

func TestServer(t *testing.T) {
	us := &Server{}
	server := grpc.NewServer()
	gen.RegisterUserServiceServer(server, us)
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	err = server.Serve(l)
	t.Log(err)
}

type Server struct {
	gen.UnimplementedUserServiceServer
}

func (s *Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	//TODO implement me
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello,world",
		},
	}, nil
}
