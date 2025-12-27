package cache

import (
	"sync"
	"time"
)

// Item represents a cached item with an expiration time.
type Item struct {
	Value      interface{}
	Expiration int64
}

// Cache is a thread-safe in-memory cache.
type Cache struct {
	items map[string]Item
	mu    sync.RWMutex
}

// New creates a new instance of Cache.
func New() *Cache {
	return &Cache{
		items: make(map[string]Item),
	}
}

// Set adds an item to the cache with a specified TTL.
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl).UnixNano()
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves an item from the cache.
// It returns the value and a boolean indicating if the item was found.
// Note: Expiration logic will be implemented in the next task.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}
