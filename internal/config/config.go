package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	AccessTokenSecret string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Addr:              getenv("ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		AccessTokenTTL:    time.Duration(getenvInt("ACCESS_TOKEN_TTL_SECONDS", 3600)) * time.Second,
		RefreshTokenTTL:   time.Duration(getenvInt("REFRESH_TOKEN_TTL_DAYS", 30)) * 24 * time.Hour,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.AccessTokenSecret == "" {
		return Config{}, errors.New("ACCESS_TOKEN_SECRET is required")
	}
	if cfg.AccessTokenTTL <= 0 {
		return Config{}, errors.New("ACCESS_TOKEN_TTL_SECONDS must be > 0")
	}
	if cfg.RefreshTokenTTL <= 0 {
		return Config{}, errors.New("REFRESH_TOKEN_TTL_DAYS must be > 0")
	}

	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
