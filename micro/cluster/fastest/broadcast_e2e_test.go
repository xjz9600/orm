package broadcast

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"orm/micro"
	"orm/micro/proto/gen"
	"orm/micro/registry/etcd"
	"testing"
	"time"
)

func TestUseBroadcast(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	var eg errgroup.Group
	var servers []*Server
	for i := 0; i < 3; i++ {
		server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r))
		require.NoError(t, err)
		us := &Server{idx: i}
		servers = append(servers, us)
		gen.RegisterUserServiceServer(server, us)
		require.NoError(t, err)
		port := fmt.Sprintf(":809%d", i+1)
		eg.Go(func() error {
			return server.Start(port)
		})
	}
	time.Sleep(time.Second * 3)
	client, err := micro.NewClient(micro.ClientInsecure(),
		micro.ClientWithRegistry(r, time.Second*3))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx, respChan := UseBroadcast(ctx)
	go func() {
		res := <-respChan
		fmt.Println(res.Err)
		fmt.Println(res.Reply)
	}()
	bd := NewClusterBuilder(r, "user-service", grpc.WithInsecure())
	cc, err := client.Dial(ctx, "user-service", grpc.WithUnaryInterceptor(bd.BuildUnaryInterceptor()))
	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
	require.NoError(t, err)
	t.Log(resp)
	for _, s := range servers {
		require.Equal(t, 1, s.cnt)
	}
}

type Server struct {
	cnt int
	idx int
	gen.UnimplementedUserServiceServer
}

func (s *Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	s.cnt++
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: fmt.Sprintf("hello,world %d", s.idx),
		},
	}, nil
}
