package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()           // record start time
		path := c.Request.URL.Path    // requested path (/shorten, /health, etc.)
		raw := c.Request.URL.RawQuery // query string (?foo=bar)

		c.Next() // let Gin handle the request (call the next middleware/handler)

		// After handler finishes, we calculate metrics:
		latency := time.Since(start)    // how long it took
		clientIP := c.ClientIP()        // request’s IP address
		method := c.Request.Method      // GET/POST/etc
		statusCode := c.Writer.Status() // HTTP status (200, 404, etc)
		bodySize := c.Writer.Size()     // response size in bytes

		// Reconstruct full path with query string if exists
		if raw != "" {
			path = path + "?" + raw
		}

		// Log structured fields with zap
		logger.Info("HTTP Request",
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Int("body_size", bodySize),
			zap.Duration("latency", latency),
		)
	}
}
