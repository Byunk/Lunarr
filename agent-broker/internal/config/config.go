package config

import (
	"log/slog"
	"os"
	"strconv"
)

// Config holds application configuration from environment variables.
type Config struct {
	// Port is the HTTP server port.
	Port int
	// LogLevel is the minimum log level for logging.
	LogLevel slog.Level
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:     getEnvInt("PORT", 8080),
		LogLevel: getEnvLogLevel("LOG_LEVEL", slog.LevelInfo),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvLogLevel(key string, defaultValue slog.Level) slog.Level {
	value := getEnv(key, "")
	switch value {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	default:
		return defaultValue
	}
}
