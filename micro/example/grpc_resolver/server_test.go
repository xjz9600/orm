package grpc_resolver

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"orm/micro"
	"orm/micro/proto/gen"
	"orm/micro/registry/etcd"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	var eg errgroup.Group
	for i := 0; i < 3; i++ {
		var group string = "A"
		if i%2 == 0 {
			group = "B"
		}
		server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r), micro.ServerWithGroup(group))
		require.NoError(t, err)
		us := &Server{group: group}
		gen.RegisterUserServiceServer(server, us)
		require.NoError(t, err)
		port := fmt.Sprintf(":809%d", i+1)
		eg.Go(func() error {
			return server.Start(port)
		})
	}
	eg.Wait()
	t.Log(err)
}

type Server struct {
	group string
	gen.UnimplementedUserServiceServer
}

func (s *Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello,world",
		},
	}, nil
}
