package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/config"
)

type rateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiterMiddleware struct {
	limiters map[string]*rateLimiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

func RateLimiter(cfg config.RateLimitConfig) gin.HandlerFunc {
	rl := &RateLimiterMiddleware{
		limiters: make(map[string]*rateLimiter),
		rate:     rate.Every(cfg.Window / time.Duration(cfg.Requests)),
		burst:    cfg.Requests,
		cleanup:  time.Minute,
	}

	// Cleanup old limiters periodically
	go rl.cleanupRoutine()

	return rl.middleware
}

func (rl *RateLimiterMiddleware) middleware(c *gin.Context) {
	ip := c.ClientIP()

	limiter := rl.getLimiter(ip)
	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Rate limit exceeded",
		})
		c.Abort()
		return
	}

	c.Next()
}

func (rl *RateLimiterMiddleware) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = &rateLimiter{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = limiter
	}

	limiter.lastSeen = time.Now()
	return limiter.limiter
}

func (rl *RateLimiterMiddleware) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, limiter := range rl.limiters {
			if time.Since(limiter.lastSeen) > rl.cleanup {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}
