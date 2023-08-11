package loadbalance

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

type Filter func(info balancer.PickInfo, add resolver.Address) bool

type GroupFilterBuilder struct {
}

func (g GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, add resolver.Address) bool {
		target := add.Attributes.Value("group")
		input := info.Ctx.Value("group")
		return target == input
	}
}
