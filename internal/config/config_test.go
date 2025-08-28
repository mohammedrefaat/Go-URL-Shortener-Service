package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadYAML(t *testing.T) {
	t.Run("ValidYAMLConfig", func(t *testing.T) {
		// Create temporary YAML file
		yamlContent := `
server:
  port: "9090"
  environment: "production"
  base_url: "https://short.ly"

logging:
  level: "debug"

jwt:
  secret: "test-jwt-secret"

database:
  host: "db.example.com"
  port: "5432"
  user: "testuser"
  password: "testpass"
  name: "testdb"
  ssl_mode: "require"
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: "10m"

redis:
  host: "redis.example.com"
  port: "6379"
  password: "redispass"
  db: 1

rate_limit:
  requests: 200
  window: "30s"

snowflake:
  machine_id: 42

cache:
  url_ttl: "2h"
  analytics_ttl: "30m"

validation:
  malicious_domains:
    - "bad.example.com"
    - "evil.example.com"
`

		// Create temp file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "test_config.yaml")
		err := os.WriteFile(configFile, []byte(yamlContent), 0644)
		require.NoError(t, err)

		// Load configuration
		cfg, err := Load(configFile)
		require.NoError(t, err)

		// Assertions
		assert.Equal(t, "9090", cfg.Server.Port)
		assert.Equal(t, "production", cfg.Server.Environment)
		assert.Equal(t, "https://short.ly", cfg.Server.BaseURL)
		assert.Equal(t, "debug", cfg.Logging.Level)
		assert.Equal(t, "test-jwt-secret", cfg.JWT.Secret)

		// Database config
		assert.Equal(t, "db.example.com", cfg.Database.Host)
		assert.Equal(t, "testuser", cfg.Database.User)
		assert.Equal(t, "testpass", cfg.Database.Password)
		assert.Equal(t, "testdb", cfg.Database.Name)
		assert.Equal(t, "require", cfg.Database.SSLMode)
		assert.Equal(t, 50, cfg.Database.MaxOpenConns)
		assert.Equal(t, 10, cfg.Database.MaxIdleConns)
		assert.Equal(t, 10*time.Minute, cfg.Database.ConnMaxLifetime)

		// Redis config
		assert.Equal(t, "redis.example.com", cfg.Redis.Host)
		assert.Equal(t, "redispass", cfg.Redis.Password)
		assert.Equal(t, 1, cfg.Redis.DB)

		// Rate limit config
		assert.Equal(t, 200, cfg.RateLimit.Requests)
		assert.Equal(t, 30*time.Second, cfg.RateLimit.Window)

		// Snowflake config
		assert.Equal(t, int64(42), cfg.Snowflake.MachineID)

		// Cache config
		assert.Equal(t, 2*time.Hour, cfg.Cache.URLTTL)
		assert.Equal(t, 30*time.Minute, cfg.Cache.AnalyticsTTL)

		// Validation config
		assert.Contains(t, cfg.Validation.MaliciousDomains, "bad.example.com")
		assert.Contains(t, cfg.Validation.MaliciousDomains, "evil.example.com")

		// Test helper methods
		assert.Equal(t, "9090", cfg.Port())
		assert.Equal(t, "production", cfg.Environment())
		assert.Equal(t, "debug", cfg.LogLevel())
		assert.Equal(t, "https://short.ly", cfg.BaseURL())
		assert.Equal(t, "test-jwt-secret", cfg.JWTSecret())
		assert.Equal(t, int64(42), cfg.MachineID())

		// Test URL builders
		expectedDBURL := "postgres://testuser:testpass@db.example.com:5432/testdb?sslmode=require"
		assert.Equal(t, expectedDBURL, cfg.DatabaseURL())

		expectedRedisURL := "redis://:redispass@redis.example.com:6379/1"
		assert.Equal(t, expectedRedisURL, cfg.RedisURL())
	})

	t.Run("FileNotFound", func(t *testing.T) {
		cfg, err := Load("nonexistent.yaml")
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(configFile, []byte("invalid: yaml: content: ["), 0644)
		require.NoError(t, err)

		cfg, err := Load(configFile)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		testCases := []struct {
			name     string
			yaml     string
			errorMsg string
		}{
			{
				name: "EmptyPort",
				yaml: `
server:
  port: ""
  environment: "development"
  base_url: "http://localhost"
`,
				errorMsg: "server port is required",
			},
			{
				name: "EmptyDatabaseHost",
				yaml: `
server:
  port: "8080"
database:
  host: ""
`,
				errorMsg: "database host is required",
			},
			{
				name: "InvalidMachineID",
				yaml: `
server:
  port: "8080"
database:
  host: "localhost"
  user: "postgres"
  name: "test"
snowflake:
  machine_id: 2000
`,
				errorMsg: "snowflake machine_id must be between 0 and 1023",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "test.yaml")
				err := os.WriteFile(configFile, []byte(tc.yaml), 0644)
				require.NoError(t, err)

				cfg, err := Load(configFile)
				assert.Error(t, err)
				assert.Nil(t, cfg)
				assert.Contains(t, err.Error(), tc.errorMsg)
			})
		}
	})
}

func TestConfig_LoadFromEnv(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		cfg := LoadFromEnv()

		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "development", cfg.Server.Environment)
		assert.Equal(t, "info", cfg.Logging.Level)
		assert.Equal(t, "http://localhost:8080", cfg.Server.BaseURL)
		assert.Equal(t, 100, cfg.RateLimit.Requests)
		assert.Equal(t, 60*time.Second, cfg.RateLimit.Window)
		assert.Equal(t, int64(1), cfg.Snowflake.MachineID)
	})

	t.Run("EnvironmentOverrides", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PORT", "9090")
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("RATE_LIMIT_REQUESTS", "200")
		os.Setenv("MACHINE_ID", "5")
		defer func() {
			os.Unsetenv("PORT")
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("RATE_LIMIT_REQUESTS")
			os.Unsetenv("MACHINE_ID")
		}()

		cfg := LoadFromEnv()

		assert.Equal(t, "9090", cfg.Server.Port)
		assert.Equal(t, "production", cfg.Server.Environment)
		assert.Equal(t, "debug", cfg.Logging.Level)
		assert.Equal(t, 200, cfg.RateLimit.Requests)
		assert.Equal(t, int64(5), cfg.Snowflake.MachineID)
	})
}

func TestConfig_URLBuilders(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "dbhost",
			Port:     "5432",
			User:     "dbuser",
			Password: "dbpass",
			Name:     "dbname",
			SSLMode:  "require",
		},
		Redis: RedisConfig{
			Host:     "redishost",
			Port:     "6379",
			Password: "redispass",
			DB:       2,
		},
	}

	t.Run("DatabaseURL", func(t *testing.T) {
		expected := "postgres://dbuser:dbpass@dbhost:5432/dbname?sslmode=require"
		assert.Equal(t, expected, cfg.DatabaseURL())
	})

	t.Run("RedisURLWithPassword", func(t *testing.T) {
		expected := "redis://:redispass@redishost:6379/2"
		assert.Equal(t, expected, cfg.RedisURL())
	})

	t.Run("RedisURLWithoutPassword", func(t *testing.T) {
		cfg.Redis.Password = ""
		expected := "redis://redishost:6379/2"
		assert.Equal(t, expected, cfg.RedisURL())
	})
}
