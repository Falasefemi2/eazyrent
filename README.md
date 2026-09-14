# EazyRent (Go rewrite)

Go rewrite of [EasyRent](https://github.com/Falasefemi2/easyrent) — a house rental platform backend for
the Nigerian market. The original is TypeScript (Bun + Effect-TS + Drizzle);
this port keeps the same Postgres schema and auth behavior in boring,
idiomatic Go: stdlib HTTP, plain SQL, concrete types, no frameworks.

## Stack

- Go 1.27, stdlib `net/http` mux (no web framework)
- PostgreSQL + PostGIS (`database/sql` over pgx stdlib driver)
- Plain SQL migrations in `migrations/` (applied by `cmd/migrate`)
- JWT access tokens (`golang-jwt`), argon2id passwords, SHA-256 refresh tokens
- Swagger UI via swaggo annotations (`docs/` is generated — see `task docs`)
- `task` (go-task) as the runner; `.env` is loaded automatically

## What works so far

Auth (ported from the TS `AuthService` + `TokenService` + `PasswordService`):

- `POST /auth/signup` → 201 + token pair (also stores an email-verification token)
- `POST /auth/signin` → 200 (401 bad credentials, 403 unverified)
- `POST /auth/refresh` → 200 rotated pair (reuse = theft → all sessions revoked)
- `POST /auth/signout` → 204, idempotent
- `GET /auth/verify?token=` → email verification (single-use, 24h)
- `POST /auth/forgot-password` → 202, silent for unknown emails
- `POST /auth/reset-password` → 200, single-use 1h token, kills all sessions
- `GET /me` → current user (bearer)
- `PUT /me/avatar` → store/clear avatar URL from the upload provider
- `GET /healthz`, Swagger UI at `GET /swagger/`

Favorites:

- `POST /favorites/{id}` → favorite a listing (idempotent, 404 if listing missing)
- `DELETE /favorites/{id}` → unfavorite (idempotent, 204)
- `GET /favorites?page=&limit=` → caller's favorites, newest first, with covers + favorite counts

Still to port: geospatial radius search.

Rate limits are in-memory (no Redis): list 30/min and detail 60/min per
IP, create 10/hour per user — 429 with `Retry-After` when exceeded.

## Quickstart

Requirements: Go 1.27+, Postgres with the PostGIS extension, `task` on PATH.

```powershell
# 1. env
Copy-Item .env.example .env   # then set DATABASE_URL + ACCESS_TOKEN_SECRET

# 2. migrate + run
task dev

# 3. test the API
# Swagger: http://localhost:8080/swagger/index.html
```

Emails are logged to stdout by the dummy sender (grab verification/reset
tokens from the console). Wiring Resend: [`docs/resend-email.md`](docs/resend-email.md).
Avatar uploads (Cloudinary, no backend keys): [`docs/avatar-uploads.md`](docs/avatar-uploads.md).

## Commands

```powershell
task dev          # migrate up, then run the API
task run          # run the API (DATABASE_URL required)
task migrate      # apply pending SQL migrations
task migrate-down # roll back the latest migration
task test         # go test ./...  (.env loaded, DB tests run)
task check        # gofmt + vet + tests
task docs         # regenerate Swagger docs (needs swag CLI)
```

## Layout

```text
cmd/api        composition root: config → db → services → handler → HTTP
cmd/migrate    minimal SQL migration runner (schema_migrations table)
internal/auth  users + refresh/reset tokens: Store (SQL), Service (flows),
               Tokens (JWT), passwords (argon2id), dummy EmailSender
internal/config  env parsing + validation (fails fast, no silent defaults)
internal/db    *sql.DB wiring over pgx stdlib
internal/favorite  favorites join table: Store (SQL), Service (idempotent add/remove)
internal/web   HTTP boundary: decode + validate once, call services/stores,
                RequireAuth bearer middleware, Swagger mount
migrations     000001 core tables, 000002 password_reset_tokens
docs           resend-email.md, avatar-uploads.md (swagger.* is generated)
```

Conventions: validate untrusted input at the boundary, concrete types
everywhere else; sentinel errors mapped to HTTP statuses in one place
(`authErrorStatus`); only refresh-token hashes touch the database, never
raw tokens.
