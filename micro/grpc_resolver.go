package micro

import (
	"context"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"orm/micro/registry"
	"time"
)

type grpcResolverBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func NewRegistryBuilder(r registry.Registry, timeout time.Duration) (*grpcResolverBuilder, error) {
	return &grpcResolverBuilder{r: r, timeout: timeout}, nil
}

func (b *grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	//TODO implement me
	r := &grpcResolver{
		cc:      cc,
		r:       b.r,
		target:  target,
		timeout: b.timeout,
	}
	r.resolve()
	go r.watch()
	return r, nil
}

func (b *grpcResolverBuilder) Scheme() string {
	return "register"
}

type grpcResolver struct {
	cc      resolver.ClientConn
	r       registry.Registry
	target  resolver.Target
	timeout time.Duration
	close   chan struct{}
}

func (r *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.resolve()
}

func (r *grpcResolver) watch() {
	event, err := r.r.Subscribe(r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	select {
	case <-event:
		r.resolve()
	case <-r.close:
		return
	}
}

func (r *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	instances, err := r.r.ListServices(ctx, r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instances))
	for _, si := range instances {
		address = append(address, resolver.Address{Addr: si.Addr, Attributes: attributes.New("weight", si.Weight).WithValue("group", si.Group)})
	}
	err = r.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		r.cc.ReportError(err)
		return
	}
}

func (r *grpcResolver) Close() {
	close(r.close)
}
