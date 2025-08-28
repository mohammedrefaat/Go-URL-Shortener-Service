package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/logger"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/config"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/handler"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/middleware"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/service"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/store/postgres"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/store/redis"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
		return
	}

	// Initialize logger
	log := logger.New(cfg.Logging.Level)
	defer log.Sync()

	// Initialize repositories
	dbRepo, err := postgres.NewURLRepository(cfg.DatabaseURL())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbRepo.Close()

	cacheRepo, err := redis.NewCacheRepository(cfg.RedisURL())
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer cacheRepo.Close()

	// Initialize services
	urlService := service.NewURLService(dbRepo, cacheRepo, log, cfg)
	analyticsService := service.NewAnalyticsService(dbRepo, cacheRepo, log)

	// Initialize handlers
	urlHandler := handler.NewURLHandler(urlService, log)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService, log)
	healthHandler := handler.NewHealthHandler(dbRepo, cacheRepo)

	// Setup routes
	router := setupRoutes(cfg, urlHandler, analyticsHandler, healthHandler, log)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port(),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	log.Info("Server started", zap.String("port", cfg.Port()))

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}

func setupRoutes(cfg *config.Config, urlHandler *handler.URLHandler, analyticsHandler *handler.AnalyticsHandler, healthHandler *handler.HealthHandler, log *zap.Logger) *gin.Engine {
	if cfg.Environment() == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", healthHandler.HealthCheck)

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.Requests, cfg.RateLimit.Window)

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.Use(rateLimiter.Middleware())
	{
		v1.POST("/shorten", urlHandler.ShortenURL)
		v1.GET("/analytics/:shortCode", middleware.JWTAuth(cfg.JWTSecret()), analyticsHandler.GetAnalytics)
	}

	// Redirect route (no rate limiting for better UX)
	router.GET("/:shortCode", urlHandler.RedirectURL)

	return router
}
