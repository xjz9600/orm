package micro

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	idlesConn   chan *idleConn
	reqQueue    []*connReq
	maxCnt      int
	cnt         int
	maxIdleTime time.Duration
	factory     func() (net.Conn, error)
	lock        sync.Mutex
}

type idleConn struct {
	c              net.Conn
	lastActiveTime time.Time
}

type connReq struct {
	connChan chan net.Conn
}

func NewPool(initCnt, maxCnt, maxIdleConn int, maxIdleTime time.Duration, factory func() (net.Conn, error)) (*Pool, error) {
	if initCnt > maxIdleConn {
		return nil, errors.New("micro：初始化连接数量不能大于最大空闲连接数量")
	}
	idleCons := make(chan *idleConn, maxIdleConn)
	for i := 0; i < maxIdleConn; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		idleCons <- &idleConn{c: conn, lastActiveTime: time.Now()}
	}
	res := &Pool{
		maxCnt:      maxCnt,
		maxIdleTime: maxIdleTime,
		idlesConn:   idleCons,
		factory:     factory,
	}
	return res, nil
}

func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	for {
		select {
		case ic := <-p.idlesConn:
			if ic.lastActiveTime.Add(p.maxIdleTime).Before(time.Now()) {
				_ = ic.c.Close()
				continue
			}
			return ic.c, nil
		default:
			p.lock.Lock()
			if p.cnt >= p.maxCnt {
				req := &connReq{make(chan net.Conn, 1)}
				p.reqQueue = append(p.reqQueue, req)
				p.lock.Unlock()
				select {
				case c := <-req.connChan:
					return c, nil
				case <-ctx.Done():
					go func() {
						c := <-req.connChan
						_ = p.Put(context.Background(), c)
					}()
					return nil, ctx.Err()
				}
			}
			c, err := p.factory()
			if err != nil {
				return nil, err
			}
			p.cnt++
			p.lock.Unlock()
			return c, nil
		}
	}
}

func (p *Pool) Put(ctx context.Context, conn net.Conn) error {
	p.lock.Lock()
	if len(p.reqQueue) > 0 {
		req := p.reqQueue[0]
		p.reqQueue = p.reqQueue[1:]
		p.lock.Unlock()
		req.connChan <- conn
	}
	defer p.lock.Unlock()
	ic := &idleConn{
		c:              conn,
		lastActiveTime: time.Now(),
	}
	select {
	case p.idlesConn <- ic:
	default:
		_ = conn.Close()
		p.cnt--
	}
	return nil
}
