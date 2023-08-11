package broadcast

import (
	"context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"orm/micro/registry"
)

type ClusterBuilder struct {
	registry    registry.Registry
	serviceName string
	dialOption  []grpc.DialOption
}

func NewClusterBuilder(registry registry.Registry, serviceName string, opts ...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		registry:    registry,
		serviceName: serviceName,
		dialOption:  opts,
	}
}

func (c *ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !isBroadcast(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		instances, err := c.registry.ListServices(ctx, c.serviceName)
		if err != nil {
			return err
		}
		var eg errgroup.Group
		for _, ins := range instances {
			addr := ins.Addr
			eg.Go(func() error {
				insCC, err := grpc.Dial(addr, c.dialOption...)
				if err != nil {
					return err
				}
				return invoker(ctx, method, req, reply, insCC, opts...)
			})
		}
		return eg.Wait()
	}
}

func UseBroadcast(ctx context.Context) context.Context {
	return context.WithValue(ctx, broadcastKey{}, true)
}

type broadcastKey struct {
}

func isBroadcast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)
	return ok && val
}
