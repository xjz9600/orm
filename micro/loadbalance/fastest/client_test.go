package fastest

import (
	"context"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"orm/micro"
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
		micro.ClientWithPickerBuilder("prometheus", &Builder{
			Endpoint: "http://localhost:9090",
			Interval: time.Second * 3,
			Query:    "mySever{kind=\"user-service\",quantile=\"0.5\"}",
		}))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100000)
	defer cancel()
	cc, err := client.Dial(ctx, "user-service")
	uc := gen.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
		require.NoError(t, err)
		t.Log(resp)
		time.Sleep(5 * time.Second)
	}
}
