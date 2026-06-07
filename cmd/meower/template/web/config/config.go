// Package config loads the web server configuration from the environment.
package config

import (
	"os"

	"github.com/charmbracelet/log"
)

// Config holds the web server configuration.
type Config struct {
	// Environment is "development" or "production".
	Environment string

	// LogLevel is one of debug|info|warn|error.
	LogLevel string

	// RedisURL backs the session store and rate limiter.
	RedisURL string

	// WebBaseURL is the externally-reachable origin of the web server, used to
	// build absolute links. Empty in development.
	WebBaseURL string
}

// Load reads configuration from environment variables, applying defaults.
func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		RedisURL:    getEnv("REDIS_URL", ""),
		WebBaseURL:  getEnv("WEB_BASE_URL", ""),
	}
}

// ApplyLogLevel sets the global charmbracelet/log level from the config.
func (c *Config) ApplyLogLevel() {
	switch c.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Warn("Unknown log level, defaulting to info", "level", c.LogLevel)
		log.SetLevel(log.InfoLevel)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
