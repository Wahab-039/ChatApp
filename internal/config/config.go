package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment      string
	Port             string
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string
	JWTSecret        string
	JWTAccessTTL     time.Duration
	LoginRateLimit   int
	LoginRateWindow  time.Duration
}

// Load reads and validates application configuration from the environment.
func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtAccessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "24h"))
	if err != nil || jwtAccessTTL <= 0 {
		return nil, errors.New("JWT_ACCESS_TTL must be a positive Go duration")
	}
	loginRateLimit, err := parsePositiveInt(getEnv("LOGIN_RATE_LIMIT", "10"))
	if err != nil {
		return nil, fmt.Errorf("LOGIN_RATE_LIMIT: %w", err)
	}
	loginRateWindow, err := time.ParseDuration(getEnv("LOGIN_RATE_WINDOW", "1m"))
	if err != nil || loginRateWindow <= 0 {
		return nil, errors.New("LOGIN_RATE_WINDOW must be a positive Go duration")
	}

	cfg := &Config{
		Environment:      getEnv("APP_ENV", "development"),
		Port:             getEnv("PORT", "8080"),
		DatabaseHost:     getEnv("DB_HOST", "localhost"),
		DatabasePort:     getEnv("DB_PORT", "5432"),
		DatabaseUser:     strings.TrimSpace(os.Getenv("DB_USER")),
		DatabasePassword: os.Getenv("DB_PASSWORD"),
		DatabaseName:     strings.TrimSpace(os.Getenv("DB_NAME")),
		DatabaseSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:        strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTAccessTTL:     jwtAccessTTL,
		LoginRateLimit:   loginRateLimit,
		LoginRateWindow:  loginRateWindow,
	}

	if cfg.DatabaseUser == "" || cfg.DatabasePassword == "" || cfg.DatabaseName == "" {
		return nil, errors.New("DB_USER, DB_PASSWORD, and DB_NAME are required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	return cfg, nil
}

// DatabaseURL returns a PostgreSQL connection string for the configured database.
func (c *Config) DatabaseURL() string {
	return (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DatabaseUser, c.DatabasePassword),
		Host:   fmt.Sprintf("%s:%s", c.DatabaseHost, c.DatabasePort),
		Path:   c.DatabaseName,
	}).String() + "?sslmode=" + c.DatabaseSSLMode
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func parsePositiveInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, errors.New("must be a positive integer")
	}
	return parsed, nil
}
