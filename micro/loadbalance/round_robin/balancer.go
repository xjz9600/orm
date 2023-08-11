package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"orm/micro/loadbalance"
	"sync/atomic"
)

type Balancer struct {
	index    int32
	connects []subConn
	filter   loadbalance.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var candidates []subConn
	for _, c := range b.connects {
		if b.filter(info, c.info) {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	c := candidates[idx%int32(len(candidates))]
	return balancer.PickResult{
		SubConn: c.c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
	Filter loadbalance.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]subConn, 0, len(info.ReadySCs))
	for c, info := range info.ReadySCs {
		connections = append(connections, subConn{
			c:    c,
			info: info.Address,
		})
	}
	var filter loadbalance.Filter = func(info balancer.PickInfo, add resolver.Address) bool {
		return true
	}
	if b.Filter != nil {
		filter = b.Filter
	}

	return &Balancer{
		connects: connections,
		index:    -1,
		filter:   filter,
	}
}

type subConn struct {
	c    balancer.SubConn
	info resolver.Address
}
