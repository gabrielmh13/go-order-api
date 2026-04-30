package redis

import (
	"context"
	"sync/atomic"
	"time"

	domaincache "go-order-api/internal/domain/cache"
	"go-order-api/internal/domain/logger"

	goredis "github.com/redis/go-redis/v9"
)

type Cache struct {
	client  *goredis.Client
	logger  logger.Logger
	healthy atomic.Bool
}

const (
	cacheOperationTimeout = 150 * time.Millisecond
	healthCheckInterval   = 10 * time.Second
	healthCheckTimeout    = 2 * time.Second
)

func NewCache(ctx context.Context, client *goredis.Client, l logger.Logger) domaincache.ICache {
	c := &Cache{
		client: client,
		logger: l,
	}
	c.healthy.Store(true)
	go c.watchHealth(ctx)
	return c
}

func (c *Cache) watchHealth(ctx context.Context) {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
			err := c.client.Ping(pingCtx).Err()
			cancel()

			wasHealthy := c.healthy.Load()
			if err != nil {
				c.healthy.Store(false)
				if wasHealthy {
					c.logger.Warning("Redis became unavailable", logger.Any("error", err))
				}
			} else {
				c.healthy.Store(true)
				if !wasHealthy {
					c.logger.Info("Redis recovered")
				}
			}
		}
	}
}

func (c *Cache) IsHealthy(_ context.Context) bool {
	return c.healthy.Load()
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
