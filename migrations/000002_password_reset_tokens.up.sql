-- One-time password-reset tokens. Same shape as refresh_tokens: only the
-- SHA-256 hash is stored, used_at enforces single use, expiry is checked
-- by the service. Deleting the user cascades their tokens.
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    used_at TIMESTAMPTZ
);

CREATE INDEX password_reset_tokens_user_idx ON password_reset_tokens (user_id);
