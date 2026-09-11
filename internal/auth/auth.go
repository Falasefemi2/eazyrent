package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

type User struct {
	ID                         uuid.UUID      `json:"id"`
	Email                      string         `json:"email"`
	Phone                      string         `json:"phone"`
	PasswordHash               string         `json:"-"`
	FullName                   string         `json:"full_name"`
	AvatarURL                  sql.NullString `json:"avatar_url"`
	EmailVerified              bool           `json:"email_verified"`
	VerificationToken          sql.NullString `json:"verification_token"`
	VerificationTokenExpiresAt sql.NullTime   `json:"verification_token_expires_at"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID    `json:"id"`
	UserID    uuid.UUID    `json:"user_id"`
	ExpiresAt time.Time    `json:"expires_at"`
	RevokedAt sql.NullTime `json:"revoked_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

const userColumns = `id, email, phone, password_hash, full_name, avatar_url,
	email_verified, verification_token, verification_token_expires_at,
	created_at, updated_at`

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (User, error) {
	var u User
	err := s.Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&u.FullName,
		&u.AvatarURL,
		&u.EmailVerified,
		&u.VerificationToken,
		&u.VerificationTokenExpiresAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return u, err
}

const findUserByEmailQuery = `
	SELECT ` + userColumns + `
	FROM users
	WHERE email = $1
	LIMIT 1
`

func (s Store) FindByEmail(ctx context.Context, email string) (User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, findUserByEmailQuery, email))
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

const findUserByIDQuery = `
	SELECT ` + userColumns + `
	FROM users
	WHERE id = $1
	LIMIT 1
`

func (s Store) FindByID(ctx context.Context, id uuid.UUID) (User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, findUserByIDQuery, id))
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

const createUserQuery = `
	INSERT INTO users (email, phone, password_hash, full_name)
	VALUES ($1, $2, $3, $4)
	RETURNING ` + userColumns + `
`

const updateAvatarQuery = `
	UPDATE users
	SET avatar_url = $2, updated_at = now()
	WHERE id = $1
`

func (s Store) UpdateAvatar(ctx context.Context, id uuid.UUID, url string) error {
	var avatar sql.NullString
	if url != "" {
		avatar = sql.NullString{String: url, Valid: true}
	}
	_, err := s.db.ExecContext(ctx, updateAvatarQuery, id, avatar)
	return err
}

func (s Store) CreateUser(ctx context.Context, email, phone, passwordHash, fullName string) (User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, createUserQuery, email, phone, passwordHash, fullName))
	if err != nil {
		return User{}, err
	}
	return u, nil
}

const storeRefreshTokenQuery = `
	INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)
`

func (s Store) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, storeRefreshTokenQuery, userID, tokenHash, expiresAt)
	return err
}

const findRefreshTokenQuery = `
	SELECT id, user_id, expires_at, revoked_at
	FROM refresh_tokens
	WHERE token_hash = $1
	LIMIT 1
`

func (s Store) FindRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, error) {
	var t RefreshToken
	err := s.db.QueryRowContext(ctx, findRefreshTokenQuery, tokenHash).
		Scan(&t.ID, &t.UserID, &t.ExpiresAt, &t.RevokedAt)
	if err == sql.ErrNoRows {
		return RefreshToken{}, ErrRefreshTokenNotFound
	}
	if err != nil {
		return RefreshToken{}, err
	}
	return t, nil
}

const revokeRefreshTokenQuery = `
	UPDATE refresh_tokens
	SET revoked_at = now()
	WHERE id = $1
`

func (s Store) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, revokeRefreshTokenQuery, id)
	return err
}

const revokeAllUserTokensQuery = `
	UPDATE refresh_tokens
	SET revoked_at = now()
	WHERE user_id = $1 AND revoked_at IS NULL
`

func (s Store) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, revokeAllUserTokensQuery, userID)
	return err
}

const storeVerificationTokenQuery = `
	UPDATE users
	SET verification_token = $2, verification_token_expires_at = $3
	WHERE id = $1
`

func (s Store) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, storeVerificationTokenQuery, userID, token, expiresAt)
	return err
}

const findUserByVerificationTokenQuery = `
	SELECT ` + userColumns + `
	FROM users
	WHERE verification_token = $1
	LIMIT 1
`

func (s Store) FindByVerificationToken(ctx context.Context, token string) (User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, findUserByVerificationTokenQuery, token))
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

const markEmailVerifiedQuery = `
	UPDATE users
	SET email_verified = TRUE, verification_token = NULL, verification_token_expires_at = NULL
	WHERE id = $1
`

func (s Store) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, markEmailVerifiedQuery, userID)
	return err
}
