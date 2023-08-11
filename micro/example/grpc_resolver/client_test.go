package grpc_resolver

import (
	"context"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"orm/micro"
	"orm/micro/loadbalance"
	"orm/micro/loadbalance/round_robin"
	"orm/micro/proto/gen"
	"orm/micro/registry/etcd"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	client, err := micro.NewClient(micro.ClientInsecure(),
		micro.ClientWithRegistry(r, time.Second*3),
		micro.ClientWithPickerBuilder("GROUP_ROUND_ROBIN", &round_robin.Builder{
			Filter: (loadbalance.GroupFilterBuilder{}).Build(),
		}))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cc, err := client.Dial(ctx, "user-service")
	ctx = context.WithValue(ctx, "group", "A")
	uc := gen.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
		require.NoError(t, err)
		t.Log(resp)
	}
}
