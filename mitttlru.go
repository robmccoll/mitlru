package mitlru

import (
	"container/list"
	"sync"
	"time"
)

// TTLRUCache stores up to capacity triples of keys, values, and timestamps,
// then begins removing the least recently read/created
// triples for every new triple beyond insert.
type TTLRUCache struct {
	capacity  int
	mapping   map[interface{}]*triple
	order     *list.List
	timeorder *list.List
	ttl       time.Duration
	lock      sync.RWMutex
}

type triple struct {
	key          interface{}
	val          interface{}
	expiration   time.Time
	timeelement  *list.Element
	orderelement *list.Element
}

// NewTTLRUCache returns a cache that starts removing elements after capacity
// elements have been added
func NewTTLRUCache(capacity int, ttl time.Duration) *TTLRUCache {
	rtn := &TTLRUCache{
		capacity:  capacity,
		mapping:   make(map[interface{}]*triple),
		order:     list.New(),
		timeorder: list.New(),
		ttl:       ttl,
	}
	go rtn.timer()
	return rtn
}

func (lru *TTLRUCache) timer() {
	for {
		lru.lock.Lock()
		now := time.Now()
		for lru.order.Len() > 0 {
			efront := lru.timeorder.Front()
			tfront := efront.Value.(*triple)
			if tfront.expiration.After(now) {
				break
			}
			lru.order.Remove(tfront.orderelement)
			lru.timeorder.Remove(efront)
			delete(lru.mapping, tfront.key)
		}
		lru.lock.Unlock()
		time.Sleep(1 * time.Second)
	}
}

// Purge clears everything from the cache
func (lru *TTLRUCache) Purge() {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	lru.mapping = make(map[interface{}]*triple)
	lru.order = list.New()
	lru.timeorder = list.New()
}

// Add stores a key value par in the cache. If the key was already present,
// it is moved to the front
func (lru *TTLRUCache) Add(key, val interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if etriple, ok := lru.mapping[key]; ok {
		etriple.val = val
		lru.order.MoveToFront(etriple.orderelement)
		lru.timeorder.MoveToBack(etriple.timeelement)
		etriple.expiration = time.Now().Add(lru.ttl)
		return
	}

	etriple := &triple{key, val, time.Now().Add(lru.ttl), nil, nil}
	lru.mapping[key] = etriple
	etriple.orderelement = lru.order.PushFront(etriple)
	etriple.timeelement = lru.timeorder.PushBack(etriple)

	for lru.order.Len() > lru.capacity {
		element := lru.order.Back()
		etriple := element.Value.(*triple)
		lru.order.Remove(element)
		lru.timeorder.Remove(etriple.timeelement)
		delete(lru.mapping, etriple.key)
	}
}

// Get returns the value if the keyh exists in the cache
func (lru *TTLRUCache) Get(key interface{}) (value interface{}, foundInCache bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if element, ok := lru.mapping[key]; ok {
		lru.order.MoveToFront(element.orderelement)
		return element.val, true
	}

	return nil, false
}

// Remove removes a key / value combination
func (lru *TTLRUCache) Remove(key interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if element, ok := lru.mapping[key]; ok {
		lru.order.Remove(element.orderelement)
		lru.timeorder.Remove(element.timeelement)
		delete(lru.mapping, element.key)
	}
}

// Len returns the number of triples in the cache
func (lru *TTLRUCache) Len() int {
	lru.lock.RLock()
	defer lru.lock.RUnlock()
	return lru.order.Len()
}

// Capacity returns the capacity of the cache
func (lru *TTLRUCache) Capacity() int {
	return lru.capacity
}
