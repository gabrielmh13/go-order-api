package redis

import (
	"context"
	"time"

	domaincache "go-order-api/internal/domain/cache"
	"go-order-api/internal/domain/logger"

	goredis "github.com/redis/go-redis/v9"
)

type Cache struct {
	client *goredis.Client
	logger logger.Logger
}

const cacheOperationTimeout = 150 * time.Millisecond

func NewCache(client *goredis.Client, l logger.Logger) domaincache.ICache {
	return &Cache{
		client: client,
		logger: l,
	}
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	cacheCtx, cancel := context.WithTimeout(ctx, cacheOperationTimeout)
	defer cancel()

	value, err := c.client.Get(cacheCtx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return "", nil
		}
		c.logger.Error("Failed to get value from Redis", err, logger.Any("key", key))
		return "", err
	}
	return value, nil
}

func (c *Cache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	cacheCtx, cancel := context.WithTimeout(ctx, cacheOperationTimeout)
	defer cancel()

	if err := c.client.Set(cacheCtx, key, value, ttl).Err(); err != nil {
		c.logger.Error("Failed to set value in Redis", err, logger.Any("key", key), logger.Any("ttl", ttl.String()))
		return err
	}
	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	cacheCtx, cancel := context.WithTimeout(ctx, cacheOperationTimeout)
	defer cancel()

	if err := c.client.Del(cacheCtx, key).Err(); err != nil {
		c.logger.Error("Failed to delete value from Redis", err, logger.Any("key", key))
		return err
	}
	return nil
}
