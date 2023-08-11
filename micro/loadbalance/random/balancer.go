package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math/rand"
	"orm/micro/loadbalance"
)

type Balancer struct {
	connects []subConn
	filter   loadbalance.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var candidates []subConn
	for _, c := range b.connects {
		if b.filter != nil && !b.filter(info, c.info) {
			continue
		}
		candidates = append(candidates, c)
	}
	idx := rand.Intn(len(candidates))
	return balancer.PickResult{
		SubConn: candidates[idx].c,
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
	return &Balancer{
		connects: connections,
		filter:   b.Filter,
	}
}

type subConn struct {
	c    balancer.SubConn
	info resolver.Address
}
