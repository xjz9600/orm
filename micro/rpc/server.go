package rpc

import (
	"context"
	"errors"
	"net"
	"orm/micro/rpc/compress"
	"orm/micro/rpc/message"
	"orm/micro/rpc/serialize"
	"orm/micro/rpc/serialize/json"
	"reflect"
	"strconv"
	"time"
)

type Server struct {
	services  map[string]reflectionStub
	serialize map[uint8]serialize.Serialize
	compress  map[uint8]compress.Compress
}

func NewServer() *Server {
	res := &Server{
		services:  make(map[string]reflectionStub, 16),
		serialize: make(map[uint8]serialize.Serialize, 4),
		compress:  make(map[uint8]compress.Compress, 4),
	}
	json := &json.Serializer{}
	res.RegisterSerialize(json)
	noCompress := &compress.NoCompress{}
	res.RegisterCompress(noCompress)
	return res
}

func (s *Server) RegisterSerialize(serialize serialize.Serialize) {
	s.serialize[serialize.Code()] = serialize
}

func (s *Server) RegisterCompress(compress compress.Compress) {
	s.compress[compress.Code()] = compress
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		value:     reflect.ValueOf(service),
		serialize: s.serialize,
		compress:  s.compress,
	}
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if err := s.handleConn(conn); err != nil {
				conn.Close()
			}
		}()
	}

}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}
		req := message.DecodeReq(reqBs)
		ctx := context.Background()
		cancel := func() {}
		oneWay, ok := req.Meta["one-way"]
		if ok && oneWay == "true" {
			ctx = CtxWithOneWay(ctx)
		}
		if deadlineStr, ok := req.Meta["deadline"]; ok {
			deadline, err := strconv.ParseInt(deadlineStr, 10, 64)
			if err == nil {
				t := time.UnixMilli(deadline)
				ctx, cancel = context.WithDeadline(ctx, t)
			}
		}
		resp, err := s.Invoke(ctx, req)
		cancel()
		if err != nil {
			resp.Error = []byte(err.Error())
		}
		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()
		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	service, ok := s.services[req.ServiceName]
	res := &message.Response{
		RequestID:  req.RequestID,
		Version:    req.Version,
		Serializer: req.Serializer,
		Compressor: req.Compressor,
	}
	if !ok {
		return res, errors.New("你要调用的服务不存在")
	}
	if isOneWay(ctx) {
		go func() {
			service.invoke(ctx, req)
		}()
		return &message.Response{}, errors.New("micro：微服务服务端 oneway 请求")
	}
	resp, err := service.invoke(ctx, req)
	res.Data = resp
	if err != nil {
		return res, err
	}
	return res, nil
}

type reflectionStub struct {
	value     reflect.Value
	serialize map[uint8]serialize.Serialize
	compress  map[uint8]compress.Compress
}

func (r *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(ctx)
	method := r.value.MethodByName(req.MethodName)
	inReq := reflect.New(method.Type().In(1).Elem())
	compress, ok := r.compress[req.Compressor]
	if !ok {
		return nil, errors.New("micro：不支持的压缩算法")
	}
	serializer, ok := r.serialize[req.Serializer]
	if !ok {
		return nil, errors.New("micro：不支持的序列化协议")
	}
	data, err := compress.UnCompress(req.Data)
	if err != nil {
		return nil, err
	}
	err = serializer.Decode(data, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq
	results := method.Call(in)
	if results[1].Interface() != nil {
		err = results[1].Interface().(error)
	}
	var res []byte
	var er error
	if results[0].IsNil() {
		return nil, err
	} else {
		res, er = serializer.Encode(results[0].Interface())
		if er != nil {
			return nil, er
		}
		res, er = compress.Compress(res)
		if er != nil {
			return nil, er
		}
		return res, err
	}
}
