package cache

import (
	"container/list"
	"sync"
)

type EvictionNotifier interface {
	OnEviction(key interface{})
}

type LRU struct {
	maxEntries int

	mu    sync.Mutex
	ll    *list.List
	cache map[interface{}]*list.Element
}

type entry struct {
	key, value interface{}
}

const runEvictionNotifier = true

func NewLRU(maxEntries int) *LRU {
	return &LRU{
		maxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

func (c *LRU) Add(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}

	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele

	if c.ll.Len() > c.maxEntries {
		c.removeOldest(runEvictionNotifier)
	}
}

func (c *LRU) Get(key interface{}) (value interface{}, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

func (c *LRU) RemoveOldest() (key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.removeOldest(!runEvictionNotifier)
}

func (c *LRU) Remove(key interface{}) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, found := c.cache[key]; found {
		_, value := c.remove(ele)
		return value
	}
	return nil
}

func (c *LRU) removeOldest(runEvictionNotifier bool) (key, value interface{}) {
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	if runEvictionNotifier {
		ent := ele.Value.(*entry)
		if notifier, ok := ent.value.(EvictionNotifier); ok {
			notifier.OnEviction(ent.key)
		}
	}
	return c.remove(ele)
}

func (c *LRU) remove(ele *list.Element) (key, value interface{}) {
	c.ll.Remove(ele)
	ent := ele.Value.(*entry)
	delete(c.cache, ent.key)
	return ent.key, ent.value
}

func (c *LRU) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

func (c *LRU) PeekOldest() (key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	ent := ele.Value.(*entry)
	return ent.key, ent.value
}

func (c *LRU) PeekNewest() (key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ele := c.ll.Front()
	if ele == nil {
		return
	}
	ent := ele.Value.(*entry)
	return ent.key, ent.value
}

type Iterator struct {
	lru     *LRU
	this    *list.Element
	forward bool
}

func (c *LRU) NewIterator() *Iterator {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &Iterator{lru: c, this: c.ll.Front(), forward: true}
}

func (c *LRU) NewReverseIterator() *Iterator {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &Iterator{lru: c, this: c.ll.Back(), forward: false}
}

func (i *Iterator) GetAndAdvance() (interface{}, interface{}, bool) {
	i.lru.mu.Lock()
	defer i.lru.mu.Unlock()
	if i.this == nil {
		return nil, nil, false
	}
	ent := i.this.Value.(*entry)
	if i.forward {
		i.this = i.this.Next()
	} else {
		i.this = i.this.Prev()
	}
	return ent.key, ent.value, true
}
