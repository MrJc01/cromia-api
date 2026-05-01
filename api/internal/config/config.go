package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	DBDriver         string
	DBDSN            string
	MasterAPIKey     string
	PythonWorkers    int
	PythonTimeoutSec int
	RateLimitBackend string // "memory" | "redis"
	RateLimitPerKey  int
	RateLimitWindow  time.Duration
	LogLevel         string // "debug" | "info" | "warn" | "error"
}

func Load() *Config {
	c := &Config{
		Port:             getEnv("API_PORT", "8080"),
		DBDriver:         getEnv("DB_DRIVER", "sqlite3"),
		DBDSN:            getEnv("DB_DSN", "data.db"),
		MasterAPIKey:     getEnv("MASTER_API_KEY", "crom_sk_master_secret_123456"),
		PythonWorkers:    getEnvInt("PYTHON_WORKERS", 2),
		PythonTimeoutSec: getEnvInt("PYTHON_TIMEOUT_SEC", 60),
		RateLimitBackend: getEnv("RATE_LIMIT_BACKEND", "memory"),
		RateLimitPerKey:  getEnvInt("RATE_LIMIT_PER_KEY", 60),
		RateLimitWindow:  time.Minute,
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
	return c
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	s := getEnv(key, "")
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
