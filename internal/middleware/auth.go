package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Authorization required",
				Message: "Authorization header is required",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Invalid authorization header",
				Message: "Authorization header must be in format 'Bearer <token>'",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Invalid token",
				Message: "The provided token is invalid or expired",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Store claims in context if needed
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_claims", claims)
		}

		c.Next()
	}
}
