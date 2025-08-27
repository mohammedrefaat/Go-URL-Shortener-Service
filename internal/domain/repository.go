package domain

import (
	"context"
	"time"
)

type URLRepository interface {
	CreateURL(ctx context.Context, url *URL) error
	GetURLByShortCode(ctx context.Context, shortCode string) (*URL, error)
	GetURLByOriginalURL(ctx context.Context, originalURL string) (*URL, error)
	UpdateClickCount(ctx context.Context, shortCode string) error
	GetAnalytics(ctx context.Context, shortCode string, days int) (*AnalyticsResponse, error)
	DeleteExpiredURLs(ctx context.Context) error
	HealthCheck(ctx context.Context) error
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string, value int64) error
	HealthCheck(ctx context.Context) error
}
