package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/femi/golang-easyrent/internal/auth"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func openAuthTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping DB test")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}
	return db
}

func TestRequireAuth(t *testing.T) {
	db := openAuthTestDB(t)
	store := auth.NewStore(db)
	tokens := auth.Tokens{Secret: []byte("web-test-secret-please-ignore-01"), AccessTTL: time.Hour}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := fmt.Sprintf("middleware-test-%d@example.com", time.Now().UnixNano())
	phone := fmt.Sprintf("+234%013d", time.Now().UnixNano()%10000000000000)
	u, err := store.CreateUser(ctx, email, phone, "hash-placeholder", "Middleware User")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, u.ID)
	})

	raw, err := tokens.SignAccessToken(u.ID, u.Email)
	if err != nil {
		t.Fatalf("SignAccessToken: %v", err)
	}

	a := Auth{Users: store, Tokens: tokens}
	protected := a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		cu, ok := CurrentUserOf(r)
		if !ok {
			writeError(w, http.StatusInternalServerError, "no current user")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"user": cu.UserID.String(), "email": cu.Email})
	})

	call := func(header string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		return rec
	}

	if rec := call(""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing header: want 401, got %d", rec.Code)
	}
	if rec := call("Token " + raw); rec.Code != http.StatusUnauthorized {
		t.Fatalf("non-bearer scheme: want 401, got %d", rec.Code)
	}
	if rec := call("Bearer bogus"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad token: want 401, got %d", rec.Code)
	}

	rec := call("Bearer " + raw)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid token: want 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["user"] != u.ID.String() || body["email"] != email {
		t.Fatalf("wrong current user: %v", body)
	}

	expiredTokens := auth.Tokens{Secret: []byte("web-test-secret-please-ignore-01"), AccessTTL: -time.Second}
	expired, err := expiredTokens.SignAccessToken(u.ID, u.Email)
	if err != nil {
		t.Fatalf("SignAccessToken expired: %v", err)
	}
	if rec := call("Bearer " + expired); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expired token: want 401, got %d", rec.Code)
	}

	// Token for a deleted user must stop working.
	if _, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, u.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if rec := call("Bearer " + raw); rec.Code != http.StatusUnauthorized {
		t.Fatalf("deleted user: want 401, got %d", rec.Code)
	}
}
