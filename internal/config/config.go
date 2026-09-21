package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	Env           string
	DBURL         string
	RedisURL      string
	JWTSecret     string
	JWTExpiry     time.Duration
	RefreshExpiry time.Duration
}

func Load() *Config {
	_ = godotenv.Load()
	c := &Config{
		Port:          env("APP_PORT", "8080"),
		Env:           env("APP_ENV", "development"),
		DBURL:         env("DB_URL", "postgres://thriftin:thriftin@localhost:5432/thriftin?sslmode=disable"),
		RedisURL:      env("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:     env("JWT_SECRET", "change-me"),
		JWTExpiry:     mustDuration(env("JWT_EXPIRY", "15m")),
		RefreshExpiry: mustDuration(env("REFRESH_EXPIRY", "168h")),
	}
	return c
}
func env(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
func mustDuration(s string) time.Duration { d, _ := time.ParseDuration(s); if d == 0 { d = 15 * time.Minute }; return d }
