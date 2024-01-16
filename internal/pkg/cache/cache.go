package cache

type Cache interface {
	Add(key string, expiration int64) error // Add - add rand value with expiration (in seconds) to cache
	Contains(key string) (bool, error)      // Contains - check existence of the key in cache
	Delete(key string) error                // Delete - delete key from cache
}

type cacheValue struct {
	SetTime    int64
	Expiration int64
}
