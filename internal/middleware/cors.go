package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow any domain (*) to call your API
		c.Header("Access-Control-Allow-Origin", "*")

		// Allow cookies/credentials to be sent with requests
		c.Header("Access-Control-Allow-Credentials", "true")

		// Tell the browser what request headers it’s allowed to send
		c.Header("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// Tell the browser what HTTP methods are allowed
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		// If the browser is just “checking” (preflight request)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // Return immediately with "No Content"
			return
		}

		// Otherwise, continue to the real handler
		c.Next()
	}
}
