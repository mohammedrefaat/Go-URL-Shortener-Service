package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/service"
)

type URLHandler struct {
	urlService *service.URLService
	logger     *zap.Logger
}

// NewURLHandler creates a new URLHandler
func NewURLHandler(urlService *service.URLService, logger *zap.Logger) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		logger:     logger,
	}
}

// ShortenURL handles the URL shortening requests
func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req domain.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil { // Bind JSON request to struct
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.urlService.ShortenURL(c.Request.Context(), &req) // Call the URL shortening service
	if err != nil {
		switch err {
		case service.ErrInvalidURL: // The provided URL is not valid or is blacklisted
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Invalid URL",
				Message: "The provided URL is not valid or is blacklisted",
				Code:    http.StatusBadRequest,
			})
		case service.ErrCustomAliasTaken: // The custom alias is already in use
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Error:   "Custom alias taken",
				Message: "The custom alias is already in use",
				Code:    http.StatusConflict,
			})
		default: // The provided URL is not valid or is blacklisted
			h.logger.Error("Failed to shorten URL", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to shorten URL",
				Code:    http.StatusInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, response) // Respond with the created short URL
}

func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode") // Get the short code from the URL

	originalURL, err := h.urlService.GetOriginalURL(c.Request.Context(), shortCode) // Get the original URL from the service
	if err != nil {
		switch err {
		case service.ErrURLNotFound: // The short URL does not exist
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Error:   "URL not found",
				Message: "The short URL does not exist",
				Code:    http.StatusNotFound,
			})
		case service.ErrURLExpired: // The short URL has expired
			c.JSON(http.StatusGone, domain.ErrorResponse{
				Error:   "URL expired",
				Message: "The short URL has expired",
				Code:    http.StatusGone,
			})
		default: // The short URL is invalid
			h.logger.Error("Failed to get original URL", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to process request",
				Code:    http.StatusInternalServerError,
			})
		}
		return
	}

	c.Redirect(http.StatusMovedPermanently, originalURL) // Redirect to the original URL
}
