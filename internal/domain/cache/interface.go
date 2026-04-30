package cache

import (
	"context"
	"time"
)

type ICache interface {
	IsHealthy(ctx context.Context) bool
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
