package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/service"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGetAnalytics_FromCache(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	urlRepo := new(mocks.MockURLRepository)
	cacheRepo := new(mocks.MockCacheRepository)

	expected := &domain.AnalyticsResponse{ShortCode: "abc123", ClickCount: 42}

	// Cache hit
	cacheRepo.On("Get", ctx, "analytics:abc123:7", &domain.AnalyticsResponse{}).
		Run(func(args mock.Arguments) {
			// inject expected into dest
			dest := args.Get(2).(*domain.AnalyticsResponse)
			*dest = *expected
		}).
		Return(nil)

	svc := service.NewAnalyticsService(urlRepo, cacheRepo, logger)

	resp, err := svc.GetAnalytics(ctx, "abc123", 7)
	require.NoError(t, err)
	require.Equal(t, expected, resp)

	cacheRepo.AssertExpectations(t)
	urlRepo.AssertNotCalled(t, "GetAnalytics", ctx, "abc123", 7)
}

func TestGetAnalytics_FromDB_AndCacheSet(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	urlRepo := new(mocks.MockURLRepository)
	cacheRepo := new(mocks.MockCacheRepository)

	expected := &domain.AnalyticsResponse{ShortCode: "abc123", ClickCount: 99}

	// Cache miss
	cacheRepo.On("Get", ctx, "analytics:abc123:7", &domain.AnalyticsResponse{}).
		Return(errors.New("cache miss"))

	// DB hit
	urlRepo.On("GetAnalytics", ctx, "abc123", 7).
		Return(expected, nil)

	// Set cache
	cacheRepo.On("Set", ctx, "analytics:abc123:7", expected, 15*time.Minute).
		Return(nil)

	svc := service.NewAnalyticsService(urlRepo, cacheRepo, logger)

	resp, err := svc.GetAnalytics(ctx, "abc123", 7)
	require.NoError(t, err)
	require.Equal(t, expected, resp)

	cacheRepo.AssertExpectations(t)
	urlRepo.AssertExpectations(t)
}

func TestGetAnalytics_DBError(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	urlRepo := new(mocks.MockURLRepository)
	cacheRepo := new(mocks.MockCacheRepository)

	// Cache miss
	cacheRepo.On("Get", ctx, "analytics:abc123:7", &domain.AnalyticsResponse{}).
		Return(errors.New("cache miss"))

	// DB error
	urlRepo.On("GetAnalytics", ctx, "abc123", 7).
		Return(nil, errors.New("db error"))

	svc := service.NewAnalyticsService(urlRepo, cacheRepo, logger)

	resp, err := svc.GetAnalytics(ctx, "abc123", 7)
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get analytics")

	cacheRepo.AssertExpectations(t)
	urlRepo.AssertExpectations(t)
}
