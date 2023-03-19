package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
	"github.com/palantir/stacktrace"
)

// RedisCache is the Cache implementation in redis
type RedisCache struct {
	tracer telemetry.Tracer
	client *redis.Client
}

// NewRedisCache creates a new instance of RedisCache
func NewRedisCache(tracer telemetry.Tracer, client *redis.Client) Cache {
	return &RedisCache{
		tracer: tracer,
		client: client,
	}
}

// Get an item from the redis cache
func (cache *RedisCache) Get(ctx context.Context, key string) (value string, err error) {
	ctx, span := cache.tracer.Start(ctx)
	defer span.End()

	response, err := cache.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("no item found in redis with key [%s]", key))
	}
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("cannot get item in redis with key [%s]", key))
	}
	return response, nil
}

// Set an item in the redis cache
func (cache *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	ctx, span := cache.tracer.Start(ctx)
	defer span.End()

	err := cache.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return cache.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, "cannot set item in redis"))
	}
	return nil
}
