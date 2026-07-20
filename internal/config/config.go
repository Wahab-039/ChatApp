package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

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
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment:      getEnv("APP_ENV", "development"),
		Port:             getEnv("PORT", "8080"),
		DatabaseHost:     getEnv("DB_HOST", "localhost"),
		DatabasePort:     getEnv("DB_PORT", "5432"),
		DatabaseUser:     strings.TrimSpace(os.Getenv("DB_USER")),
		DatabasePassword: os.Getenv("DB_PASSWORD"),
		DatabaseName:     strings.TrimSpace(os.Getenv("DB_NAME")),
		DatabaseSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	if cfg.DatabaseUser == "" || cfg.DatabasePassword == "" || cfg.DatabaseName == "" {
		return nil, errors.New("DB_USER, DB_PASSWORD, and DB_NAME are required")
	}

	return cfg, nil
}

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
