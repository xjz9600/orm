package registry

import (
	"context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error
	ListServices(ctx context.Context, name string) ([]ServiceInstance, error)
	Subscribe(name string) (<-chan Event, error)
	io.Closer
}

type ServiceInstance struct {
	Name   string
	Addr   string
	Weight uint32
	Group  string
}

type Event struct {
}
