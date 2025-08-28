package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create rate limiter: 2 requests per 5 seconds
	rateLimiter := NewRateLimiter(2, 5*time.Second)

	router := gin.New()
	router.Use(rateLimiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// First request should pass
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should pass
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Third request should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	router := gin.New()
	router.Use(JWTAuth(secret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "protected"})
	})

	t.Run("NoAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidHeader")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ValidToken", func(t *testing.T) {
		token, err := utils.GenerateJWT(secret, "user123", time.Hour)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create zap observer to capture logs
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	// Create a test router with middleware
	router := gin.New()
	router.Use(Logger(logger))
	router.GET("/ping", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond) // simulate latency
		c.String(http.StatusOK, "pong")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/ping?foo=bar", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Validate response
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "pong", w.Body.String())

	// Validate logs
	logs := recorded.All()
	require.Len(t, logs, 1, "expected one log entry")

	entry := logs[0]
	require.Equal(t, "HTTP Request", entry.Message)
	require.Equal(t, "GET", entry.ContextMap()["method"])
	require.Equal(t, "/ping?foo=bar", entry.ContextMap()["path"])
	require.Equal(t, int64(http.StatusOK), entry.ContextMap()["status"].(int64))
	require.Contains(t, entry.ContextMap()["client_ip"].(string), "127.0.0.1")
	require.GreaterOrEqual(t, entry.ContextMap()["latency"].(time.Duration).Milliseconds(), int64(10))
}
