package pokecache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	createdAt 	time.Time
	val 		[]byte
}

type Cache struct {
	mu 			sync.Mutex
	pokeCache 	map[string]CacheEntry
	interval 	time.Duration
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		pokeCache: make(map[string]CacheEntry),
		interval: interval,
	}
	go c.reapLoop()
	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pokeCache[key] = CacheEntry{
		createdAt: time.Now(),
		val: val,
	}	
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.pokeCache[key]
	return entry.val, ok
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)

	for range ticker.C {
		c.mu.Lock()
		
		for key, entry := range c.pokeCache {
			if time.Since(entry.createdAt) > c.interval {
				delete(c.pokeCache, key)
			}
		}
		c.mu.Unlock()
	}
}