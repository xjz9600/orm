package leastactive

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync/atomic"
)

type Balancer struct {
	connects []*activeConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connects) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	res := &activeConn{
		cnt: math.MaxUint32,
	}
	for _, c := range b.connects {
		if res == nil || atomic.LoadUint32(&c.cnt) <= res.cnt {
			res = c
		}
	}
	atomic.AddUint32(&res.cnt, 1)
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint32(&res.cnt, -1)
		},
	}, nil
}

type Builder struct {
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*activeConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, &activeConn{
			c: c,
		})
	}
	return &Balancer{
		connects: connections,
	}
}

type activeConn struct {
	cnt uint32
	c   balancer.SubConn
}
