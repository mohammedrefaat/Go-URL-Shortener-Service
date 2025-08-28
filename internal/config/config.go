package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Logging    LoggingConfig    `yaml:"logging"`
	JWT        JWTConfig        `yaml:"jwt"`
	Database   DatabaseConfig   `yaml:"database"`
	Redis      RedisConfig      `yaml:"redis"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
	Snowflake  SnowflakeConfig  `yaml:"snowflake"`
	Cache      CacheConfig      `yaml:"cache"`
	Validation ValidationConfig `yaml:"validation"`
}

type ServerConfig struct {
	Port        string `yaml:"port"`
	Environment string `yaml:"environment"`
	BaseURL     string `yaml:"base_url"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RateLimitConfig struct {
	Requests int           `yaml:"requests"`
	Window   time.Duration `yaml:"window"`
}

type SnowflakeConfig struct {
	MachineID int64 `yaml:"machine_id"`
}

type CacheConfig struct {
	URLTTL       time.Duration `yaml:"url_ttl"`
	AnalyticsTTL time.Duration `yaml:"analytics_ttl"`
}

type ValidationConfig struct {
	MaliciousDomains []string `yaml:"malicious_domains"`
}

// Legacy fields for backward compatibility
func (c *Config) Port() string        { return c.Server.Port }
func (c *Config) Environment() string { return c.Server.Environment }
func (c *Config) LogLevel() string    { return c.Logging.Level }
func (c *Config) BaseURL() string     { return c.Server.BaseURL }
func (c *Config) JWTSecret() string   { return c.JWT.Secret }
func (c *Config) MachineID() int64    { return c.Snowflake.MachineID }

// DatabaseURL builds the PostgreSQL connection string
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// RedisURL builds the Redis connection string
func (c *Config) RedisURL() string {
	if c.Redis.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%s/%d",
			c.Redis.Password,
			c.Redis.Host,
			c.Redis.Port,
			c.Redis.DB,
		)
	}
	return fmt.Sprintf("redis://%s:%s/%d",
		c.Redis.Host,
		c.Redis.Port,
		c.Redis.DB,
	)
}

// Load loads configuration from YAML file
func Load(configPath string) (*Config, error) {
	// Default config path if not provided
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Read YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// LoadFromEnv provides backward compatibility for environment variable configuration
func LoadFromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
			BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "244159"),
			Name:            getEnv("DB_NAME", "urlshortener"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   time.Duration(getEnvAsInt("RATE_LIMIT_WINDOW", 60)) * time.Second,
		},
		Snowflake: SnowflakeConfig{
			MachineID: int64(getEnvAsInt("MACHINE_ID", 1)),
		},
		Cache: CacheConfig{
			URLTTL:       1 * time.Hour,
			AnalyticsTTL: 15 * time.Minute,
		},
		Validation: ValidationConfig{
			MaliciousDomains: []string{
				"malware.example.com",
				"phishing.example.com",
			},
		},
	}
}

func (c *Config) validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Snowflake.MachineID < 0 || c.Snowflake.MachineID > 1023 {
		return fmt.Errorf("snowflake machine_id must be between 0 and 1023")
	}
	if c.RateLimit.Requests <= 0 {
		return fmt.Errorf("rate_limit requests must be positive")
	}
	if c.RateLimit.Window <= 0 {
		return fmt.Errorf("rate_limit window must be positive")
	}
	return nil
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
