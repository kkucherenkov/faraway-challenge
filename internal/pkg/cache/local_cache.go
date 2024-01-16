package cache

import (
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/clock"
	"sync"
)

type LocalCache struct {
	dataMap map[string]cacheValue
	lock    *sync.Mutex
	clock   clock.Clock
}

func NewLocalCache(clock clock.Clock) Cache {
	return &LocalCache{
		dataMap: make(map[string]cacheValue),
		lock:    &sync.Mutex{},
		clock:   clock,
	}
}

func (c *LocalCache) Add(key string, expiration int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.dataMap[key] = cacheValue{
		SetTime:    c.clock.Now().Unix(),
		Expiration: expiration,
	}
	return nil
}

func (c *LocalCache) Contains(key string) (bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	value, ok := c.dataMap[key]
	if ok && c.clock.Now().Unix()-value.SetTime > value.Expiration {
		return false, nil
	}
	return ok, nil
}

func (c *LocalCache) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.dataMap, key)
	return nil
}
