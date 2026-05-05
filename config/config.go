package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Server
	Domain string
	IP     string
	Listen string

	// Database
	Driver string
	DSN    string

	// Features
	Swagger   bool
	WithGuest bool
	TestMode  bool

	// Auth
	AuthExpire time.Duration

	// Defaults
	DefaultCleanInterval         int64
	DefaultQueryApiMaxItem       int
	DefaultMaxCallbackErrorCount int64
	DefaultLanguage              string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Domain:                       getEnv("DOMAIN", "example.com"),
		IP:                           getEnv("IP", "0.0.0.0"),
		Listen:                       getEnv("LISTEN", ":8080"),
		Driver:                       getEnv("DB_DRIVER", "sqlite"),
		DSN:                          getEnv("DB_DSN", "file:godnslog.db?cache=shared&mode=rwc"),
		Swagger:                      getBoolEnv("SWAGGER", false),
		WithGuest:                    getBoolEnv("WITH_GUEST", false),
		TestMode:                     getBoolEnv("TEST_MODE", false),
		AuthExpire:                   getDurationEnv("AUTH_EXPIRE", 24*time.Hour),
		DefaultCleanInterval:         getInt64Env("DEFAULT_CLEAN_INTERVAL", 3600),
		DefaultQueryApiMaxItem:       getIntEnv("DEFAULT_QUERY_API_MAX_ITEM", 1000),
		DefaultMaxCallbackErrorCount: getInt64Env("DEFAULT_MAX_CALLBACK_ERROR_COUNT", 10),
		DefaultLanguage:              getEnv("DEFAULT_LANGUAGE", "en-US"),
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if c.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if c.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}
	if c.AuthExpire <= 0 {
		return fmt.Errorf("auth expire must be positive")
	}
	return nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv gets boolean environment variable with default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getIntEnv gets integer environment variable with default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getInt64Env gets int64 environment variable with default value
func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDurationEnv gets duration environment variable with default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}
