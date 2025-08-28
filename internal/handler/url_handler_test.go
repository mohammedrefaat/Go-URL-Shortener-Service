package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/config"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/service"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/store/mocks"
)

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestURLHandler_ShortenURL(t *testing.T) {
	mockRepo := new(mocks.MockURLRepository)
	mockCache := new(mocks.MockCacheRepository)
	logger := zaptest.NewLogger(t)

	urlService := service.NewURLService(mockRepo, mockCache, logger, &config.Config{
		Server: config.ServerConfig{
			BaseURL: "http://localhost:8080",
		},
		Snowflake: config.SnowflakeConfig{
			MachineID: 1,
		},
	})
	urlHandler := NewURLHandler(urlService, logger)

	router := setupGin()
	router.POST("/shorten", urlHandler.ShortenURL)

	t.Run("SuccessfulShorten", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockRepo.On("GetURLByOriginalURL", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
		mockRepo.On("CreateURL", mock.Anything, mock.Anything).Return(nil)
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		reqBody := domain.ShortenRequest{
			URL: "https://example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response domain.ShortenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", response.OriginalURL)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		reqBody := domain.ShortenRequest{
			URL: "not-a-url",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestURLHandler_RedirectURL(t *testing.T) {
	mockRepo := new(mocks.MockURLRepository)
	mockCache := new(mocks.MockCacheRepository)
	logger := zaptest.NewLogger(t)

	urlService := service.NewURLService(mockRepo, mockCache, logger, nil)
	urlHandler := NewURLHandler(urlService, logger)

	router := setupGin()
	router.GET("/:shortCode", urlHandler.RedirectURL)

	t.Run("SuccessfulRedirect", func(t *testing.T) {
		url := &domain.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			CreatedAt:   time.Now(),
		}

		mockCache.On("Get", mock.Anything, "url:abc123", mock.Anything).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*domain.URL)
				*arg = *url
			}).Return(nil)
		mockCache.On("Increment", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		req := httptest.NewRequest("GET", "/abc123", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Location"))
	})

	t.Run("URLNotFound", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, "url:notfound", mock.Anything).Return(errors.New("not found"))
		mockRepo.On("GetURLByShortCode", mock.Anything, "notfound").Return(nil, errors.New("not found"))

		req := httptest.NewRequest("GET", "/notfound", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHealthHandler_HealthCheck(t *testing.T) {
	mockRepo := new(mocks.MockURLRepository)
	mockCache := new(mocks.MockCacheRepository)

	healthHandler := NewHealthHandler(mockRepo, mockCache)
	router := setupGin()
	router.GET("/health", healthHandler.HealthCheck)

	t.Run("HealthyServices", func(t *testing.T) {
		mockRepo.On("HealthCheck", mock.Anything).Return(nil)
		mockCache.On("HealthCheck", mock.Anything).Return(nil)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "connected", response["database"])
		assert.Equal(t, "connected", response["cache"])
	})

	/*t.Run("DatabaseUnhealthy", func(t *testing.T) {
		mockRepo.On("HealthCheck", mock.Anything).Return(errors.New("db error"))
		mockCache.On("HealthCheck", mock.Anything).Return(nil)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "unhealthy", response["status"])
		assert.Equal(t, "disconnected", response["database"])
	})*/

	/*t.Run("CacheUnhealthy", func(t *testing.T) {
		mockRepo.On("HealthCheck", mock.Anything).Return(nil)
		mockCache.On("HealthCheck", mock.Anything).Return(errors.New("cache error"))

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Cache issues don't make service completely unhealthy

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "degraded", response["status"])
		assert.Equal(t, "disconnected", response["cache"])
	})*/
}

func TestAnalyticsHandler_GetAnalytics(t *testing.T) {
	mockRepo := new(mocks.MockURLRepository)
	mockCache := new(mocks.MockCacheRepository)
	logger := zaptest.NewLogger(t)

	analyticsService := service.NewAnalyticsService(mockRepo, mockCache, logger)
	analyticsHandler := NewAnalyticsHandler(analyticsService, logger)

	router := setupGin()
	router.GET("/analytics/:shortCode", analyticsHandler.GetAnalytics)

	t.Run("SuccessfulAnalytics", func(t *testing.T) {
		analytics := &domain.AnalyticsResponse{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			ClickCount:  10,
			CreatedAt:   time.Now(),
			DailyStats:  []domain.DailyStat{},
		}

		mockCache.On("Get", mock.Anything, "analytics:abc123:30", mock.Anything).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*domain.AnalyticsResponse)
				*arg = *analytics
			}).Return(nil)

		req := httptest.NewRequest("GET", "/analytics/abc123", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.AnalyticsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "abc123", response.ShortCode)
		assert.Equal(t, int64(10), response.ClickCount)
	})

	t.Run("AnalyticsNotFound", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, "analytics:notfound:30", mock.Anything).Return(errors.New("not found"))
		mockRepo.On("GetAnalytics", mock.Anything, "notfound", 30).Return(nil, errors.New("not found"))

		req := httptest.NewRequest("GET", "/analytics/notfound", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("CustomDaysParameter", func(t *testing.T) {
		analytics := &domain.AnalyticsResponse{
			ShortCode:   "def456",
			OriginalURL: "https://example.org",
			ClickCount:  5,
			CreatedAt:   time.Now(),
			DailyStats:  []domain.DailyStat{},
		}

		mockCache.On("Get", mock.Anything, "analytics:def456:7", mock.Anything).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*domain.AnalyticsResponse)
				*arg = *analytics
			}).Return(nil)

		req := httptest.NewRequest("GET", "/analytics/def456?days=7", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
