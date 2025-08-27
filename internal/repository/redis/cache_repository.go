package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(redisURL string) (*CacheRepository, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CacheRepository{client: client}, nil
}

func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *CacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal([]byte(data), dest)
}

func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *CacheRepository) Increment(ctx context.Context, key string, value int64) error {
	return r.client.IncrBy(ctx, key, value).Err()
}

func (r *CacheRepository) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *CacheRepository) Close() error {
	return r.client.Close()
}

func (r *CacheRepository) SetURLMapping(ctx context.Context, shortKey string, url string, ttl time.Duration) error {
	// Store shortKey → URL
	if err := r.Set(ctx, "shortKey:"+shortKey, url, ttl); err != nil {
		return fmt.Errorf("failed to cache shortKey mapping: %w", err)
	}
	// Store url → shortKey
	if err := r.Set(ctx, "url:"+url, shortKey, ttl); err != nil {
		return fmt.Errorf("failed to cache url mapping: %w", err)
	}
	return nil
}

func (r *CacheRepository) GetURLByShortKey(ctx context.Context, shortKey string) (string, error) {
	var url string
	if err := r.Get(ctx, "shortKey:"+shortKey, &url); err != nil {
		return "", err
	}
	return url, nil
}

func (r *CacheRepository) GetShortKeyByURL(ctx context.Context, url string) (string, error) {
	var shortKey string
	if err := r.Get(ctx, "url:"+url, &shortKey); err != nil {
		return "", err
	}
	return shortKey, nil
}
