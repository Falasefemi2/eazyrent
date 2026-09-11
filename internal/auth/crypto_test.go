package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordRoundtrip(t *testing.T) {
	hash, err := HashPassword("correct-horse-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if strings.Contains(hash, "correct-horse-123") {
		t.Fatal("hash must not contain the raw password")
	}

	ok, err := VerifyPassword(hash, "correct-horse-123")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("correct password should verify")
	}

	ok, err = VerifyPassword(hash, "wrong-password")
	if err != nil {
		t.Fatalf("VerifyPassword mismatch: %v", err)
	}
	if ok {
		t.Fatal("wrong password must not verify")
	}
}

func TestAccessTokenRoundtrip(t *testing.T) {
	tokens := Tokens{Secret: []byte("test-secret-please-ignore-0123456789"), AccessTTL: time.Hour}
	id := uuid.New()

	raw, err := tokens.SignAccessToken(id, "user@example.com")
	if err != nil {
		t.Fatalf("SignAccessToken: %v", err)
	}

	payload, err := tokens.VerifyAccessToken(raw)
	if err != nil {
		t.Fatalf("VerifyAccessToken: %v", err)
	}
	if payload.UserID != id || payload.Email != "user@example.com" {
		t.Fatalf("payload mismatch: %+v", payload)
	}

	if _, err := tokens.VerifyAccessToken(raw + "tampered"); err == nil {
		t.Fatal("tampered token must not verify")
	}

	other := Tokens{Secret: []byte("different-secret-please-ignore-0123"), AccessTTL: time.Hour}
	if _, err := other.VerifyAccessToken(raw); err == nil {
		t.Fatal("token signed with another secret must not verify")
	}
}

func TestAccessTokenExpiry(t *testing.T) {
	tokens := Tokens{Secret: []byte("test-secret-please-ignore-0123456789"), AccessTTL: -time.Second}

	raw, err := tokens.SignAccessToken(uuid.New(), "user@example.com")
	if err != nil {
		t.Fatalf("SignAccessToken: %v", err)
	}
	if _, err := tokens.VerifyAccessToken(raw); err != ErrTokenExpired {
		t.Fatalf("expired token: want ErrTokenExpired, got %v", err)
	}
}

func TestRefreshTokenShape(t *testing.T) {
	raw, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if len(raw) != 128 {
		t.Fatalf("64 bytes as hex should be 128 chars, got %d", len(raw))
	}

	other, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if raw == other {
		t.Fatal("consecutive refresh tokens must differ")
	}

	hash := HashRefreshToken(raw)
	if len(hash) != 64 || hash == raw {
		t.Fatal("stored hash should be 64 hex chars, distinct from the raw token")
	}
	if HashRefreshToken(raw) != hash {
		t.Fatal("hashing must be deterministic")
	}
}

func TestVerificationTokenShape(t *testing.T) {
	raw, err := GenerateVerificationToken()
	if err != nil {
		t.Fatalf("GenerateVerificationToken: %v", err)
	}
	if len(raw) != 64 {
		t.Fatalf("32 bytes as hex should be 64 chars, got %d", len(raw))
	}
}
