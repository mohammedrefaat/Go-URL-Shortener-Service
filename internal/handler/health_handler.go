package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
)

type HealthHandler struct {
	urlRepo   domain.URLRepository
	cacheRepo domain.CacheRepository
}

func NewHealthHandler(urlRepo domain.URLRepository, cacheRepo domain.CacheRepository) *HealthHandler {
	return &HealthHandler{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := "healthy"
	code := http.StatusOK

	// Check database
	dbStatus := "connected"
	if err := h.urlRepo.HealthCheck(c.Request.Context()); err != nil {
		dbStatus = "disconnected"
		status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	// Check cache
	cacheStatus := "connected"
	if err := h.cacheRepo.HealthCheck(c.Request.Context()); err != nil {
		cacheStatus = "disconnected"
		status = "degraded"
		if code == http.StatusOK {
			code = http.StatusOK // Cache issues don't make service completely unhealthy
		}
	}

	c.JSON(code, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"database":  dbStatus,
		"cache":     cacheStatus,
	})
}
