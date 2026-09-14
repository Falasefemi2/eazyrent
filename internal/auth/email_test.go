package auth

import (
	"strings"
	"testing"
)

func TestVerificationHTML(t *testing.T) {
	html := verificationHTML("Ada Lovelace", "http://localhost:8080/auth/verify?token=abc123")
	mustContain := []string{"EASYRENT", "Confirm your email", "Ada Lovelace", "Verify email", "abc123", "24 hours", "#2563eb", "If the button doesn't work"}
	for _, s := range mustContain {
		if !strings.Contains(html, s) {
			t.Fatalf("verification html missing %q", s)
		}
	}
	if strings.Contains(html, "<script") {
		t.Fatal("html should not contain script")
	}
}

func TestVerificationHTMLEscaping(t *testing.T) {
	html := verificationHTML("<b>evil</b>", "http://localhost:8080/auth/verify?token=x")
	if strings.Contains(html, "<b>evil</b>") {
		t.Fatal("name not escaped")
	}
	if !strings.Contains(html, "&lt;b&gt;evil&lt;/b&gt;") {
		t.Fatal("escaped name missing")
	}
}

func TestResetHTML(t *testing.T) {
	html := resetHTML("Femi", "http://localhost:8080/reset-password?token=xyz")
	mustContain := []string{"Reset your password", "Femi", "Reset password", "xyz", "1 hour", "signed out"}
	for _, s := range mustContain {
		if !strings.Contains(html, s) {
			t.Fatalf("reset html missing %q", s)
		}
	}
}

func TestLinkEscaping(t *testing.T) {
	token := "a+b/c?d=e&f"
	link := verificationLink("http://localhost:8080", token)
	if !strings.Contains(link, "a%2Bb%2Fc%3Fd%3De%26f") {
		t.Fatalf("token not query-escaped: %q", link)
	}
}

func TestEmptyNameFallback(t *testing.T) {
	html := verificationHTML("", "http://localhost:8080/auth/verify?token=x")
	if !strings.Contains(html, "Hi there") {
		t.Fatalf("empty name should fallback to there, got %q", html)
	}
}
