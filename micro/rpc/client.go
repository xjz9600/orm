package rpc

import (
	"context"
	"errors"
	"github.com/silenceper/pool"
	"net"
	"orm/micro/rpc/compress"
	"orm/micro/rpc/message"
	"orm/micro/rpc/serialize"
	"orm/micro/rpc/serialize/json"
	"reflect"
	"strconv"
	"time"
)

func (c *Client) InitService(service Service) error {
	return setFuncField(service, c, c.serializer, c.compress)
}

type ClientOpt func(client *Client)

func setFuncField(service Service, p Proxy, s serialize.Serialize, c compress.Compress) error {
	if service == nil {
		return errors.New("rpc：不支持 nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc：只支持指向结构体的一级指针")
	}
	val = val.Elem()
	typ = typ.Elem()
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)
		if fieldVal.CanSet() {
			fn := func(args []reflect.Value) (results []reflect.Value) {
				ctx := args[0].Interface().(context.Context)
				reqData, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{reflect.Zero(fieldTyp.Type.Out(0)), reflect.ValueOf(err)}
				}
				compressData, err := c.Compress(reqData)
				if err != nil {
					return []reflect.Value{reflect.Zero(fieldTyp.Type.Out(0)), reflect.ValueOf(err)}
				}
				meta := make(map[string]string, 2)
				if isOneWay(ctx) {
					meta["one-way"] = "true"
				}
				if deadline, ok := ctx.Deadline(); ok {
					meta["deadline"] = strconv.FormatInt(deadline.Unix(), 10)
				}
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        compressData,
					Serializer:  s.Code(),
					Meta:        meta,
					Compressor:  c.Code(),
				}
				req.CalculateHeaderLength()
				req.CalculateBodyLength()
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{reflect.Zero(fieldTyp.Type.Out(0)), reflect.ValueOf(err)}
				}
				if len(resp.Error) > 0 {
					err = errors.New(string(resp.Error))
				}
				if len(resp.Data) > 0 {
					data, er := c.UnCompress(resp.Data)
					if er != nil {
						return []reflect.Value{retVal, reflect.ValueOf(er)}
					}
					er = s.Decode(data, retVal.Interface())
					if er != nil {
						return []reflect.Value{retVal, reflect.ValueOf(er)}
					}
				}
				if err == nil {
					return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
				}
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}
	return nil
}

type Client struct {
	pool       pool.Pool
	serializer serialize.Serialize
	compress   compress.Compress
}

func NewClient(addr string, opts ...ClientOpt) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap:  1,
		MaxCap:      30,
		MaxIdle:     10,
		IdleTimeout: time.Minute,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, time.Second*3)
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	res := &Client{
		pool:       p,
		serializer: &json.Serializer{},
		compress:   &compress.NoCompress{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func ClientWithSerializer(sl serialize.Serialize) ClientOpt {
	return func(client *Client) {
		client.serializer = sl
	}
}

func ClientWithCompress(cp compress.Compress) ClientOpt {
	return func(client *Client) {
		client.compress = cp
	}
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	ch := make(chan struct{}, 1)
	var (
		resp *message.Response
		err  error
	)
	go func() {
		resp, err = c.doInvoke(ctx, req)
		ch <- struct{}{}
		close(ch)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)
	resp, err := c.send(ctx, data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		c.pool.Put(conn)
	}()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	if isOneWay(ctx) {
		return nil, errors.New("micro：这时一个 oneway 调用，你不需要处理任何结果")
	}
	return ReadMsg(conn)
}
