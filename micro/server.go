package micro

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"orm/micro/registry"
	"time"
)

type Server struct {
	registry        registry.Registry
	registryTimeout time.Duration
	*grpc.Server
	name   string
	weight uint32
	group  string
}

type ServerOption func(server *Server)

func NewServer(name string, opts ...ServerOption) (*Server, error) {
	res := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registryTimeout: time.Second * 10,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func ServerWithRegistry(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = r
	}
}

func ServerWithWeight(weight uint32) ServerOption {
	return func(server *Server) {
		server.weight = weight
	}
}

func ServerGrpcUnaryServerInterceptor(serverOption ...grpc.ServerOption) ServerOption {
	return func(server *Server) {
		server.Server = grpc.NewServer(serverOption...)
	}
}

func ServerWithGroup(group string) ServerOption {
	return func(server *Server) {
		server.group = group
	}
}

func (s *Server) Start(addr string, opts ...grpc.ServerOption) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registryTimeout)
		defer cancel()
		err = s.registry.Register(ctx, registry.ServiceInstance{
			Name:   s.name,
			Addr:   listener.Addr().String(),
			Weight: s.weight,
			Group:  s.group,
		})
		if err != nil {
			return err
		}
	}
	return s.Serve(listener)
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}
