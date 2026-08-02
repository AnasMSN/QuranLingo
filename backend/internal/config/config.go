package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	Env         string
	Port        string
	DatabaseURL string

	JWTAccessSecret  string
	JWTRefreshSecret string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration

	// AdminUsername/AdminPasswordHash/AdminSessionSecret gate the
	// server-rendered /admin panel behind a signed-cookie session issued by
	// its own login page. All three are optional together: if any is empty,
	// the admin panel is not mounted at all (fails closed rather than
	// mounting with an empty/guessable credential or session key).
	AdminUsername      string
	AdminPasswordHash  string
	AdminSessionSecret string
}

// Load reads backend/.env (if present) and environment variables into a Config.
// A missing .env file is not an error — real deployments set env vars directly.
func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{
		Env:              getEnv("APP_ENV", "development"),
		Port:             getEnv("APP_PORT", "8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),

		AdminUsername:      os.Getenv("ADMIN_USERNAME"),
		AdminPasswordHash:  os.Getenv("ADMIN_PASSWORD_HASH"),
		AdminSessionSecret: os.Getenv("ADMIN_SESSION_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTAccessSecret == "" || cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required")
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	cfg.AccessTTL = accessTTL
	cfg.RefreshTTL = refreshTTL

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
