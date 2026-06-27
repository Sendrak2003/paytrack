package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration, all overridable via environment variables.
// Sensible defaults let the app run locally with zero setup.
type Config struct {
	Port           string
	DBDriver       string // "sqlite" or "postgres"
	DBDSN          string
	AllowedOrigins []string

	// Connection pool tuning.
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	LogLevel  string // "debug", "info", "warn", "error"
	LogFormat string // "json" or "text"

	// HTTP server timeouts.
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration

	// AI extraction (bank statement parsing). All optional; empty = regex fallback.
	AIProvider string // "", "openai", "ollama"
	AIBaseURL  string
	AIAPIKey   string
	AIModel    string
}

func Load() Config {
	driver := getEnv("DB_DRIVER", "sqlite")

	// For SQLite, multiple open connections fight over a single writer, so the
	// pool is intentionally tiny. For Postgres a real pool makes sense.
	defaultMaxOpen := 1
	if driver == "postgres" {
		defaultMaxOpen = 20
	}

	defaultDSN := "./data/payments.db"
	if driver == "postgres" {
		defaultDSN = "host=localhost user=paytrack password=paytrack dbname=paytrack port=5432 sslmode=disable"
	}

	return Config{
		Port:            getEnv("PORT", "8080"),
		DBDriver:        driver,
		DBDSN:           getEnv("DB_DSN", defaultDSN),
		AllowedOrigins:  splitNonEmpty(getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", defaultMaxOpen),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", defaultMaxOpen),
		ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_SEC", 1800)) * time.Second,
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "text"),

		ReadTimeout:       time.Duration(getEnvInt("HTTP_READ_TIMEOUT_SEC", 15)) * time.Second,
		ReadHeaderTimeout: time.Duration(getEnvInt("HTTP_READ_HEADER_TIMEOUT_SEC", 5)) * time.Second,
		WriteTimeout:      time.Duration(getEnvInt("HTTP_WRITE_TIMEOUT_SEC", 30)) * time.Second,
		IdleTimeout:       time.Duration(getEnvInt("HTTP_IDLE_TIMEOUT_SEC", 60)) * time.Second,

		// AI is fully optional. With no env set, the parser uses a regex fallback
		// and needs no API keys / tokens.
		AIProvider: getEnv("AI_PROVIDER", ""),
		AIBaseURL:  getEnv("AI_BASE_URL", ""),
		AIAPIKey:   getEnv("AI_API_KEY", ""),
		AIModel:    getEnv("AI_MODEL", "gpt-4o-mini"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitNonEmpty(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
