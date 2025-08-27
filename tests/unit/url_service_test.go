package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/tests/mocks"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/service"
)

func TestURLService_ShortenURL(t *testing.T) {
	mockRepo := new(mocks.MockURLRepository)
	mockCache := new(mocks.MockCacheRepository)
	logger := zaptest.NewLogger(t)

	urlService := service.NewURLService(mockRepo, mockCache, logger)

	t.Run("SuccessfulShortening", func(t *testing.T) {
		req := &domain.ShortenRequest{
			URL: "https://example.com",
		}

		mockRepo.On("GetURLByOriginalURL", mock.Anything, "https://example.com").
			Return(nil, errors.New("not found"))
		mockRepo.On("GetURLByShortCode", mock.Anything, mock.AnythingOfType("string")).
			Return(nil, errors.New("not found"))
		mockRepo.On("CreateURL", mock.Anything, mock.AnythingOfType("*domain.URL")).
			Return(nil)
		mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, time.Hour).
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
		req := &domain.ShortenRequest{
			URL: "invalid-url",
		}

		response, err := urlService.ShortenURL(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidURL, err)
		assert.Nil(t, response)
	})

	t.Run("CustomAliasTaken", func(t *testing.T) {
		req := &domain.ShortenRequest{
			URL:         "https://example.com",
			CustomAlias: "taken",
		}

		existingURL := &domain.URL{
			ShortCode:   "taken",
			OriginalURL: "https://other.com",
		}

		mockRepo.On("GetURLByOriginalURL", mock.Anything, "https://example.com").
			Return(nil, errors.New("not found"))
		mockRepo.On("GetURLByShortCode", mock.Anything, "taken").
			Return(existingURL, nil)

		response, err := urlService.ShortenURL(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, service.ErrCustomAliasTaken, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestURLService_GetOriginalURL(t *testing.T) {
	mockRepo := new(mocks.MockURLRepository)
	mockCache := new(mocks.MockCacheRepository)
	logger := zaptest.NewLogger(t)

	urlService := service.NewURLService(mockRepo, mockCache, logger)

	t.Run("CacheHit", func(t *testing.T) {
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
		mockRepo.On("UpdateClickCount", mock.Anything, "abc123").Return(nil)

		originalURL, err := urlService.GetOriginalURL(context.Background(), "abc123")

		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", originalURL)
		mockCache.AssertExpectations(t)
	})

	t.Run("CacheMissDatabaseHit", func(t *testing.T) {
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
		mockRepo.On("UpdateClickCount", mock.Anything, "def456").Return(nil)

		originalURL, err := urlService.GetOriginalURL(context.Background(), "def456")

		assert.NoError(t, err)
		assert.Equal(t, "https://example.org", originalURL)
		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("URLNotFound", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, "url:notfound", mock.AnythingOfType("*domain.URL")).
			Return(errors.New("not found"))
		mockRepo.On("GetURLByShortCode", mock.Anything, "notfound").
			Return(nil, errors.New("not found"))

		originalURL, err := urlService.GetOriginalURL(context.Background(), "notfound")

		assert.Error(t, err)
		assert.Equal(t, service.ErrURLNotFound, err)
		assert.Empty(t, originalURL)
		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ExpiredURL", func(t *testing.T) {
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
		assert.Equal(t, service.ErrURLExpired, err)
		assert.Empty(t, originalURL)
		mockCache.AssertExpectations(t)
	})
}
