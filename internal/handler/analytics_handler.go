package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/service"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
	logger           *zap.Logger
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

func (h *AnalyticsHandler) GetAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// Parse days parameter (default to 30)
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 && parsedDays <= 365 {
			days = parsedDays
		}
	}

	analytics, err := h.analyticsService.GetAnalytics(c.Request.Context(), shortCode, days)
	if err != nil {
		h.logger.Error("Failed to get analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to retrieve analytics",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
