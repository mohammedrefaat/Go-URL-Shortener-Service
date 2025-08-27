package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	Environment string
	LogLevel    string
	BaseURL     string
	JWTSecret   string
	DatabaseURL string
	RedisURL    string
	RateLimit   RateLimitConfig
	MachineID   int64
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		DatabaseURL: buildDatabaseURL(),
		RedisURL:    buildRedisURL(),
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   time.Duration(getEnvAsInt("RATE_LIMIT_WINDOW", 60)) * time.Second,
		},
		MachineID: 1,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func buildDatabaseURL() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "244159")
	dbname := getEnv("DB_NAME", "urlshortener")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=" + sslmode
}

func buildRedisURL() string {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")

	if password != "" {
		return "redis://:" + password + "@" + host + ":" + port + "/0"
	}
	return "redis://" + host + ":" + port + "/0"
}
