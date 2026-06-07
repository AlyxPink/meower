// Package config loads and validates the api server configuration from the
// environment. Centralizing env access here (rather than scattering os.Getenv
// across the codebase) makes the full set of configuration knobs discoverable
// in one place and gives each a typed, defaulted accessor.
package config

import (
	"os"

	"github.com/charmbracelet/log"
)

// Config holds the api server configuration.
type Config struct {
	// Environment is "development" or "production"; it drives log formatting
	// (JSON in prod) and tracing/sampling defaults.
	Environment string

	// LogLevel is one of debug|info|warn|error.
	LogLevel string

	// DatabaseURL is the PostgreSQL connection string.
	DatabaseURL string

	// RedisURL is the Redis connection string (used by the rate limiter).
	RedisURL string

	// WebBaseURL is the externally-reachable origin of the web server, used to
	// build absolute links (e.g. in emails). Empty in development.
	WebBaseURL string
}

// Load reads configuration from environment variables, applying defaults.
func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
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
