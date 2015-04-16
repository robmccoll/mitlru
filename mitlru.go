package mitlru

import (
	"container/list"
	"sync"
)

// LRUCache stores up to capacity pairs of keys and values,
// then begins removing the least recently read/created
// pairs for every new pair beyond insert.
type LRUCache struct {
	capacity int
	mapping  map[interface{}]*list.Element
	order    *list.List
	lock     sync.RWMutex
}

type pair struct {
	key interface{}
	val interface{}
}

// NewLRUCache returns a cache that starts removing elements after capacity
// elements have been added
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		mapping:  make(map[interface{}]*list.Element),
		order:    list.New(),
	}
}

// Purge clears everything from the cache
func (lru *LRUCache) Purge() {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	lru.mapping = make(map[interface{}]*list.Element)
	lru.order = list.New()
}

// Add stores a key value par in the cache. If the key was already present,
// it is moved to the front
func (lru *LRUCache) Add(key, val interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if element, ok := lru.mapping[key]; ok {
		element.Value.(*pair).val = val
		lru.order.MoveToFront(element)
		return
	}

	lru.mapping[key] = lru.order.PushFront(&pair{key, val})

	for lru.order.Len() > lru.capacity {
		element := lru.order.Back()
		lru.order.Remove(element)
		delete(lru.mapping, element.Value.(*pair).key)
	}
}

// Get returns the value if the keyh exists in the cache
func (lru *LRUCache) Get(key interface{}) (value interface{}, foundInCache bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if element, ok := lru.mapping[key]; ok {
		lru.order.MoveToFront(element)
		return element.Value.(*pair).val, true
	}

	return nil, false
}

// Remove removes a key / value combination
func (lru *LRUCache) Remove(key interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if element, ok := lru.mapping[key]; ok {
		lru.order.Remove(element)
		delete(lru.mapping, element.Value.(*pair).key)
	}
}

// Len returns the number of pairs in the cache
func (lru *LRUCache) Len() int {
	lru.lock.RLock()
	defer lru.lock.RUnlock()
	return lru.order.Len()
}

// Capacity returns the capacity of the cache
func (lru *LRUCache) Capacity() int {
	lru.lock.RLock()
	defer lru.lock.RUnlock()
	return lru.capacity
}
