package llm

import (
	"container/list"
	"sync"
	"time"
)

type cacheEntry struct {
	key       string
	resp      *ChatResponse
	expiresAt time.Time
}

type Cache struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	lru      *list.List
}

func NewCache(capacity int, ttl time.Duration) *Cache {
	return &Cache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element),
		lru:      list.New(),
	}
}

func (c *Cache) Get(key string) (*ChatResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	entry := elem.Value.(*cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.lru.Remove(elem)
		delete(c.items, key)
		return nil, false
	}

	c.lru.MoveToFront(elem)
	return entry.resp, true
}

func (c *Cache) Set(key string, resp *ChatResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.lru.MoveToFront(elem)
		elem.Value.(*cacheEntry).resp = resp
		elem.Value.(*cacheEntry).expiresAt = time.Now().Add(c.ttl)
		return
	}

	entry := &cacheEntry{
		key:       key,
		resp:      resp,
		expiresAt: time.Now().Add(c.ttl),
	}

	if c.lru.Len() >= c.capacity {
		oldest := c.lru.Back()
		if oldest != nil {
			c.lru.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}

	elem := c.lru.PushFront(entry)
	c.items[key] = elem
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.lru.Init()
}

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Len()
}

func (c *Cache) Cap() int { return c.capacity }

func (c *Cache) TTL() time.Duration { return c.ttl }
