package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/utils"
)

var (
	ErrURLNotFound      = errors.New("URL not found")
	ErrURLExpired       = errors.New("URL has expired")
	ErrInvalidURL       = errors.New("invalid URL")
	ErrCustomAliasTaken = errors.New("custom alias already taken")
)

type URLService struct {
	urlRepo   domain.URLRepository
	cacheRepo domain.CacheRepository
	logger    *zap.Logger
	baseURL   string
}

func NewURLService(urlRepo domain.URLRepository, cacheRepo domain.CacheRepository, logger *zap.Logger) *URLService {
	return &URLService{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
		logger:    logger,
		baseURL:   "http://localhost:8080", // Should come from config
	}
}

func (s *URLService) ShortenURL(ctx context.Context, req *domain.ShortenRequest) (*domain.ShortenResponse, error) {
	// Validate URL
	if !utils.IsValidURL(req.URL) {
		return nil, ErrInvalidURL
	}

	// Check if URL already exists
	if existing, err := s.urlRepo.GetURLByOriginalURL(ctx, req.URL); err == nil {
		return s.buildResponse(existing), nil
	}

	// Generate short code
	var shortCode string
	var err error

	if req.CustomAlias != "" {
		// Check if custom alias is available
		if _, err := s.urlRepo.GetURLByShortCode(ctx, req.CustomAlias); err == nil {
			return nil, ErrCustomAliasTaken
		}
		shortCode = req.CustomAlias
	} else {
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	// Create URL record
	url := &domain.URL{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		CreatedAt:   time.Now(),
		ExpiresAt:   req.ExpiresAt,
	}

	if err := s.urlRepo.CreateURL(ctx, url); err != nil {
		s.logger.Error("Failed to create URL", zap.Error(err))
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Cache the URL
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	if err := s.cacheRepo.Set(ctx, cacheKey, url, time.Hour); err != nil {
		s.logger.Warn("Failed to cache URL", zap.Error(err))
	}

	s.logger.Info("URL shortened successfully",
		zap.String("short_code", shortCode),
		zap.String("original_url", req.URL),
	)

	return s.buildResponse(url), nil
}

func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	var cachedURL domain.URL
	if err := s.cacheRepo.Get(ctx, cacheKey, &cachedURL); err == nil {
		if s.isExpired(&cachedURL) {
			return "", ErrURLExpired
		}

		// Increment click count asynchronously
		go s.incrementClickCount(context.Background(), shortCode)

		return cachedURL.OriginalURL, nil
	}

	// Fallback to database
	url, err := s.urlRepo.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", ErrURLNotFound
	}

	if s.isExpired(url) {
		return "", ErrURLExpired
	}

	// Cache for future requests
	if err := s.cacheRepo.Set(ctx, cacheKey, url, time.Hour); err != nil {
		s.logger.Warn("Failed to cache URL", zap.Error(err))
	}

	// Increment click count asynchronously
	go s.incrementClickCount(context.Background(), shortCode)

	return url.OriginalURL, nil
}

func (s *URLService) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		shortCode := utils.GenerateShortCode(6)

		// Check if short code already exists
		if _, err := s.urlRepo.GetURLByShortCode(ctx, shortCode); err != nil {
			// Short code doesn't exist, we can use it
			return shortCode, nil
		}
	}
	return "", errors.New("failed to generate unique short code")
}

func (s *URLService) incrementClickCount(ctx context.Context, shortCode string) {
	// Try to increment in cache first
	cacheKey := fmt.Sprintf("clicks:%s", shortCode)
	if err := s.cacheRepo.Increment(ctx, cacheKey, 1); err == nil {
		return
	}

	// Fallback to database
	if err := s.urlRepo.UpdateClickCount(ctx, shortCode); err != nil {
		s.logger.Error("Failed to increment click count",
			zap.String("short_code", shortCode),
			zap.Error(err),
		)
	}
}

func (s *URLService) isExpired(url *domain.URL) bool {
	if url.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*url.ExpiresAt)
}

func (s *URLService) buildResponse(url *domain.URL) *domain.ShortenResponse {
	return &domain.ShortenResponse{
		ShortURL:    s.baseURL + "/" + url.ShortCode,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		ExpiresAt:   url.ExpiresAt,
		CreatedAt:   url.CreatedAt,
	}
}
