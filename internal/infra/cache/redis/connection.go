package redis

import (
	"context"
	"time"

	"go-order-api/internal/domain/logger"

	goredis "github.com/redis/go-redis/v9"
)

const startupPingTimeout = 1 * time.Second

func NewConnection(ctx context.Context, addr, password string, db int, l logger.Logger) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  200 * time.Millisecond,
		ReadTimeout:  150 * time.Millisecond,
		WriteTimeout: 150 * time.Millisecond,
		PoolTimeout:  200 * time.Millisecond,
		MaxRetries:   0,
	})

	pingCtx, cancel := context.WithTimeout(ctx, startupPingTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		l.Error("Failed to ping Redis", err, logger.Any("addr", addr), logger.Any("db", db))
		return client, err
	}

	l.Info("Connected to Redis successfully", logger.Any("addr", addr), logger.Any("db", db))
	return client, nil
}
