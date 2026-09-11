package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPasswordResetLifecycle(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	store := NewStore(db)
	svc := NewService(
		store,
		Tokens{Secret: []byte("test-secret-please-ignore-0123456789"), AccessTTL: time.Hour},
		EmailSender{},
		30*24*time.Hour,
	)

	suffix := time.Now().UnixNano()
	email := fmt.Sprintf("reset-test-%d@example.com", suffix)
	phone := fmt.Sprintf("+234%013d", suffix%10000000000000)

	u, err := store.CreateUser(ctx, email, phone, mustHash(t, "old-password-123"), "Reset User")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, u.ID)
	})
	if err := store.MarkEmailVerified(ctx, u.ID); err != nil {
		t.Fatalf("MarkEmailVerified: %v", err)
	}

	// Unknown emails succeed silently: no account probing.
	if err := svc.RequestPasswordReset(ctx, "nobody@example.com"); err != nil {
		t.Fatalf("RequestPasswordReset unknown: want nil, got %v", err)
	}

	if err := svc.RequestPasswordReset(ctx, email); err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	var issued int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = $1`, u.ID).Scan(&issued); err != nil {
		t.Fatalf("count reset tokens: %v", err)
	}
	if issued != 1 {
		t.Fatalf("want 1 reset token issued, got %d", issued)
	}

	// A live session that the reset must kill.
	if err := store.StoreRefreshToken(ctx, u.ID, HashRefreshToken("session-hash"), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("StoreRefreshToken: %v", err)
	}

	// ResetPassword with a known raw token (inserted directly so the test
	// holds the raw value; only hashes reach the database).
	const rawToken = "reset-raw-token-for-test"
	if err := store.StorePasswordResetToken(ctx, u.ID, HashRefreshToken(rawToken), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("StorePasswordResetToken: %v", err)
	}
	if err := svc.ResetPassword(ctx, "bogus-token", "new-password-123"); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("ResetPassword bogus: want ErrInvalidResetToken, got %v", err)
	}
	if err := svc.ResetPassword(ctx, rawToken, "new-password-123"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// Old password dead, new password works.
	if _, err := svc.SignIn(ctx, email, "old-password-123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("SignIn old password: want ErrInvalidCredentials, got %v", err)
	}
	if _, err := svc.SignIn(ctx, email, "new-password-123"); err != nil {
		t.Fatalf("SignIn new password: %v", err)
	}

	// Token consumed: reuse rejected.
	if err := svc.ResetPassword(ctx, rawToken, "another-password"); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("ResetPassword reuse: want ErrInvalidResetToken, got %v", err)
	}

	// Pre-reset session revoked.
	stored, err := store.FindRefreshToken(ctx, HashRefreshToken("session-hash"))
	if err != nil {
		t.Fatalf("FindRefreshToken: %v", err)
	}
	if !stored.RevokedAt.Valid {
		t.Fatal("reset should revoke all sessions")
	}

	// Expired token rejected.
	const expiredRaw = "reset-raw-token-expired"
	if err := store.StorePasswordResetToken(ctx, u.ID, HashRefreshToken(expiredRaw), time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("StorePasswordResetToken expired: %v", err)
	}
	if err := svc.ResetPassword(ctx, expiredRaw, "new-password-123"); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("ResetPassword expired: want ErrTokenExpired, got %v", err)
	}
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	return hash
}
