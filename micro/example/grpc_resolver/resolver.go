package grpc_resolver

import "google.golang.org/grpc/resolver"

type Builder struct {
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	//TODO implement me
	r := &Resolver{
		cc: cc,
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (b *Builder) Scheme() string {
	return "register"
}

type Resolver struct {
	cc resolver.ClientConn
}

func (r *Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	//TODO implement me
	// localhost:8090
	err := r.cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{
			{
				Addr: "localhost:8090",
			},
		},
	})
	if err != nil {
		r.cc.ReportError(err)
	}
}

func (r *Resolver) Close() {
	//TODO implement me
	panic("implement me")
}
