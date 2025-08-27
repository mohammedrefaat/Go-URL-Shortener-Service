package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		cfg := Load()

		assert.Equal(t, "8080", cfg.Port)
		assert.Equal(t, "development", cfg.Environment)
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
		assert.Equal(t, 100, cfg.RateLimit.Requests)
		assert.Equal(t, 60*time.Second, cfg.RateLimit.Window)
	})

	t.Run("EnvironmentOverrides", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PORT", "9090")
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("RATE_LIMIT_REQUESTS", "200")
		defer func() {
			os.Unsetenv("PORT")
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("RATE_LIMIT_REQUESTS")
		}()

		cfg := Load()

		assert.Equal(t, "9090", cfg.Port)
		assert.Equal(t, "production", cfg.Environment)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, 200, cfg.RateLimit.Requests)
	})
}
