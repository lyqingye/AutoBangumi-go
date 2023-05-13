package utils

import (
	"sync"
	"time"
)

type SimpleTTLCache struct {
	inner map[interface{}]WrapCacheValue
	lock  sync.RWMutex
}

type WrapCacheValue struct {
	Value      interface{}
	ExpireTime time.Time
}

func NewSimpleTTLCache(period time.Duration) *SimpleTTLCache {
	cache := SimpleTTLCache{
		inner: make(map[interface{}]WrapCacheValue),
		lock:  sync.RWMutex{},
	}
	go func() {
		for {
			now := time.Now()
			func() {
				cache.lock.Lock()
				defer cache.lock.Unlock()
				for k, v := range cache.inner {
					if now.After(v.ExpireTime) {
						delete(cache.inner, k)
					}
				}
			}()
			time.Sleep(period)
		}
	}()
	return &cache
}

func (c *SimpleTTLCache) Put(key interface{}, value interface{}, ttl time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.inner[key] = WrapCacheValue{
		Value:      value,
		ExpireTime: time.Now().Add(ttl),
	}
}

func (c *SimpleTTLCache) Get(key interface{}) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RLocker().Unlock()

	wrapped, found := c.inner[key]
	if !found {
		return nil, found
	}
	return wrapped.Value, true
}

func (c *SimpleTTLCache) Delete(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.inner, key)
}
