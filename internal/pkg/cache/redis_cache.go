package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/clock"
	"time"
)

type RedisCache struct {
	ctx    context.Context
	client *redis.Client
	clock  clock.Clock
}

func NewRedisCache(ctx context.Context, host string, port int) (Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// check connection by setting test value
	err := rdb.Set(ctx, "key", "value", 0).Err()

	return &RedisCache{
		ctx:    ctx,
		client: rdb,
	}, err
}

func (c *RedisCache) Add(key string, expiration int64) error {
	return c.client.Set(c.ctx, key, "value", time.Duration(expiration*1e9)*time.Second).Err()
}

func (c *RedisCache) Contains(key string) (bool, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	return val != "", err
}

func (c *RedisCache) Delete(key string) error {
	err := c.client.Del(c.ctx, key).Err()
	return err
}
