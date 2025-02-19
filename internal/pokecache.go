package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	data  map[string]cacheEntry
	mutex sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	c := Cache{data: map[string]cacheEntry{}, mutex: sync.Mutex{}}
	go c.reapLoop(interval)
	return &c
}

func (c *Cache) Add(key string, val []byte) {
	c.mutex.Lock()
	c.data[key] = cacheEntry{createdAt: time.Now(), val: val}
	c.mutex.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	entry, ok := c.data[key]
	val := entry.val
	c.mutex.Unlock()
	return val, ok
}

func (c *Cache) reapLoop(interval time.Duration) {
	for {
		time.Sleep(interval)
		c.mutex.Lock()
		for entry := range c.data {
			if time.Since(c.data[entry].createdAt) > interval {
				delete(c.data, entry)
			}
		}
		c.mutex.Unlock()
	}
}
