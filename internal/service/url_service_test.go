package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/store/mocks"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/config"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
)

func TestURLService_ShortenURL(t *testing.T) {
	t.Run("SuccessfulShortening", func(t *testing.T) {
		// Create fresh mocks for each test
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, &config.Config{
			MachineID: 1,
			BaseURL:   "http://localhost:8080",
		})

		req := &domain.ShortenRequest{
			URL: "https://example.com",
		}

		// Mock cache lookup (cache miss)
		mockCache.On("Get", mock.Anything, "lurl:https://example.com", mock.AnythingOfType("*domain.URL")).
			Return(errors.New("not found"))

		// Mock database lookup for existing URL (not found)
		mockRepo.On("GetURLByOriginalURL", mock.Anything, "https://example.com").
			Return(nil, errors.New("not found"))

		// Mock URL creation
		mockRepo.On("CreateURL", mock.Anything, mock.AnythingOfType("*domain.URL")).
			Return(nil)

		// Mock both cache Set calls (the service makes two Set calls)
		mockCache.On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
			return key != "lurl:https://example.com" // This matches the url:shortcode pattern
		}), mock.Anything, time.Hour).Return(nil)

		mockCache.On("Set", mock.Anything, "lurl:https://example.com", mock.Anything, time.Hour).
			Return(nil)

		response, err := urlService.ShortenURL(context.Background(), req)

		assert.NoError(t, err)
		assert.NotEmpty(t, response.ShortCode)
		assert.Equal(t, "https://example.com", response.OriginalURL)
		assert.Contains(t, response.ShortURL, response.ShortCode)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, &config.Config{
			MachineID: 1,
			BaseURL:   "http://localhost:8080",
		})

		req := &domain.ShortenRequest{
			URL: "invalid-url",
		}

		response, err := urlService.ShortenURL(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidURL, err)
		assert.Nil(t, response)
		// No expectations to assert since validation happens before any repo/cache calls
	})

	t.Run("CustomAliasTaken", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, &config.Config{
			MachineID: 1,
			BaseURL:   "http://localhost:8080",
		})

		req := &domain.ShortenRequest{
			URL:         "https://example.com",
			CustomAlias: "taken",
		}

		existingURL := &domain.URL{
			ShortCode:   "taken",
			OriginalURL: "https://other.com",
		}

		// Mock cache lookup (cache miss)
		mockCache.On("Get", mock.Anything, "lurl:https://example.com", mock.AnythingOfType("*domain.URL")).
			Return(errors.New("not found"))

		// Mock database lookup for existing URL (not found)
		mockRepo.On("GetURLByOriginalURL", mock.Anything, "https://example.com").
			Return(nil, errors.New("not found"))

		// Mock custom alias check - this should find the existing alias
		mockRepo.On("GetURLByShortCode", mock.Anything, "taken").
			Return(existingURL, nil)

		response, err := urlService.ShortenURL(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, ErrCustomAliasTaken, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("URLFoundInCache", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, &config.Config{
			MachineID: 1,
			BaseURL:   "http://localhost:8080",
		})

		req := &domain.ShortenRequest{
			URL: "https://cached-example.com",
		}

		cachedURL := &domain.URL{
			ShortCode:   "cached123",
			OriginalURL: "https://cached-example.com",
			CreatedAt:   time.Now(),
			ExpiresAt:   nil, // Not expired
		}

		// Mock cache hit
		mockCache.On("Get", mock.Anything, "lurl:https://cached-example.com", mock.AnythingOfType("*domain.URL")).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*domain.URL)
				*arg = *cachedURL
			}).
			Return(nil)

		response, err := urlService.ShortenURL(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, "cached123", response.ShortCode)
		assert.Equal(t, "https://cached-example.com", response.OriginalURL)
		assert.Contains(t, response.ShortURL, response.ShortCode)
		mockCache.AssertExpectations(t)
	})

	t.Run("URLFoundInDatabase", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, &config.Config{
			MachineID: 1,
			BaseURL:   "http://localhost:8080",
		})

		req := &domain.ShortenRequest{
			URL: "https://db-example.com",
		}

		existingURL := &domain.URL{
			ShortCode:   "db123",
			OriginalURL: "https://db-example.com",
			CreatedAt:   time.Now(),
			ExpiresAt:   nil,
		}

		// Cache miss
		mockCache.On("Get", mock.Anything, "lurl:https://db-example.com", mock.AnythingOfType("*domain.URL")).
			Return(errors.New("not found"))

		// Database hit
		mockRepo.On("GetURLByOriginalURL", mock.Anything, "https://db-example.com").
			Return(existingURL, nil)

		// Cache the found URL
		mockCache.On("Set", mock.Anything, "lurl:https://db-example.com", existingURL, time.Hour).
			Return(nil)

		response, err := urlService.ShortenURL(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, "db123", response.ShortCode)
		assert.Equal(t, "https://db-example.com", response.OriginalURL)
		assert.Contains(t, response.ShortURL, response.ShortCode)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestURLService_GetOriginalURL(t *testing.T) {
	t.Run("CacheHit", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, nil)

		cachedURL := &domain.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			CreatedAt:   time.Now(),
		}

		mockCache.On("Get", mock.Anything, "url:abc123", mock.AnythingOfType("*domain.URL")).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*domain.URL)
				*arg = *cachedURL
			}).
			Return(nil)

		// Mock the increment call that happens asynchronously
		mockCache.On("Increment", mock.Anything, "clicks:abc123", int64(1)).
			Return(nil)

		originalURL, err := urlService.GetOriginalURL(context.Background(), "abc123")

		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", originalURL)

		// Wait a bit for the goroutine to complete
		time.Sleep(100 * time.Millisecond)
		mockCache.AssertExpectations(t)
	})

	t.Run("CacheMissDatabaseHit", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, nil)

		dbURL := &domain.URL{
			ShortCode:   "def456",
			OriginalURL: "https://example.org",
			CreatedAt:   time.Now(),
		}

		mockCache.On("Get", mock.Anything, "url:def456", mock.AnythingOfType("*domain.URL")).
			Return(errors.New("not found"))
		mockRepo.On("GetURLByShortCode", mock.Anything, "def456").
			Return(dbURL, nil)
		mockCache.On("Set", mock.Anything, "url:def456", dbURL, time.Hour).
			Return(nil)

		// Mock the increment call - first try cache, then fallback to database
		mockCache.On("Increment", mock.Anything, "clicks:def456", int64(1)).
			Return(errors.New("cache error"))
		mockRepo.On("UpdateClickCount", mock.Anything, "def456").Return(nil)

		originalURL, err := urlService.GetOriginalURL(context.Background(), "def456")

		assert.NoError(t, err)
		assert.Equal(t, "https://example.org", originalURL)

		// Wait a bit for the goroutine to complete
		time.Sleep(100 * time.Millisecond)
		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("URLNotFound", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, nil)

		mockCache.On("Get", mock.Anything, "url:notfound", mock.AnythingOfType("*domain.URL")).
			Return(errors.New("not found"))
		mockRepo.On("GetURLByShortCode", mock.Anything, "notfound").
			Return(nil, errors.New("not found"))

		originalURL, err := urlService.GetOriginalURL(context.Background(), "notfound")

		assert.Error(t, err)
		assert.Equal(t, ErrURLNotFound, err)
		assert.Empty(t, originalURL)
		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExpiredURL", func(t *testing.T) {
		mockRepo := new(mocks.MockURLRepository)
		mockCache := new(mocks.MockCacheRepository)
		logger := zaptest.NewLogger(t)
		urlService := NewURLService(mockRepo, mockCache, logger, nil)

		expiredTime := time.Now().Add(-time.Hour)
		expiredURL := &domain.URL{
			ShortCode:   "expired",
			OriginalURL: "https://expired.com",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			ExpiresAt:   &expiredTime,
		}

		mockCache.On("Get", mock.Anything, "url:expired", mock.AnythingOfType("*domain.URL")).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*domain.URL)
				*arg = *expiredURL
			}).
			Return(nil)

		originalURL, err := urlService.GetOriginalURL(context.Background(), "expired")

		assert.Error(t, err)
		assert.Equal(t, ErrURLExpired, err)
		assert.Empty(t, originalURL)
		mockCache.AssertExpectations(t)
	})
}
