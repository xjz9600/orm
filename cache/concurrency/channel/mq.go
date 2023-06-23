package channel

import (
	"errors"
	"sync"
)

type Broker struct {
	mutex sync.RWMutex
	data  []chan Msg
}

type Msg struct {
	Content string
}

func (b *Broker) Send(m Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, d := range b.data {
		select {
		case d <- m:
		default:
			return errors.New("消息队列已满")
		}
	}
	return nil
}

func (b *Broker) Subscribe(capacity int) (<-chan Msg, error) {
	defer b.mutex.Unlock()
	res := make(chan Msg, capacity)
	b.mutex.Lock()
	b.data = append(b.data, res)
	return res, nil
}

func (b *Broker) Close() error {
	b.mutex.Lock()
	data := b.data
	b.data = nil
	for _, d := range data {
		close(d)
	}
	return nil
}
