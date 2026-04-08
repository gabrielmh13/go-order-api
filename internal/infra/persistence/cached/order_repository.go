package cached

import (
	"context"
	"encoding/json"
	"time"

	domaincache "go-order-api/internal/domain/cache"
	"go-order-api/internal/domain/logger"
	"go-order-api/internal/domain/order"
)

type OrderRepository struct {
	repo   order.IOrderRepository
	cache  domaincache.ICache
	ttl    time.Duration
	logger logger.Logger
}

func NewOrderRepository(repo order.IOrderRepository, cache domaincache.ICache, ttl time.Duration, l logger.Logger) order.IOrderRepository {
	return &OrderRepository{
		repo:   repo,
		cache:  cache,
		ttl:    ttl,
		logger: l,
	}
}

func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
	if err := r.repo.Save(ctx, o); err != nil {
		return err
	}

	r.setCache(ctx, o)
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	key := r.cacheKey(id)

	if r.cache != nil {
		cachedValue, err := r.cache.Get(ctx, key)
		if err == nil && cachedValue != "" {
			var cachedOrder order.Order
			if unmarshalErr := json.Unmarshal([]byte(cachedValue), &cachedOrder); unmarshalErr == nil {
				return &cachedOrder, nil
			} else {
				r.logger.Error("Failed to unmarshal cached order", unmarshalErr, logger.Any("orderId", id))
			}
			if deleteErr := r.cache.Delete(ctx, key); deleteErr != nil {
				r.logger.Warning("Failed to delete invalid cached order", logger.Any("orderId", id))
			}
		} else if err != nil {
			r.logger.Warning("Falling back to repository after cache read failure", logger.Any("orderId", id))
		}
	}

	o, repoErr := r.repo.GetByID(ctx, id)
	if repoErr != nil {
		return nil, repoErr
	}

	r.setCache(ctx, o)
	return o, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status order.OrderStatus) error {
	if err := r.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	if r.cache == nil {
		return nil
	}

	if err := r.cache.Delete(ctx, r.cacheKey(id)); err != nil {
		r.logger.Warning("Failed to invalidate cached order after status update", logger.Any("orderId", id))
	}
	return nil
}

func (r *OrderRepository) cacheKey(id string) string {
	return "order:" + id
}

func (r *OrderRepository) setCache(ctx context.Context, o *order.Order) {
	if r.cache == nil {
		return
	}

	payload, err := json.Marshal(o)
	if err != nil {
		r.logger.Error("Failed to marshal order for cache", err, logger.Any("orderId", o.ID))
		return
	}

	if err := r.cache.Set(ctx, r.cacheKey(o.ID), string(payload), r.ttl); err != nil {
		r.logger.Warning("Failed to cache order", logger.Any("orderId", o.ID))
	}
}
