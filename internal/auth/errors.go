package auth

import "errors"

var (
	ErrEmailTaken               = errors.New("email already registered")
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrEmailNotVerified         = errors.New("email not verified")
	ErrInvalidToken             = errors.New("invalid token")
	ErrTokenExpired             = errors.New("token expired")
	ErrInvalidVerificationToken = errors.New("invalid verification token")
	ErrInvalidResetToken        = errors.New("invalid reset token")
)
