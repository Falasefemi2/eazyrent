-- Core EasyRent schema ported from the Drizzle definitions.
-- Notes:
--   * users.verification columns are snake_cased here (email_verified,
--     verification_token, verification_token_expires_at). The Drizzle source
--     used camelCase quoted identifiers ("emailVerified", "verificationToken",
--     "verificationTokenExpiry").
--   * status enum values keep the Drizzle spelling verbatim
--     ('avaiable', 'rented', 'inative') so existing code stays compatible.
--     Consider renaming to ('available', 'rented', 'inactive') later.
--   * listings.location needs PostGIS for geography(POINT,4326).

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TYPE status AS ENUM ('avaiable', 'rented', 'inative');
CREATE TYPE media_type AS ENUM ('image', 'video');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    phone TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verification_token TEXT,
    verification_token_expires_at TIMESTAMPTZ
);

CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    landlord_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    rooms INTEGER DEFAULT 0,
    furnished BOOLEAN NOT NULL DEFAULT FALSE,
    status status DEFAULT 'avaiable',
    location geography(POINT, 4326) NOT NULL,
    address VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX listings_status_idx ON listings (status);
CREATE INDEX listings_price_idx ON listings (price);
CREATE INDEX listings_landlord_idx ON listings (landlord_id);
CREATE INDEX listings_created_at_idx ON listings (created_at);
CREATE INDEX listings_location_idx ON listings USING GIST (location);

CREATE TABLE listing_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID NOT NULL REFERENCES listings (id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    type media_type NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX listing_media_listing_idx ON listing_media (listing_id);

CREATE TABLE favorites (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    listing_id UUID NOT NULL REFERENCES listings (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, listing_id)
);

CREATE INDEX favorites_listing_idx ON favorites (listing_id);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX refresh_tokens_user_idx ON refresh_tokens (user_id);
