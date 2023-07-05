package limit

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"
)

func NewErrKeyNotFound(key string) error {
	return fmt.Errorf("key %s 不存在", key)
}

type MaxMemoryCache struct {
	CacheV1
	max        int64
	used       int64
	head, tail *DLinkedNode
	cache      map[string]*DLinkedNode
	mutex      *sync.RWMutex
}

type DLinkedNode struct {
	key  string
	prev *DLinkedNode
	next *DLinkedNode
}

func NewMaxMemoryCache(max int64, cache CacheV1) *MaxMemoryCache {
	head := &DLinkedNode{}
	tail := &DLinkedNode{}
	head.next = tail
	tail.prev = head
	maxCache := &MaxMemoryCache{cache, max, 0, head, tail, map[string]*DLinkedNode{}, &sync.RWMutex{}}
	maxCache.CacheV1.OnEvicted(maxCache.delete)
	return maxCache
}

func (m *MaxMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if _, ok := m.cache[key]; !ok {
		return nil, NewErrKeyNotFound(key)
	}
	val, err := m.CacheV1.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	m.moveToHead(m.cache[key])
	return val, nil
}
func (m *MaxMemoryCache) addToHead(node *DLinkedNode) {
	node.prev = m.head
	node.next = m.head.next
	m.head.next.prev = node
	m.head.next = node
}

func (m *MaxMemoryCache) removeNode(node *DLinkedNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (m *MaxMemoryCache) moveToHead(node *DLinkedNode) {
	m.removeNode(node)
	m.addToHead(node)
}

func (m *MaxMemoryCache) removeTail() *DLinkedNode {
	node := m.tail.prev
	m.removeNode(node)
	return node
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.cache[key]; !ok {
		return m.addNewNode(ctx, key, val)
	} else {
		node := m.cache[key]
		oldVal, err := m.CacheV1.LoadAndDelete(ctx, key)
		if err != nil {
			return err
		}
		if bytes.Compare(oldVal, val) == 0 {
			m.moveToHead(node)
		} else {
			if err := m.CacheV1.Delete(ctx, key); err != nil {
				return err
			}
			return m.addNewNode(ctx, key, val)
		}
	}
	return nil
}

func (m *MaxMemoryCache) GetAllKeys() []string {
	var result []string
	pre := m.head.next
	for pre != m.tail {
		result = append(result, pre.key)
		pre = pre.next
	}
	return result
}

func (m *MaxMemoryCache) addNewNode(ctx context.Context, key string, val []byte) error {
	node := &DLinkedNode{key: key}
	m.cache[key] = node
	m.addToHead(node)
	m.used += int64(len(val))
	for m.used > m.max {
		removed := m.removeTail()
		if err := m.CacheV1.Delete(ctx, removed.key); err != nil {
			return err
		}
	}
	return nil
}

func (m *MaxMemoryCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.CacheV1.Delete(ctx, key)
}

func (m *MaxMemoryCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.CacheV1.LoadAndDelete(ctx, key)
}

func (m *MaxMemoryCache) delete(key string, val []byte) {
	removeNode, ok := m.cache[key]
	if ok {
		m.used = m.used - int64(len(val))
		m.removeNode(removeNode)
		delete(m.cache, removeNode.key)
	}
}
