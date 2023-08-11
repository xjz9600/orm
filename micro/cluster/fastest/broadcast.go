package broadcast

import (
	"context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"orm/micro/registry"
	"reflect"
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
		ok, ch := isBroadcast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		instances, err := c.registry.ListServices(ctx, c.serviceName)
		if err != nil {
			return err
		}
		defer close(ch)
		var eg errgroup.Group
		replyType := reflect.TypeOf(reply).Elem()
		for _, ins := range instances {
			addr := ins.Addr
			eg.Go(func() error {
				insCC, err := grpc.Dial(addr, c.dialOption...)
				var newResp Resp
				if err != nil {
					newResp = Resp{Err: err}
				} else {
					newReply := reflect.New(replyType).Interface()
					newResp = Resp{
						Err:   invoker(ctx, method, req, newReply, insCC, opts...),
						Reply: newReply,
					}
				}
				select {
				default:
					return err
				case ch <- newResp:
					return nil
				}
			})
		}
		return eg.Wait()
	}
}

func UseBroadcast(ctx context.Context) (context.Context, <-chan Resp) {
	ch := make(chan Resp)
	return context.WithValue(ctx, broadcastKey{}, ch), ch
}

type broadcastKey struct {
}

func isBroadcast(ctx context.Context) (bool, chan Resp) {
	val, ok := ctx.Value(broadcastKey{}).(chan Resp)
	return ok, val
}

type Resp struct {
	Err   error
	Reply any
}
