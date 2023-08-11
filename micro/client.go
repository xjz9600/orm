package micro

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"orm/micro/registry"
	"time"
)

type ClientOption func(*Client)

type Client struct {
	insecure bool
	r        registry.Registry
	timeout  time.Duration
	balancer balancer.Builder
}

func (c *Client) Dial(ctx context.Context, serviceName string, dialOption ...grpc.DialOption) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.r != nil {
		rb, err := NewRegistryBuilder(c.r, c.timeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	if c.balancer != nil {
		opts = append(opts, grpc.WithDefaultServiceConfig(
			fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`,
				c.balancer.Name())))
	}
	if len(dialOption) > 0 {
		opts = append(opts, dialOption...)
	}
	return grpc.DialContext(ctx, fmt.Sprintf("register:///%s", serviceName), opts...)
}

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func ClientInsecure() ClientOption {
	return func(client *Client) {
		client.insecure = true
	}
}

func ClientWithRegistry(r registry.Registry, timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.r = r
		client.timeout = timeout
	}
}

func ClientWithPickerBuilder(name string, b base.PickerBuilder) ClientOption {
	return func(client *Client) {
		builder := base.NewBalancerBuilder(name, b, base.Config{HealthCheck: true})
		balancer.Register(builder)
		client.balancer = builder
	}
}
