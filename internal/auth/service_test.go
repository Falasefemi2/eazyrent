package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestAuthServiceLifecycle runs the full auth flow against the dev database:
// sign-up (incl. duplicate), sign-in gates, email verification, access
// token check, refresh rotation with reuse detection, and sign-out.
func TestAuthServiceLifecycle(t *testing.T) {
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
	email := fmt.Sprintf("svc-test-%d@example.com", suffix)
	phone := fmt.Sprintf("+234%013d", suffix%10000000000000)

	pair, err := svc.SignUp(ctx, email, "correct-horse-123", phone, "Service User")
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE email = $1`, email)
	})
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("SignUp should return both tokens")
	}

	otherPhone := fmt.Sprintf("+234%013d", (suffix+1)%10000000000000)
	if _, err := svc.SignUp(ctx, email, "another-password", otherPhone, "Dupe User"); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate SignUp: want ErrEmailTaken, got %v", err)
	}

	if _, err := svc.SignIn(ctx, email, "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("SignIn wrong password: want ErrInvalidCredentials, got %v", err)
	}
	if _, err := svc.SignIn(ctx, "nobody@example.com", "whatever"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("SignIn unknown email: want ErrInvalidCredentials, got %v", err)
	}
	if _, err := svc.SignIn(ctx, email, "correct-horse-123"); !errors.Is(err, ErrEmailNotVerified) {
		t.Fatalf("SignIn unverified: want ErrEmailNotVerified, got %v", err)
	}

	if err := svc.VerifyEmail(ctx, "not-a-real-token"); !errors.Is(err, ErrInvalidVerificationToken) {
		t.Fatalf("VerifyEmail bogus: want ErrInvalidVerificationToken, got %v", err)
	}

	u, err := store.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if !u.VerificationToken.Valid {
		t.Fatal("new user should hold a verification token")
	}
	if err := svc.VerifyEmail(ctx, u.VerificationToken.String); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}

	pair2, err := svc.SignIn(ctx, email, "correct-horse-123")
	if err != nil {
		t.Fatalf("SignIn after verify: %v", err)
	}

	payload, err := svc.Tokens.VerifyAccessToken(pair2.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccessToken: %v", err)
	}
	if payload.UserID != u.ID || payload.Email != email {
		t.Fatal("access token payload does not match the user")
	}

	pair3, err := svc.Refresh(ctx, pair2.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if pair3.RefreshToken == pair2.RefreshToken {
		t.Fatal("Refresh should rotate the refresh token")
	}

	// Reusing the rotated-out token looks like theft: rejected, and the
	// rotated-in token is revoked along with everything else.
	if _, err := svc.Refresh(ctx, pair2.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Refresh reuse: want ErrInvalidToken, got %v", err)
	}
	if _, err := svc.Refresh(ctx, pair3.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Refresh after revoke-all: want ErrInvalidToken, got %v", err)
	}

	// Sign-out on a fresh account so the theft lockdown above is out of scope.
	email2 := fmt.Sprintf("svc-test-out-%d@example.com", suffix)
	phone2 := fmt.Sprintf("+234%013d", (suffix+2)%10000000000000)
	pair4, err := svc.SignUp(ctx, email2, "correct-horse-123", phone2, "Logout User")
	if err != nil {
		t.Fatalf("SignUp 2: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE email = $1`, email2)
	})

	if err := svc.SignOut(ctx, pair4.RefreshToken); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	stored, err := store.FindRefreshToken(ctx, HashRefreshToken(pair4.RefreshToken))
	if err != nil {
		t.Fatalf("FindRefreshToken after SignOut: %v", err)
	}
	if !stored.RevokedAt.Valid {
		t.Fatal("token should be revoked after SignOut")
	}
	if err := svc.SignOut(ctx, "unknown-token"); err != nil {
		t.Fatalf("SignOut unknown token should be a no-op, got %v", err)
	}
}
