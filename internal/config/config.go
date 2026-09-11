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
	accessTTLSeconds, err := getenvInt("ACCESS_TOKEN_TTL_SECONDS", 3600)
	if err != nil {
		return Config{}, err
	}
	refreshTTLDays, err := getenvInt("REFRESH_TOKEN_TTL_DAYS", 30)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr:              getenv("ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		AccessTokenTTL:    time.Duration(accessTTLSeconds) * time.Second,
		RefreshTokenTTL:   time.Duration(refreshTTLDays) * 24 * time.Hour,
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

func getenvInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, errors.New(key + " must be an integer")
	}
	return n, nil
}
