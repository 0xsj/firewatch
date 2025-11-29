package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/redis/go-redis/v9"
)

// redisCache is a Redis implementation of cache.Cache.
type redisCache struct {
	client *redis.Client
}

// New creates a new Redis cache.
func New(options *redis.Options) (cache.Cache, error) {
	client := redis.NewClient(options)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisCache{
		client: client,
	}, nil
}

// NewFromClient creates a cache from an existing Redis client.
func NewFromClient(client *redis.Client) cache.Cache {
	return &redisCache{
		client: client,
	}
}

// Get retrieves a value by key.
func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, cache.ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

// Set stores a value with TTL.
func (c *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// Delete removes a value by key.
func (c *redisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

// Exists checks if a key exists.
func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return n > 0, nil
}

// Increment atomically increments a key by delta and returns the new value.
func (c *redisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	val, err := c.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis increment failed: %w", err)
	}
	return val, nil
}

// Expire sets a TTL on an existing key.
func (c *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	set, err := c.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	if !set {
		return cache.ErrKeyNotFound
	}
	return nil
}

// TTL returns the remaining time-to-live for a key.
func (c *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}

	// Redis returns -2 if the key does not exist
	if ttl == -2*time.Second {
		return 0, cache.ErrKeyNotFound
	}

	// Redis returns -1 if the key exists but has no expiration
	if ttl == -1*time.Second {
		return 0, nil
	}

	return ttl, nil
}

// Close closes the Redis connection.
func (c *redisCache) Close() error {
	return c.client.Close()
}
