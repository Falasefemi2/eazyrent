package config

import (
	"testing"
	"time"
)

func TestLoadAuth(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	t.Setenv("ACCESS_TOKEN_SECRET", "test-secret")
	t.Setenv("ACCESS_TOKEN_TTL_SECONDS", "900")
	t.Setenv("REFRESH_TOKEN_TTL_DAYS", "7")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("EMAIL_FROM", "EasyRent <test@example.com>")
	t.Setenv("APP_URL", "http://localhost:8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("AccessTokenTTL: got %v", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 7*24*time.Hour {
		t.Fatalf("RefreshTokenTTL: got %v", cfg.RefreshTokenTTL)
	}
	if cfg.ResendAPIKey != "re_test" || cfg.EmailFrom != "EasyRent <test@example.com>" || cfg.AppURL != "http://localhost:8080" {
		t.Fatalf("mail config: %+v", cfg)
	}
}

func TestLoadMissingSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	t.Setenv("ACCESS_TOKEN_SECRET", "")
	t.Setenv("RESEND_API_KEY", "re_test")

	if _, err := Load(); err == nil {
		t.Fatal("Load should fail without ACCESS_TOKEN_SECRET")
	}
}

func TestLoadMissingDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("ACCESS_TOKEN_SECRET", "test-secret")
	t.Setenv("RESEND_API_KEY", "re_test")

	if _, err := Load(); err == nil {
		t.Fatal("Load should fail without DATABASE_URL")
	}
}

func TestLoadBadInt(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	t.Setenv("ACCESS_TOKEN_SECRET", "test-secret")
	t.Setenv("ACCESS_TOKEN_TTL_SECONDS", "not-a-number")
	t.Setenv("RESEND_API_KEY", "re_test")

	if _, err := Load(); err == nil {
		t.Fatal("Load should fail on non-integer ACCESS_TOKEN_TTL_SECONDS")
	}
}

func TestLoadMissingResendKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	t.Setenv("ACCESS_TOKEN_SECRET", "test-secret")
	t.Setenv("RESEND_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load should fail without RESEND_API_KEY")
	}
}
