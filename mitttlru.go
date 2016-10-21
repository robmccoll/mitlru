package mitlru

import (
	"sync"
	"time"

	"github.com/robmccoll/mitlru/list"
)

// TTLRUCache stores up to capacity triples of keys, values, and timestamps,
// then begins removing the least recently read/created
// triples for every new triple beyond insert.
type TTLRUCache struct {
	capacity   int
	mapping    map[interface{}]*triple
	order      *list.List
	timeorder  *list.List
	ttl        time.Duration
	lock       sync.RWMutex
	triplePool sync.Pool
}

type triple struct {
	key          interface{}
	val          interface{}
	expiration   time.Time
	timeelement  list.Element
	orderelement list.Element
}

func newTriple() interface{} {
	return &triple{}
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
	rtn.triplePool.New = newTriple
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
			lru.order.Remove(&tfront.orderelement)
			lru.timeorder.Remove(efront)
			delete(lru.mapping, tfront.key)
			lru.triplePool.Put(tfront)
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

// Add stores a key value pair in the cache. If the key was already present,
// it is moved to the front
func (lru *TTLRUCache) Add(key, val interface{}) {
	lru.AddWithExpire(key, val, time.Now().Add(lru.ttl))
}

// AddWithTTL stores a key value pair in the cache. If the key was already present,
// it is moved to the front.  Additionally, the given TTL is used (overriding the
// default ttl).
func (lru *TTLRUCache) AddWithTTL(key, val interface{}, ttl time.Duration) {
	lru.AddWithExpire(key, val, time.Now().Add(ttl))
}

// AddWithExpire stores a key value pair in the cache. If the key was already present,
// it is moved to the front.  Additionally, the given expiration time is used (overriding the
// default ttl).
func (lru *TTLRUCache) AddWithExpire(key, val interface{}, expiration time.Time) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if etriple, ok := lru.mapping[key]; ok {
		etriple.val = val
		lru.order.MoveToFront(&etriple.orderelement)
		lru.timeorder.MoveToBack(&etriple.timeelement)
		etriple.expiration = expiration
		return
	}

	etriple := lru.triplePool.Get().(*triple)
	etriple.key = key
	etriple.val = val
	etriple.expiration = expiration
	etriple.orderelement.Value = etriple
	etriple.timeelement.Value = etriple
	lru.order.MoveToFront(&etriple.orderelement)
	lru.timeorder.MoveToBack(&etriple.timeelement)
	lru.mapping[key] = etriple

	for lru.order.Len() > lru.capacity {
		element := lru.order.Back()
		etriple := element.Value.(*triple)
		lru.order.Remove(element)
		lru.timeorder.Remove(&etriple.timeelement)
		delete(lru.mapping, etriple.key)
		lru.triplePool.Put(etriple)
	}
}

// Get returns the value if the keyh exists in the cache
func (lru *TTLRUCache) Get(key interface{}) (value interface{}, foundInCache bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if element, ok := lru.mapping[key]; ok {
		lru.order.MoveToFront(&element.orderelement)
		return element.val, true
	}

	return nil, false
}

// Remove removes a key / value combination
func (lru *TTLRUCache) Remove(key interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if element, ok := lru.mapping[key]; ok {
		lru.order.Remove(&element.orderelement)
		lru.timeorder.Remove(&element.timeelement)
		delete(lru.mapping, element.key)
		lru.triplePool.Put(element)
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
