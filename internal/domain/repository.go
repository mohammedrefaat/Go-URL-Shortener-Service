package domain

import (
	"context"
	"time"
)

type URLRepository interface {
	CreateURL(ctx context.Context, url *URL) error                                            // Create a new URL
	GetURLByShortCode(ctx context.Context, shortCode string) (*URL, error)                    // Get a URL by its short code
	GetURLByOriginalURL(ctx context.Context, originalURL string) (*URL, error)                // Get a URL by its original URL
	UpdateClickCount(ctx context.Context, shortCode string) error                             // Update the click count for a URL
	GetAnalytics(ctx context.Context, shortCode string, days int) (*AnalyticsResponse, error) // Get analytics for a URL
	DeleteExpiredURLs(ctx context.Context) error                                              // Delete expired URLs
	HealthCheck(ctx context.Context) error                                                    // Check the health of the database
	IsShortCodeExists(ctx context.Context, shortCode string) (bool, error)                    // Check if a short code exists
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error // Set a value in the cache
	Get(ctx context.Context, key string, dest interface{}) error                     // Get a value from the cache
	Delete(ctx context.Context, key string) error                                    // Delete a value from the cache
	Increment(ctx context.Context, key string, value int64) error                    // Increment a value in the cache
	HealthCheck(ctx context.Context) error                                           // Check the health of the cache
	Cleanup(ctx context.Context) error                                               // Cleanup expired cache entries
	GetCounter(ctx context.Context, key string) (int64, error)
}

type AnalyticsRepository interface {
	RecordClick(ctx context.Context, analytics *URLAnalytics) error
	GetClickCount(ctx context.Context, shortCode string) (int64, error)
	GetDailyStats(ctx context.Context, shortCode string, days int) ([]DailyStat, error)
	GetLastAccessed(ctx context.Context, shortCode string) (*time.Time, error)
}
