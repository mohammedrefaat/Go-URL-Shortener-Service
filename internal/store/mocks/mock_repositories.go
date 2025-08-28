package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) CreateURL(ctx context.Context, url *domain.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockURLRepository) GetURLByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockURLRepository) GetURLByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	args := m.Called(ctx, originalURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockURLRepository) UpdateClickCount(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}

func (m *MockURLRepository) GetAnalytics(ctx context.Context, shortCode string, days int) (*domain.AnalyticsResponse, error) {
	args := m.Called(ctx, shortCode, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AnalyticsResponse), args.Error(1)
}

func (m *MockURLRepository) DeleteExpiredURLs(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockURLRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *MockURLRepository) IsShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) Increment(ctx context.Context, key string, value int64) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCacheRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheRepository) Cleanup(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheRepository) GetCounter(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}
