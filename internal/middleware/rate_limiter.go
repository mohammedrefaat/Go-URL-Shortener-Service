package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type tokenBucket struct {
	tokens     int
	capacity   int
	refill     time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

type RateLimiter struct {
	buckets map[string]*tokenBucket
	mutex   sync.RWMutex
	rate    int
	window  time.Duration
}

func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    requestsPerWindow,
		window:  window,
	}

	// Clean up expired buckets periodically
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rl.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
				"code":    http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mutex.RLock()
	bucket, exists := rl.buckets[key]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		bucket = &tokenBucket{
			tokens:     rl.rate,
			capacity:   rl.rate,
			refill:     rl.window,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
		rl.mutex.Unlock()
	}

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	// Refill tokens if enough time has passed
	now := time.Now()
	if now.Sub(bucket.lastRefill) >= bucket.refill {
		bucket.tokens = bucket.capacity
		bucket.lastRefill = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			bucket.mutex.Lock()
			if now.Sub(bucket.lastRefill) > rl.window*2 {
				delete(rl.buckets, key)
			}
			bucket.mutex.Unlock()
		}
		rl.mutex.Unlock()
	}
}
