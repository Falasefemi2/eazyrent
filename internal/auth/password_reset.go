package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        uuid.UUID    `json:"id"`
	UserID    uuid.UUID    `json:"user_id"`
	ExpiresAt time.Time    `json:"expires_at"`
	UsedAt    sql.NullTime `json:"used_at"`
	CreatedAt time.Time    `json:"created_at"`
}

const storePasswordResetTokenQuery = `
	INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)
`

func (s Store) StorePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, storePasswordResetTokenQuery, userID, tokenHash, expiresAt)
	return err
}

const findPasswordResetTokenQuery = `
	SELECT id, user_id, expires_at, used_at, created_at
	FROM password_reset_tokens
	WHERE token_hash = $1
	LIMIT 1
`

func (s Store) FindPasswordResetToken(ctx context.Context, tokenHash string) (PasswordResetToken, error) {
	var t PasswordResetToken
	err := s.db.QueryRowContext(ctx, findPasswordResetTokenQuery, tokenHash).
		Scan(&t.ID, &t.UserID, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return PasswordResetToken{}, ErrInvalidResetToken
	}
	if err != nil {
		return PasswordResetToken{}, err
	}
	return t, nil
}

const markPasswordResetTokenUsedQuery = `
	UPDATE password_reset_tokens
	SET used_at = now()
	WHERE id = $1
`

func (s Store) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, markPasswordResetTokenUsedQuery, id)
	return err
}

const updatePasswordQuery = `
	UPDATE users
	SET password_hash = $2, updated_at = now()
	WHERE id = $1
`

func (s Store) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := s.db.ExecContext(ctx, updatePasswordQuery, userID, passwordHash)
	return err
}
