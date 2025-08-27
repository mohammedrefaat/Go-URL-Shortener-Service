package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
)

type AnalyticsService struct {
	urlRepo   domain.URLRepository
	cacheRepo domain.CacheRepository
	logger    *zap.Logger
}

func NewAnalyticsService(urlRepo domain.URLRepository, cacheRepo domain.CacheRepository, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
		logger:    logger,
	}
}

func (s *AnalyticsService) GetAnalytics(ctx context.Context, shortCode string, days int) (*domain.AnalyticsResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("analytics:%s:%d", shortCode, days)
	var cachedAnalytics domain.AnalyticsResponse
	if err := s.cacheRepo.Get(ctx, cacheKey, &cachedAnalytics); err == nil {
		return &cachedAnalytics, nil
	}

	// Fallback to database
	analytics, err := s.urlRepo.GetAnalytics(ctx, shortCode, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Cache for future requests
	if err := s.cacheRepo.Set(ctx, cacheKey, analytics, 15*time.Minute); err != nil {
		s.logger.Warn("Failed to cache analytics", zap.Error(err))
	}

	return analytics, nil
}
