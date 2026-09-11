package auth

import "log"

type EmailSender struct{}

func (EmailSender) SendVerificationEmail(to, fullName, token string) error {
	log.Printf("auth email=verification to=%s name=%q token=%s (dummy sender: not delivered)", to, fullName, token)
	return nil
}

func (EmailSender) SendPasswordResetEmail(to, fullName, token string) error {
	log.Printf("auth email=password_reset to=%s name=%q token=%s (dummy sender: not delivered)", to, fullName, token)
	return nil
}
