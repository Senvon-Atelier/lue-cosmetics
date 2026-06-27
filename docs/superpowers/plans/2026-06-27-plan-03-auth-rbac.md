# Auth + RBAC Implementation Plan (Plan 3 of 15)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land hand-rolled Go authentication (email/password + Google OAuth), DB-backed sessions, an `auth` package with service + handlers + middleware, the first auth-gated endpoint `GET /api/v1/me`, and the integration-test RBAC matrix the spec mandates. After this plan, every future endpoint can sit behind `RequireSession` or `RequireRole("admin")` and trust the context.

**Architecture (per the updated spec):**
- Identity tables (`users`, `password_credentials`, `oauth_accounts`, `sessions`, `verification_tokens`, `user_roles`) live in a single migration.
- `internal/auth/` package: `password` (argon2id), `token` (random + hash), `repository` (sqlc wrapper), `service` (Signup/Login/Logout/GetSession/etc.), `handler` (HTTP), `google` (OAuth flow), `middleware` (`RequireSession`, `RequireRole`, `MustBeAdmin`).
- Session token: 32 random bytes from `crypto/rand`, base64-url encoded, lives **only** in the cookie. DB stores `sha256(token)`.
- Email sending is **stubbed** in this plan — a `email.Sender` interface with a `LogSender` implementation that writes a slog line. Real Resend integration arrives in Plan 5. Per-spec allowlist behavior is implemented now (signup of a non-allowlisted address marks the user verified server-side and skips the verify token).
- Two new env vars: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`. Empty in dev (the `/auth/google/*` handlers respond 503 `not_configured` if either is missing — server still boots).
- Three more new env vars: `SESSION_COOKIE_NAME` (default `rue_session`), `SESSION_COOKIE_DOMAIN` (default empty = host-only), `EMAIL_ALLOWLIST` (comma-separated; default empty).

**Tech Stack additions:**
- `golang.org/x/crypto/argon2` — password hashing
- `golang.org/x/oauth2` + `golang.org/x/oauth2/google` — OAuth client
- `google.golang.org/api/idtoken` — Google ID token verification
- Postgres `citext` extension — case-insensitive email column
- Postgres `inet` type — `sessions.ip`

## Global Constraints

- **Module path:** `github.com/oti-adjei/ruecosmetics`.
- **Working directory:** `/Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/`. Backend paths relative to `backend/`.
- **Argon2id params (production):** memory `64*1024` KiB (64 MiB), time `3`, parallelism `4`, salt 16 B, key 32 B — RFC 9106 first-class params. PHC-formatted string output.
- **Argon2id params (test):** memory `8*1024` KiB (8 MiB), time `1`, parallelism `1`. The `password.Params` struct is plumbed through `auth.Service` so tests can override; never expose this knob via env (production is locked).
- **Session token format:** 32 random bytes → `base64.URLEncoding.EncodeToString(b)` (43 chars, no padding). Cookie name from `Config.SessionCookieName` (default `rue_session`). Stored as `sha256(decodedBytes)` in `sessions.token_hash` (`bytea`).
- **Other token format (verify / reset):** same generation; same hash-on-store rule; surfaced to the user via URL query in the (stubbed) email.
- **Cookie attributes:** HttpOnly, SameSite=Lax, Path=`/`. `Secure` is on iff `Config.Env != "development"`. `Domain` from `Config.SessionCookieDomain` (empty = host-only — fine for prod where frontend and API share eTLD+1).
- **Session lifetime:** 30 days. `last_used_at` refreshed on every authenticated request. `expires_at` rolled forward iff the existing one is more than 24h in the past (cheap heuristic; avoids a write on every request).
- **Login response:** 401 with envelope code `unauthorized` on any failure (wrong email OR wrong password OR unverified email when verification is required). No discriminating between the cases in the response body — the slog line on the server side can be specific.
- **Password reset request:** always returns 204. No enumeration through the response.
- **Roles:** `customer` (default for self-signup) and `admin` (assigned via seed only in v1).
- **Email allowlist:** if `Config.EmailAllowlist` contains the signup address (or is `"*"`), the address is "allowlisted" → real verification flow. Otherwise → server-side auto-verify at signup; no email queued. The `email.Sender` interface still gets called; the `LogSender` only writes a slog line.
- **OAuth state cookie:** name `rue_oauth_state`, HttpOnly, SameSite=Lax, Secure same as session, 10-minute Max-Age.
- **Error codes used:** `unauthorized` (401), `forbidden` (403), `validation_failed` (400), `not_found` (404), `conflict` (409 — email already in use), `internal_error` (500), `not_configured` (503 — Google OAuth not set up). The `not_configured` code is NEW; add it to `httpx/error.go`.
- **Commit identity:** `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit ...` on every commit.
- **Bundled commits (per user preference):** 15 tasks land as **6 commits**:
  1. Tasks 1-4 — `feat(auth): identity schema, sqlc queries, argon2id + token primitives`
  2. Tasks 5-6 — `feat(auth): service, signup/login/logout/session handlers`
  3. Tasks 7-9 — `feat(auth): RequireSession + RequireRole middleware, GET /me`
  4. Tasks 10-11 — `feat(auth): Google OAuth start/callback`
  5. Tasks 12-13 — `feat(auth): email verify + password reset (handlers, email stubbed)`
  6. Tasks 14-15 — `feat(auth): RBAC integration matrix, regenerate openapi`
- **HEAD before Plan 3 begins:** `8136426` (spec pivot commit).

## File Structure

```
casestud/ruecosmetics/backend/
├── internal/
│   ├── auth/                                  # NEW
│   │   ├── password.go                        # argon2id Hash + Verify + Params
│   │   ├── password_test.go
│   │   ├── token.go                           # NewToken + HashToken + ConstantTimeEqual
│   │   ├── token_test.go
│   │   ├── repository.go                      # sqlc wrappers
│   │   ├── service.go                         # Signup/Login/Logout/GetSession/Verify/Reset
│   │   ├── service_test.go
│   │   ├── handler.go                         # HTTP handlers + swag annotations
│   │   ├── handler_test.go
│   │   ├── middleware.go                      # RequireSession, RequireRole, MustBeAdmin, ctx keys
│   │   ├── middleware_test.go
│   │   └── google.go                          # Google OAuth start + callback + state cookie
│   ├── email/                                 # NEW (minimal — Plan 5 expands)
│   │   ├── sender.go                          # Sender interface + LogSender
│   │   └── sender_test.go
│   ├── me/                                    # NEW (first auth-gated endpoint)
│   │   ├── handler.go                         # GET /me
│   │   └── handler_test.go
│   ├── config/
│   │   ├── config.go                          # MODIFY — add session/oauth/email envs
│   │   └── config_test.go                     # MODIFY — assert new fields
│   ├── app/
│   │   └── app.go                             # MODIFY — wire auth.Service + email.Sender
│   ├── httpx/
│   │   └── error.go                           # MODIFY — add CodeUnauthorized="unauthorized",
│   │                                          #          CodeForbidden="forbidden",
│   │                                          #          CodeConflict="conflict",
│   │                                          #          CodeNotConfigured="not_configured"
│   │                                          # (Some already exist — additive.)
│   └── testsupport/
│       ├── postgres.go                        # MODIFY — add `WriteShippingConfig(t)` helper if not lifted yet
│       └── auth.go                            # NEW — `LoginAs(t, app, role)` helper for handler tests
├── cmd/api/
│   └── main.go                                # MODIFY — mount auth handlers + /me under /api/v1
├── migrations/
│   └── 00003_auth.sql                         # NEW — all six identity tables
├── queries/
│   └── auth.sql                               # NEW
└── docs/                                      # REGENERATED in Task 15
```

---

### Task 1: Migration 00003 — identity schema

**Files:**
- Create: `backend/migrations/00003_auth.sql`

**Interfaces:**
- Produces: `users`, `password_credentials`, `oauth_accounts`, `sessions`, `verification_tokens`, `user_roles` tables + `citext` extension.

- [ ] **Step 1: Write the migration**

File: `backend/migrations/00003_auth.sql`
```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email           citext NOT NULL UNIQUE,
    name            text NOT NULL DEFAULT '',
    image           text,
    email_verified  boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE password_credentials (
    user_id        uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash  text NOT NULL,
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE oauth_accounts (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider              text NOT NULL,
    provider_account_id   text NOT NULL,
    created_at            timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_account_id)
);
CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);

CREATE TABLE sessions (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash    bytea NOT NULL UNIQUE,
    expires_at    timestamptz NOT NULL,
    ip            inet,
    user_agent    text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    last_used_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE verification_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind        text NOT NULL CHECK (kind IN ('email_verify','password_reset')),
    token_hash  bytea NOT NULL UNIQUE,
    expires_at  timestamptz NOT NULL,
    used_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_verification_tokens_user_id ON verification_tokens(user_id);

CREATE TABLE user_roles (
    user_id  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role     text NOT NULL CHECK (role IN ('customer','admin')),
    PRIMARY KEY (user_id, role)
);

-- +goose Down
DROP TABLE user_roles;
DROP TABLE verification_tokens;
DROP TABLE sessions;
DROP TABLE oauth_accounts;
DROP TABLE password_credentials;
DROP TABLE users;
DROP EXTENSION IF EXISTS citext;
```

(No commit yet — bundled.)

---

### Task 2: sqlc queries — auth.sql

**Files:**
- Create: `backend/queries/auth.sql`
- Regenerate: `backend/internal/db/sqlc/`

**Interfaces produced (on `*sqlc.Queries`):**
- `CreateUser`, `GetUserByEmail`, `GetUserByID`, `UpdateUserEmailVerified`, `UpdateUserPasswordHash`
- `CreatePasswordCredential`, `GetPasswordCredentialByUserID`, `UpsertPasswordCredential`
- `UpsertOAuthAccount`, `GetOAuthAccount`
- `CreateSession`, `GetSessionByTokenHash`, `RefreshSession`, `DeleteSession`, `DeleteSessionsForUser`, `DeleteOtherSessionsForUser`
- `CreateVerificationToken`, `GetVerificationToken`, `MarkVerificationTokenUsed`
- `AddUserRole`, `ListUserRoles`

- [ ] **Step 1: Write the queries**

File: `backend/queries/auth.sql`
```sql
-- name: CreateUser :one
INSERT INTO users (email, name, email_verified)
VALUES ($1, $2, $3)
RETURNING id, email, name, image, email_verified, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, name, image, email_verified, created_at, updated_at
FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, name, image, email_verified, created_at, updated_at
FROM users WHERE id = $1;

-- name: UpdateUserEmailVerified :exec
UPDATE users SET email_verified = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserName :exec
UPDATE users SET name = $2, updated_at = now() WHERE id = $1;

-- name: UpsertPasswordCredential :exec
INSERT INTO password_credentials (user_id, password_hash, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (user_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash, updated_at = now();

-- name: GetPasswordCredentialByUserID :one
SELECT user_id, password_hash, updated_at
FROM password_credentials WHERE user_id = $1;

-- name: UpsertOAuthAccount :exec
INSERT INTO oauth_accounts (user_id, provider, provider_account_id)
VALUES ($1, $2, $3)
ON CONFLICT (provider, provider_account_id) DO NOTHING;

-- name: GetOAuthAccount :one
SELECT id, user_id, provider, provider_account_id, created_at
FROM oauth_accounts WHERE provider = $1 AND provider_account_id = $2;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at, ip, user_agent)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, token_hash, expires_at, ip, user_agent, created_at, last_used_at;

-- name: GetSessionByTokenHash :one
SELECT s.id, s.user_id, s.token_hash, s.expires_at, s.ip, s.user_agent, s.created_at, s.last_used_at
FROM sessions s
WHERE s.token_hash = $1 AND s.expires_at > now();

-- name: RefreshSessionLastUsed :exec
UPDATE sessions SET last_used_at = now() WHERE id = $1;

-- name: RollSessionExpiry :exec
UPDATE sessions SET expires_at = $2, last_used_at = now() WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = $1;

-- name: DeleteSessionsForUser :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteOtherSessionsForUser :exec
DELETE FROM sessions WHERE user_id = $1 AND id <> $2;

-- name: CreateVerificationToken :one
INSERT INTO verification_tokens (user_id, kind, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, kind, token_hash, expires_at, used_at, created_at;

-- name: GetUnusedVerificationToken :one
SELECT id, user_id, kind, token_hash, expires_at, used_at, created_at
FROM verification_tokens
WHERE token_hash = $1 AND kind = $2 AND used_at IS NULL AND expires_at > now();

-- name: MarkVerificationTokenUsed :exec
UPDATE verification_tokens SET used_at = now() WHERE id = $1;

-- name: AddUserRole :exec
INSERT INTO user_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: ListRolesForUser :many
SELECT role FROM user_roles WHERE user_id = $1 ORDER BY role;
```

- [ ] **Step 2: Regenerate**

```bash
cd /Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/backend
sqlc generate
```
Expected: `internal/db/sqlc/auth.sql.go` created with all the methods above. `models.go` gains `User`, `PasswordCredential`, `OauthAccount`, `Session`, `VerificationToken`, `UserRole` structs.

(No commit yet — bundled.)

---

### Task 3: Password hashing (argon2id) + tests

**Files:**
- Create: `backend/internal/auth/password.go`
- Create: `backend/internal/auth/password_test.go`

**Interfaces:**
- Produces:
  ```go
  type Params struct {
      Memory      uint32 // KiB
      Time        uint32
      Parallelism uint8
      SaltLength  uint32
      KeyLength   uint32
  }
  var DefaultParams = Params{Memory: 64*1024, Time: 3, Parallelism: 4, SaltLength: 16, KeyLength: 32}
  var TestParams    = Params{Memory: 8*1024,  Time: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}
  func Hash(password string, p Params) (string, error)        // returns PHC string
  func Verify(password, encoded string) (bool, error)         // constant-time compare
  ```
- Consumed by: Task 5 (service Signup, Login, ResetConfirm), Task 8 (test fixtures via TestParams).

- [ ] **Step 1: Add x/crypto dep**

```bash
cd /Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/backend
go get golang.org/x/crypto/argon2@latest
```

- [ ] **Step 2: Write failing tests**

File: `backend/internal/auth/password_test.go`
```go
package auth_test

import (
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
)

func TestHashFormat(t *testing.T) {
	h, err := auth.Hash("hunter2", auth.TestParams)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(h, "$argon2id$") {
		t.Errorf("prefix wrong: %s", h)
	}
	parts := strings.Split(h, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 segments, got %d in %s", len(parts), h)
	}
}

func TestVerifyRoundTrip(t *testing.T) {
	h, _ := auth.Hash("hunter2", auth.TestParams)
	ok, err := auth.Verify("hunter2", h)
	if err != nil || !ok {
		t.Errorf("Verify(correct) = %v, %v", ok, err)
	}
	ok, err = auth.Verify("wrong", h)
	if err != nil || ok {
		t.Errorf("Verify(wrong) = %v, %v", ok, err)
	}
}

func TestVerifyRejectsMalformed(t *testing.T) {
	if _, err := auth.Verify("x", "not-a-phc-string"); err == nil {
		t.Fatal("expected error on malformed hash")
	}
	if _, err := auth.Verify("x", "$argon2id$v=19$bad"); err == nil {
		t.Fatal("expected error on truncated hash")
	}
}

func TestHashIsRandomised(t *testing.T) {
	h1, _ := auth.Hash("hunter2", auth.TestParams)
	h2, _ := auth.Hash("hunter2", auth.TestParams)
	if h1 == h2 {
		t.Errorf("expected different salts → different hashes")
	}
}
```

- [ ] **Step 3: Run RED**

```bash
go test ./internal/auth/...
```
Expected: build error — package doesn't exist.

- [ ] **Step 4: Implement**

File: `backend/internal/auth/password.go`
```go
// Package auth owns identity: password hashing, session tokens, the auth
// service, HTTP handlers, RBAC middleware, and Google OAuth.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Memory      uint32
	Time        uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultParams = Params{Memory: 64 * 1024, Time: 3, Parallelism: 4, SaltLength: 16, KeyLength: 32}
var TestParams = Params{Memory: 8 * 1024, Time: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

func Hash(password string, p Params) (string, error) {
	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, p.KeyLength)
	b64salt := base64.RawStdEncoding.EncodeToString(salt)
	b64key := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Parallelism, b64salt, b64key), nil
}

func Verify(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("malformed argon2id hash")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, fmt.Errorf("unsupported argon2 version")
	}
	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Parallelism); err != nil {
		return false, fmt.Errorf("invalid params: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("salt b64: %w", err)
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("key b64: %w", err)
	}
	got := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
```

- [ ] **Step 5: Run GREEN**

```bash
go test ./internal/auth/... -v
```
Expected: all four password tests PASS. (Argon2 with TestParams is fast — under 50ms per hash.)

(No commit yet — bundled.)

---

### Task 4: Token primitives (random + hash + compare)

**Files:**
- Create: `backend/internal/auth/token.go`
- Create: `backend/internal/auth/token_test.go`

**Interfaces:**
- Produces:
  ```go
  func NewToken() (raw string, err error)     // 32 random bytes → base64.URLEncoding (no padding) → 43-char string
  func HashToken(raw string) [32]byte         // sha256 of decoded bytes
  func TokenEquals(raw string, stored [32]byte) bool   // constant-time
  ```
- Consumed by: session creation (Task 5), verify/reset token issuance (Task 12, 13).

- [ ] **Step 1: Write failing test**

File: `backend/internal/auth/token_test.go`
```go
package auth_test

import (
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
)

func TestNewTokenLength(t *testing.T) {
	tok, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if len(tok) != 43 {
		t.Errorf("token len = %d, want 43", len(tok))
	}
}

func TestNewTokenUnique(t *testing.T) {
	a, _ := auth.NewToken()
	b, _ := auth.NewToken()
	if a == b {
		t.Errorf("tokens collided: %s == %s", a, b)
	}
}

func TestHashAndEquals(t *testing.T) {
	tok, _ := auth.NewToken()
	h := auth.HashToken(tok)
	if !auth.TokenEquals(tok, h) {
		t.Errorf("TokenEquals same token = false")
	}
	other, _ := auth.NewToken()
	if auth.TokenEquals(other, h) {
		t.Errorf("TokenEquals different token = true")
	}
}

func TestHashRejectsMalformed(t *testing.T) {
	// Hash of an unparseable token should still produce SOME 32-byte value
	// (we hash the raw string when decoding fails — but TokenEquals against
	// a different stored hash must return false).
	tok := "not-a-real-token!!!"
	h := auth.HashToken(tok)
	if h == ([32]byte{}) {
		t.Errorf("expected non-zero hash")
	}
}
```

- [ ] **Step 2: RED**

```bash
go test ./internal/auth/... -run Token
```
Expected: symbols undefined.

- [ ] **Step 3: Implement**

File: `backend/internal/auth/token.go`
```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

const tokenBytes = 32

// NewToken returns a fresh URL-safe session/verification token: 32 random
// bytes encoded with base64.URLEncoding without padding (43 chars).
func NewToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns sha256 of the decoded bytes if the input is valid
// base64-url; otherwise sha256 of the raw input. Either way it is suitable
// for use as a DB lookup key — but the "raw-input" fallback path is unlikely
// to match anything we ourselves issued.
func HashToken(raw string) [32]byte {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return sha256.Sum256([]byte(raw))
	}
	return sha256.Sum256(b)
}

// TokenEquals compares a presented raw token against a stored hash in
// constant time.
func TokenEquals(raw string, stored [32]byte) bool {
	h := HashToken(raw)
	return subtle.ConstantTimeCompare(h[:], stored[:]) == 1
}
```

- [ ] **Step 4: GREEN**

```bash
go test ./internal/auth/... -v
```
Expected: password + token tests all PASS.

- [ ] **Step 5: Add the new error code constants to httpx**

Edit `backend/internal/httpx/error.go`. Add to the `const` block:
```go
CodeNotConfigured = "not_configured"
```
(The codes `CodeUnauthorized`, `CodeForbidden`, `CodeConflict` already exist from Plan 1; verify presence and add only what's missing.)

- [ ] **Step 6: Add new env vars to `Config`**

Edit `backend/internal/config/config.go`. Add fields:
```go
SessionCookieName     string   `envconfig:"SESSION_COOKIE_NAME" default:"rue_session"`
SessionCookieDomain   string   `envconfig:"SESSION_COOKIE_DOMAIN" default:""`
GoogleClientID        string   `envconfig:"GOOGLE_CLIENT_ID" default:""`
GoogleClientSecret    string   `envconfig:"GOOGLE_CLIENT_SECRET" default:""`
GoogleRedirectURL     string   `envconfig:"GOOGLE_REDIRECT_URL" default:"http://localhost:8080/api/v1/auth/google/callback"`
EmailAllowlist        []string `envconfig:"EMAIL_ALLOWLIST" default:""`
FrontendBaseURL       string   `envconfig:"FRONTEND_BASE_URL" default:"http://localhost:5173"`
```

Update `backend/.env.example` with all seven (commented unless they have meaningful dev defaults).

Extend `config_test.go` with a single test that sets all the new env vars and asserts they parse into the right fields.

(No commit yet — bundled with Tasks 1-3.)

- [ ] **Step 7: Commit Bundle 1 (Tasks 1-4)**

```bash
cd /Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics
git add -A
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -F /tmp/rue-p3-b1.txt
```

`/tmp/rue-p3-b1.txt`:
```
feat(auth): identity schema, sqlc queries, argon2id + token primitives

- migrations/00003: users, password_credentials, oauth_accounts, sessions,
  verification_tokens, user_roles with citext + bytea + check constraints
- queries/auth.sql: full CRUD for the identity domain
- internal/auth/password.go: argon2id Hash/Verify with PHC encoding
- internal/auth/token.go: NewToken/HashToken/TokenEquals
- internal/httpx/error.go: CodeNotConfigured
- internal/config: SESSION_*, GOOGLE_*, EMAIL_ALLOWLIST, FRONTEND_BASE_URL envs
```

---

### Task 5: `auth.Service` — Signup, Login, Logout, GetSession

**Files:**
- Create: `backend/internal/auth/repository.go`
- Create: `backend/internal/auth/service.go`
- Create: `backend/internal/auth/service_test.go`
- Create: `backend/internal/email/sender.go`
- Create: `backend/internal/email/sender_test.go`

**Interfaces:**
- Produces (in `internal/email`):
  ```go
  type Sender interface {
      Send(ctx context.Context, to, template string, data map[string]any) error
  }
  type LogSender struct { Log *slog.Logger }   // satisfies Sender; logs payload only
  ```
- Produces (in `internal/auth`):
  ```go
  type Repository struct { /* *sqlc.Queries + db.Pool for tx */ }
  func NewRepository(pool db.Pool) *Repository

  type Service struct {
      Repo            *Repository
      Pool            db.Pool
      Email           email.Sender
      Log             *slog.Logger
      Params          password.Params
      Allowlist       []string         // exact-match (case-insensitive); empty = no addresses allowlisted; "*" allowlists everything
      SessionLifetime time.Duration    // default 30 * 24h
      Now             func() time.Time // injectable clock
  }
  func NewService(repo *Repository, pool db.Pool, log *slog.Logger, sender email.Sender, allowlist []string) *Service

  type SignupInput struct { Email, Password, Name string }
  type SignupResult struct { UserID uuid.UUID; SessionToken string; SessionExpires time.Time; EmailVerified bool }
  func (s *Service) Signup(ctx, in SignupInput, ip net.IP, ua string) (SignupResult, error)

  type LoginInput struct { Email, Password string }
  type LoginResult struct { UserID uuid.UUID; Role string; SessionToken string; SessionExpires time.Time }
  func (s *Service) Login(ctx, in LoginInput, ip net.IP, ua string) (LoginResult, error)

  func (s *Service) Logout(ctx, rawToken string) error

  type SessionView struct { UserID uuid.UUID; Email, Name, Role string; EmailVerified bool }
  func (s *Service) GetSession(ctx, rawToken string) (SessionView, error)
  ```
- Sentinels:
  ```go
  var (
      ErrEmailInUse     = errors.New("email already in use")
      ErrInvalidCreds   = errors.New("invalid credentials")
      ErrNoSession      = errors.New("no session")
      ErrInvalidToken   = errors.New("invalid token")
  )
  ```

- [ ] **Step 1: Implement `internal/email/sender.go`**

```go
// Package email exposes the Sender interface that auth + orders flow through.
// Plan 5 will add a Resend-backed implementation; this plan ships only the
// LogSender, which writes the payload to slog without delivering anything.
package email

import (
	"context"
	"log/slog"
)

type Sender interface {
	Send(ctx context.Context, to, template string, data map[string]any) error
}

type LogSender struct {
	Log *slog.Logger
}

func (s LogSender) Send(ctx context.Context, to, template string, data map[string]any) error {
	s.Log.InfoContext(ctx, "email (stubbed)", "to", to, "template", template, "data", data)
	return nil
}
```

File: `backend/internal/email/sender_test.go`
```go
package email_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/email"
)

func TestLogSenderReturnsNil(t *testing.T) {
	s := email.LogSender{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := s.Send(context.Background(), "x@y.test", "welcome", nil); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Implement `internal/auth/repository.go`**

```go
package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

type Repository struct {
	q    *sqlcq.Queries
	pool db.Pool
}

func NewRepository(pool db.Pool) *Repository {
	return &Repository{q: sqlcq.New(pool), pool: pool}
}

// Pool exposes the *pgxpool.Pool so the service can run transactions.
func (r *Repository) Pool() db.Pool { return r.pool }

// withQueries returns a Queries bound to a transaction.
func (r *Repository) withQueries(tx pgx.Tx) *sqlcq.Queries { return sqlcq.New(tx) }

// ErrNotFound is returned by Get*-style methods when no row matches.
var ErrNotFound = errors.New("not found")

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (sqlcq.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.User{}, ErrNotFound
	}
	return u, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (sqlcq.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.User{}, ErrNotFound
	}
	return u, err
}

func (r *Repository) GetPasswordCredential(ctx context.Context, userID uuid.UUID) (sqlcq.PasswordCredential, error) {
	c, err := r.q.GetPasswordCredentialByUserID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.PasswordCredential{}, ErrNotFound
	}
	return c, err
}

func (r *Repository) GetSessionByTokenHash(ctx context.Context, hash []byte) (sqlcq.Session, error) {
	s, err := r.q.GetSessionByTokenHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Session{}, ErrNotFound
	}
	return s, err
}

func (r *Repository) ListRolesForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return r.q.ListRolesForUser(ctx, userID)
}

func (r *Repository) DeleteSession(ctx context.Context, hash []byte) error {
	return r.q.DeleteSession(ctx, hash)
}

func (r *Repository) RefreshSessionLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.RefreshSessionLastUsed(ctx, id)
}

func (r *Repository) RollSessionExpiry(ctx context.Context, id uuid.UUID, expires time.Time) error {
	return r.q.RollSessionExpiry(ctx, sqlcq.RollSessionExpiryParams{ID: id, ExpiresAt: expires})
}

func (r *Repository) GetUnusedVerificationToken(ctx context.Context, hash []byte, kind string) (sqlcq.VerificationToken, error) {
	v, err := r.q.GetUnusedVerificationToken(ctx, sqlcq.GetUnusedVerificationTokenParams{TokenHash: hash, Kind: kind})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.VerificationToken{}, ErrNotFound
	}
	return v, err
}

func (r *Repository) MarkVerificationTokenUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkVerificationTokenUsed(ctx, id)
}

func (r *Repository) UpsertPasswordCredential(ctx context.Context, userID uuid.UUID, hash string) error {
	return r.q.UpsertPasswordCredential(ctx, sqlcq.UpsertPasswordCredentialParams{UserID: userID, PasswordHash: hash})
}

func (r *Repository) DeleteOtherSessionsForUser(ctx context.Context, userID, keep uuid.UUID) error {
	return r.q.DeleteOtherSessionsForUser(ctx, sqlcq.DeleteOtherSessionsForUserParams{UserID: userID, ID: keep})
}

func (r *Repository) UpsertOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerAccountID string) error {
	return r.q.UpsertOAuthAccount(ctx, sqlcq.UpsertOAuthAccountParams{
		UserID: userID, Provider: provider, ProviderAccountID: providerAccountID,
	})
}

func (r *Repository) GetOAuthAccount(ctx context.Context, provider, providerAccountID string) (sqlcq.OauthAccount, error) {
	a, err := r.q.GetOAuthAccount(ctx, sqlcq.GetOAuthAccountParams{Provider: provider, ProviderAccountID: providerAccountID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.OauthAccount{}, ErrNotFound
	}
	return a, err
}
```

Add `import "time"` to repository.go.

> Implementer note: the exact sqlc parameter struct names depend on `sqlc generate` output — adjust to match. The semantics above are the contract.

- [ ] **Step 3: Implement `internal/auth/service.go`**

This is the heart of Plan 3. Below is the full file — paste verbatim then adjust sqlc parameter struct names if needed.

```go
package auth

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"log/slog"
)

var (
	ErrEmailInUse   = errors.New("auth: email already in use")
	ErrInvalidCreds = errors.New("auth: invalid credentials")
	ErrNoSession    = errors.New("auth: no session")
	ErrInvalidToken = errors.New("auth: invalid token")
)

const DefaultSessionLifetime = 30 * 24 * time.Hour
const sessionRollThreshold = 24 * time.Hour

type Service struct {
	Repo            *Repository
	Email           email.Sender
	Log             *slog.Logger
	Params          Params
	Allowlist       []string
	SessionLifetime time.Duration
	Now             func() time.Time
}

func NewService(repo *Repository, log *slog.Logger, sender email.Sender, allowlist []string) *Service {
	return &Service{
		Repo: repo, Log: log, Email: sender,
		Params: DefaultParams, Allowlist: normalizeAllowlist(allowlist),
		SessionLifetime: DefaultSessionLifetime,
		Now:             time.Now,
	}
}

func normalizeAllowlist(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func (s *Service) isAllowlisted(emailAddr string) bool {
	addr := strings.ToLower(strings.TrimSpace(emailAddr))
	for _, a := range s.Allowlist {
		if a == "*" || a == addr {
			return true
		}
	}
	return false
}

type SignupInput struct {
	Email    string
	Password string
	Name     string
}

type SignupResult struct {
	UserID         uuid.UUID
	SessionToken   string
	SessionExpires time.Time
	EmailVerified  bool
}

func (s *Service) Signup(ctx context.Context, in SignupInput, ip net.IP, ua string) (SignupResult, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if !validEmail(in.Email) || len(in.Password) < 8 {
		return SignupResult{}, ErrInvalidCreds
	}
	if _, err := s.Repo.GetUserByEmail(ctx, in.Email); err == nil {
		return SignupResult{}, ErrEmailInUse
	} else if !errors.Is(err, ErrNotFound) {
		return SignupResult{}, err
	}
	hash, err := Hash(in.Password, s.Params)
	if err != nil {
		return SignupResult{}, err
	}
	allow := s.isAllowlisted(in.Email)
	emailVerified := !allow // non-allowlisted → auto-verified at signup

	rawToken, err := NewToken()
	if err != nil {
		return SignupResult{}, err
	}
	tokenHash := HashToken(rawToken)
	expires := s.Now().Add(s.SessionLifetime)

	var result SignupResult
	err = db.WithTx(ctx, s.Repo.Pool(), func(tx db.Tx) error {
		q := sqlcq.New(tx)
		user, err := q.CreateUser(ctx, sqlcq.CreateUserParams{
			Email: in.Email, Name: in.Name, EmailVerified: emailVerified,
		})
		if err != nil {
			return err
		}
		if err := q.UpsertPasswordCredential(ctx, sqlcq.UpsertPasswordCredentialParams{
			UserID: user.ID, PasswordHash: hash,
		}); err != nil {
			return err
		}
		if err := q.AddUserRole(ctx, sqlcq.AddUserRoleParams{UserID: user.ID, Role: "customer"}); err != nil {
			return err
		}
		ipa := netToInet(ip)
		_, err = q.CreateSession(ctx, sqlcq.CreateSessionParams{
			UserID:    user.ID,
			TokenHash: tokenHash[:],
			ExpiresAt: expires,
			Ip:        ipa,
			UserAgent: ua,
		})
		if err != nil {
			return err
		}
		result = SignupResult{
			UserID: user.ID, SessionToken: rawToken,
			SessionExpires: expires, EmailVerified: emailVerified,
		}
		return nil
	})
	if err != nil {
		return SignupResult{}, err
	}

	// Email side effect AFTER tx commits.
	if allow {
		verifyRaw, _ := NewToken()
		verifyHash := HashToken(verifyRaw)
		_, _ = s.Repo.q.CreateVerificationToken(ctx, sqlcq.CreateVerificationTokenParams{
			UserID:    result.UserID,
			Kind:      "email_verify",
			TokenHash: verifyHash[:],
			ExpiresAt: s.Now().Add(24 * time.Hour),
		})
		_ = s.Email.Send(ctx, in.Email, "verify_email", map[string]any{
			"token": verifyRaw, "name": in.Name,
		})
	} else {
		_ = s.Email.Send(ctx, in.Email, "welcome", map[string]any{"name": in.Name})
	}
	return result, nil
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	UserID         uuid.UUID
	Role           string
	SessionToken   string
	SessionExpires time.Time
}

func (s *Service) Login(ctx context.Context, in LoginInput, ip net.IP, ua string) (LoginResult, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	user, err := s.Repo.GetUserByEmail(ctx, in.Email)
	if errors.Is(err, ErrNotFound) {
		// Burn argon2 anyway to keep timing constant.
		_, _ = Hash("decoy-password-for-constant-time", s.Params)
		return LoginResult{}, ErrInvalidCreds
	}
	if err != nil {
		return LoginResult{}, err
	}
	cred, err := s.Repo.GetPasswordCredential(ctx, user.ID)
	if errors.Is(err, ErrNotFound) {
		_, _ = Hash("decoy-password-for-constant-time", s.Params)
		return LoginResult{}, ErrInvalidCreds
	}
	if err != nil {
		return LoginResult{}, err
	}
	ok, err := Verify(in.Password, cred.PasswordHash)
	if err != nil || !ok {
		return LoginResult{}, ErrInvalidCreds
	}
	roles, err := s.Repo.ListRolesForUser(ctx, user.ID)
	if err != nil {
		return LoginResult{}, err
	}
	role := primaryRole(roles)

	rawToken, _ := NewToken()
	tokenHash := HashToken(rawToken)
	expires := s.Now().Add(s.SessionLifetime)
	_, err = s.Repo.q.CreateSession(ctx, sqlcq.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: tokenHash[:],
		ExpiresAt: expires,
		Ip:        netToInet(ip),
		UserAgent: ua,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		UserID: user.ID, Role: role,
		SessionToken: rawToken, SessionExpires: expires,
	}, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	h := HashToken(rawToken)
	return s.Repo.DeleteSession(ctx, h[:])
}

type SessionView struct {
	SessionID     uuid.UUID
	UserID        uuid.UUID
	Email, Name   string
	Role          string
	EmailVerified bool
}

func (s *Service) GetSession(ctx context.Context, rawToken string) (SessionView, error) {
	if rawToken == "" {
		return SessionView{}, ErrNoSession
	}
	h := HashToken(rawToken)
	sess, err := s.Repo.GetSessionByTokenHash(ctx, h[:])
	if errors.Is(err, ErrNotFound) {
		return SessionView{}, ErrNoSession
	}
	if err != nil {
		return SessionView{}, err
	}
	user, err := s.Repo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return SessionView{}, err
	}
	roles, err := s.Repo.ListRolesForUser(ctx, user.ID)
	if err != nil {
		return SessionView{}, err
	}
	// Best-effort touch; ignore errors (cookie still valid even if write fails).
	now := s.Now()
	_ = s.Repo.RefreshSessionLastUsed(ctx, sess.ID)
	if sess.ExpiresAt.Sub(now) < s.SessionLifetime-sessionRollThreshold {
		_ = s.Repo.RollSessionExpiry(ctx, sess.ID, now.Add(s.SessionLifetime))
	}
	return SessionView{
		SessionID: sess.ID, UserID: user.ID,
		Email: string(user.Email), Name: user.Name,
		Role: primaryRole(roles), EmailVerified: user.EmailVerified,
	}, nil
}

func primaryRole(roles []string) string {
	for _, r := range roles {
		if r == "admin" {
			return "admin"
		}
	}
	return "customer"
}

func validEmail(s string) bool {
	if len(s) < 3 || len(s) > 254 {
		return false
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	if strings.IndexByte(s[at+1:], '.') < 0 {
		return false
	}
	return true
}

func netToInet(ip net.IP) pgtype.Inet {
	if ip == nil {
		return pgtype.Inet{Valid: false}
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return pgtype.Inet{Valid: false}
	}
	return pgtype.Inet{IPNet: &net.IPNet{IP: addr.AsSlice(), Mask: fullMask(addr)}, Valid: true}
}

func fullMask(a netip.Addr) net.IPMask {
	if a.Is4() {
		return net.CIDRMask(32, 32)
	}
	return net.CIDRMask(128, 128)
}
```

> Implementer notes:
> - The `pgtype.Inet` adapter is brittle to pgx versions — if compilation fails, the simplest workaround is to drop `ip` to nullable in `CreateSession`'s call (pass `pgtype.Inet{Valid: false}`) and revisit IP capture in a later plan. Don't block on it.
> - sqlc parameter struct names assume the standard pluralization sqlc applies; verify post-generate.
> - `db.WithTx` already exists (Plan 1 Task 3). If `db.Tx` is not exported, switch to inline `pool.Begin/Commit` calls — the brief's intent is "atomic across CreateUser/UpsertPasswordCredential/AddUserRole/CreateSession," however you achieve it.

- [ ] **Step 4: Write service tests**

File: `backend/internal/auth/service_test.go`
```go
package auth_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func newService(t *testing.T) (*auth.Service, db.Pool, func()) {
	t.Helper()
	url, stop := testsupport.StartPostgres(t)
	testsupport.Migrate(t, url, "../../migrations")
	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		stop()
		t.Fatalf("pool: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := auth.NewRepository(pool)
	svc := auth.NewService(repo, logger, email.LogSender{Log: logger}, nil)
	svc.Params = auth.TestParams // fast hashes in tests
	return svc, pool, func() { pool.Close(); stop() }
}

func TestSignupCreatesUserSessionAndAutoVerifies(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	res, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "user@demo.test", Password: "hunter22", Name: "Ada",
	}, net.IPv4(127, 0, 0, 1), "go-test")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if !res.EmailVerified {
		t.Errorf("non-allowlisted address should be auto-verified")
	}
	if res.SessionToken == "" || res.SessionExpires.IsZero() {
		t.Errorf("session not minted: %+v", res)
	}
}

func TestSignupRejectsDuplicateEmail(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	_, err := svc.Signup(ctx, auth.SignupInput{Email: "dup@demo.test", Password: "12345678"}, nil, "")
	if err != nil {
		t.Fatalf("first signup: %v", err)
	}
	_, err = svc.Signup(ctx, auth.SignupInput{Email: "DUP@demo.test", Password: "12345678"}, nil, "")
	if !errors.Is(err, auth.ErrEmailInUse) {
		t.Errorf("dup signup err = %v, want ErrEmailInUse", err)
	}
}

func TestSignupRejectsShortPassword(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	_, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "x@y.test", Password: "short",
	}, nil, "")
	if !errors.Is(err, auth.ErrInvalidCreds) {
		t.Errorf("got %v, want ErrInvalidCreds", err)
	}
}

func TestLoginHappyPath(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	_, err := svc.Signup(ctx, auth.SignupInput{Email: "ok@y.test", Password: "hunter22"}, nil, "")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	res, err := svc.Login(ctx, auth.LoginInput{Email: "ok@y.test", Password: "hunter22"}, nil, "")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.Role != "customer" {
		t.Errorf("role = %s", res.Role)
	}
}

func TestLoginWrongPasswordSameErrAsMissingUser(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	_, _ = svc.Signup(ctx, auth.SignupInput{Email: "u@y.test", Password: "hunter22"}, nil, "")
	_, e1 := svc.Login(ctx, auth.LoginInput{Email: "u@y.test", Password: "WRONG"}, nil, "")
	_, e2 := svc.Login(ctx, auth.LoginInput{Email: "nobody@y.test", Password: "anything"}, nil, "")
	if !errors.Is(e1, auth.ErrInvalidCreds) || !errors.Is(e2, auth.ErrInvalidCreds) {
		t.Errorf("both errors must be ErrInvalidCreds: e1=%v e2=%v", e1, e2)
	}
}

func TestGetSessionRoundTrip(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	sr, _ := svc.Signup(ctx, auth.SignupInput{Email: "s@y.test", Password: "hunter22", Name: "Ann"}, nil, "")
	view, err := svc.GetSession(ctx, sr.SessionToken)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if view.UserID != sr.UserID {
		t.Errorf("user mismatch")
	}
	if !strings.EqualFold(view.Email, "s@y.test") {
		t.Errorf("email = %s", view.Email)
	}
	if view.Role != "customer" {
		t.Errorf("role = %s", view.Role)
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	sr, _ := svc.Signup(ctx, auth.SignupInput{Email: "lo@y.test", Password: "hunter22"}, nil, "")
	if err := svc.Logout(ctx, sr.SessionToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	_, err := svc.GetSession(ctx, sr.SessionToken)
	if !errors.Is(err, auth.ErrNoSession) {
		t.Errorf("after logout: want ErrNoSession, got %v", err)
	}
}

func TestSignupAllowlistedSendsVerifyToken(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	svc.Allowlist = []string{"vip@y.test"}
	res, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "vip@y.test", Password: "hunter22",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if res.EmailVerified {
		t.Errorf("allowlisted address should NOT be auto-verified — verify token issued")
	}
}
```

- [ ] **Step 5: Run service tests**

```bash
cd backend
go test ./internal/auth/... ./internal/email/... -v -timeout=300s
```
Expected: all PASS.

(No commit yet — bundled with Task 6.)

---

### Task 6: Auth HTTP handlers (signup, login, logout, session)

**Files:**
- Create: `backend/internal/auth/handler.go`
- Create: `backend/internal/auth/handler_test.go`

**Interfaces:**
- Produces:
  ```go
  type Handlers struct {
      Svc            *Service
      CookieName     string
      CookieDomain   string
      Secure         bool
  }
  func NewHandlers(svc *Service, cookieName, cookieDomain string, secure bool) *Handlers
  func (h *Handlers) Mount(r chi.Router)   // mounts at /auth/*
  ```

- [ ] **Step 1: Implement handlers**

File: `backend/internal/auth/handler.go`
```go
package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

type Handlers struct {
	Svc          *Service
	CookieName   string
	CookieDomain string
	Secure       bool
}

func NewHandlers(svc *Service, cookieName, cookieDomain string, secure bool) *Handlers {
	if cookieName == "" {
		cookieName = "rue_session"
	}
	return &Handlers{Svc: svc, CookieName: cookieName, CookieDomain: cookieDomain, Secure: secure}
}

func (h *Handlers) Mount(r chi.Router) {
	r.Post("/auth/signup", h.signup)
	r.Post("/auth/login", h.login)
	r.Post("/auth/logout", h.logout)
	r.Get("/auth/session", h.session)
}

type signupBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// signup godoc
//
// @Summary  Sign up with email and password
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body signupBody true "Signup payload"
// @Success  201 {object} sessionResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  409 {object} httpx.ErrorEnvelope
// @Router   /auth/signup [post]
func (h *Handlers) signup(w http.ResponseWriter, r *http.Request) {
	var body signupBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil)
		return
	}
	res, err := h.Svc.Signup(r.Context(), SignupInput(body), clientIP(r), r.UserAgent())
	switch {
	case errors.Is(err, ErrEmailInUse):
		httpx.WriteError(w, http.StatusConflict, httpx.CodeConflict, "email already in use", nil)
		return
	case errors.Is(err, ErrInvalidCreds):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid email or password (min 8 chars)", nil)
		return
	case err != nil:
		h.Svc.Log.ErrorContext(r.Context(), "signup", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "signup failed", nil)
		return
	}
	h.setSessionCookie(w, res.SessionToken, res.SessionExpires)
	httpx.WriteJSON(w, http.StatusCreated, sessionResponse{
		UserID:        res.UserID.String(),
		Email:         strings.ToLower(body.Email),
		Role:          "customer",
		EmailVerified: res.EmailVerified,
	})
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// login godoc
//
// @Summary  Log in with email and password
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body loginBody true "Login payload"
// @Success  200 {object} sessionResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Router   /auth/login [post]
func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil)
		return
	}
	res, err := h.Svc.Login(r.Context(), LoginInput(body), clientIP(r), r.UserAgent())
	if errors.Is(err, ErrInvalidCreds) {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "invalid email or password", nil)
		return
	}
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "login", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "login failed", nil)
		return
	}
	h.setSessionCookie(w, res.SessionToken, res.SessionExpires)
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{
		UserID: res.UserID.String(), Email: strings.ToLower(body.Email),
		Role: res.Role,
	})
}

// logout godoc
//
// @Summary  Log out (clear session)
// @Tags     auth
// @Produce  json
// @Success  204
// @Router   /auth/logout [post]
func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(h.CookieName); err == nil {
		_ = h.Svc.Logout(r.Context(), c.Value)
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

type sessionResponse struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Name          string `json:"name,omitempty"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

// session godoc
//
// @Summary  Get current session
// @Tags     auth
// @Produce  json
// @Success  200 {object} sessionResponse
// @Success  204 "no active session"
// @Router   /auth/session [get]
func (h *Handlers) session(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(h.CookieName)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	view, err := h.Svc.GetSession(r.Context(), c.Value)
	if errors.Is(err, ErrNoSession) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "session check failed", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{
		UserID: view.UserID.String(), Email: view.Email, Name: view.Name,
		Role: view.Role, EmailVerified: view.EmailVerified,
	})
}

func (h *Handlers) setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.CookieDomain,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handlers) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clientIP(r *http.Request) net.IP {
	// X-Forwarded-For first hop, fallback to RemoteAddr.
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		if i := strings.IndexByte(x, ','); i > 0 {
			x = x[:i]
		}
		return net.ParseIP(strings.TrimSpace(x))
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}
```

Add imports: `"net"` (for clientIP — but `chi` already has it indirectly). Run `go vet`.

- [ ] **Step 2: Write handler tests**

File: `backend/internal/auth/handler_test.go` — exercise signup/login/logout/session via `chi` + `httptest`. Pattern follows `catalog/handler_test.go` (use `strings.Contains` not handrolled). Cover:
- POST /auth/signup → 201, sets cookie, body has `email_verified: true` (non-allowlisted).
- POST /auth/signup with existing email → 409.
- POST /auth/login (right) → 200, sets cookie.
- POST /auth/login (wrong password) → 401 with `unauthorized` code.
- POST /auth/logout with active cookie → 204, clears cookie.
- GET /auth/session with active cookie → 200, body matches.
- GET /auth/session with no cookie → 204.

Use `import "strings"` for body assertions. Skip helper duplication — the `newService` from `service_test.go` is in the same `auth_test` package; reuse it for the handler tests too (just wrap with `NewHandlers`).

- [ ] **Step 3: GREEN**

```bash
go test ./internal/auth/... -v -timeout=300s
```
Expected: all PASS.

- [ ] **Step 4: Commit Bundle 2 (Tasks 5-6)**

```bash
git add -A
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -F /tmp/rue-p3-b2.txt
```

`/tmp/rue-p3-b2.txt`:
```
feat(auth): service, signup/login/logout/session handlers

- internal/email/sender.go: Sender interface + LogSender stub
- internal/auth/repository.go: sqlc wrappers for the identity domain
- internal/auth/service.go: Signup/Login/Logout/GetSession with allowlist,
  transactional signup, constant-time login, session rolling
- internal/auth/handler.go: POST /auth/signup, /auth/login, /auth/logout,
  GET /auth/session with HttpOnly+SameSite=Lax cookies
- service_test.go + handler_test.go: integration coverage
```

---

### Task 7: `RequireSession` middleware

**Files:**
- Create: `backend/internal/auth/middleware.go`
- Create: `backend/internal/auth/middleware_test.go`

**Interfaces:**
- Produces:
  ```go
  type ctxKey int
  const (
      SessionKey ctxKey = iota + 1
      UserIDKey
      RoleKey
  )
  func (h *Handlers) RequireSession(next http.Handler) http.Handler
  func GetUserID(ctx context.Context) (uuid.UUID, bool)
  func GetRole(ctx context.Context) (string, bool)
  func GetSessionView(ctx context.Context) (SessionView, bool)
  ```

- [ ] **Step 1: Implement**

File: `backend/internal/auth/middleware.go`
```go
package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

type ctxKey int

const (
	sessionKey ctxKey = iota + 1
	userIDKey
	roleKey
)

func (h *Handlers) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(h.CookieName)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
			return
		}
		view, err := h.Svc.GetSession(r.Context(), c.Value)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, view)
		ctx = context.WithValue(ctx, userIDKey, view.UserID)
		ctx = context.WithValue(ctx, roleKey, view.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	return v, ok
}

func GetRole(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(roleKey).(string)
	return v, ok
}

func GetSessionView(ctx context.Context) (SessionView, bool) {
	v, ok := ctx.Value(sessionKey).(SessionView)
	return v, ok
}
```

- [ ] **Step 2: Tests** (basic — happy + missing cookie + bad token)

File: `backend/internal/auth/middleware_test.go` — test that `RequireSession` 401s when cookie missing, 401s when token invalid, and lets through + injects context when valid. Use the existing `newService` helper, wrap with `NewHandlers`, and apply the middleware to a stub handler that reads `GetUserID(r.Context())`.

(No commit yet — bundled with Tasks 8-9.)

---

### Task 8: `RequireRole` middleware + `MustBeAdmin`

**Files:**
- Modify: `backend/internal/auth/middleware.go`
- Modify: `backend/internal/auth/middleware_test.go`

- [ ] **Step 1: Append to `middleware.go`**

```go
func (h *Handlers) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, ok := GetRole(r.Context())
			if !ok || got != role {
				httpx.WriteError(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MustBeAdmin is the belt-and-suspenders check called inside admin handlers.
// If the caller is somehow not admin, it writes 403 and returns false.
func MustBeAdmin(w http.ResponseWriter, r *http.Request) bool {
	if role, ok := GetRole(r.Context()); !ok || role != "admin" {
		httpx.WriteError(w, http.StatusForbidden, httpx.CodeForbidden, "admin required", nil)
		return false
	}
	return true
}
```

- [ ] **Step 2: Tests** — 403 when role mismatches, pass-through when role matches.

(No commit yet — bundled with Task 9.)

---

### Task 9: `/me` endpoint

**Files:**
- Create: `backend/internal/me/handler.go`
- Create: `backend/internal/me/handler_test.go`

**Interfaces:**
- Produces:
  ```go
  type Handlers struct { /* nothing required — reads from context */ }
  func NewHandlers() *Handlers
  func (h *Handlers) Mount(r chi.Router)   // r already has RequireSession
  ```

- [ ] **Step 1: Implement**

File: `backend/internal/me/handler.go`
```go
package me

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

type Handlers struct{}

func NewHandlers() *Handlers { return &Handlers{} }

func (h *Handlers) Mount(r chi.Router) {
	r.Get("/me", h.get)
}

type meResponse struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

// get godoc
//
// @Summary  Get the current user
// @Tags     me
// @Produce  json
// @Success  200 {object} meResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Router   /me [get]
func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, meResponse{
		UserID: view.UserID.String(), Email: view.Email, Name: view.Name,
		Role: view.Role, EmailVerified: view.EmailVerified,
	})
}
```

- [ ] **Step 2: Mount in `main.go`**

In `backend/cmd/api/main.go`, inside the `r.Route("/api/v1", func(api chi.Router) { ... })` block:

```go
// Auth handlers (public)
authHandlers := auth.NewHandlers(authService, cfg.SessionCookieName, cfg.SessionCookieDomain, cfg.Env != "development")
authHandlers.Mount(api)

// Auth-gated routes
api.Group(func(r chi.Router) {
    r.Use(authHandlers.RequireSession)
    me.NewHandlers().Mount(r)
})
```

Construct `authService` in `app.New` (see Task 9 Step 3) and expose on `Application`.

Add imports: `"github.com/oti-adjei/ruecosmetics/internal/auth"`, `"github.com/oti-adjei/ruecosmetics/internal/me"`.

- [ ] **Step 3: Wire `auth.Service` into `app.Application`**

Modify `backend/internal/app/app.go`:
```go
type Application struct {
    Config   *config.Config
    Pool     db.Pool
    Logger   *slog.Logger
    Shipping *shipping.Service
    Auth     *auth.Service
    Email    email.Sender
}
```

In `New(ctx, cfg)`:
```go
sender := email.LogSender{Log: logger}
repo := auth.NewRepository(pool)
authSvc := auth.NewService(repo, logger, sender, cfg.EmailAllowlist)
```

Then `return &Application{...Auth: authSvc, Email: sender, ...}`.

- [ ] **Step 4: Tests**

File: `backend/internal/me/handler_test.go` — wire up `auth.NewHandlers` + `me.NewHandlers`, sign up via the service, hit `/me` with the cookie. Assert 200 + body. Hit again without cookie → 401.

- [ ] **Step 5: Commit Bundle 3 (Tasks 7-9)**

```bash
git add -A
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -F /tmp/rue-p3-b3.txt
```

`/tmp/rue-p3-b3.txt`:
```
feat(auth): RequireSession + RequireRole middleware, GET /me

- internal/auth/middleware.go: RequireSession (validates cookie, injects
  userID + role + SessionView into ctx), RequireRole, MustBeAdmin
- internal/me/handler.go: GET /me reads from auth context
- internal/app: wire auth.Service + email.Sender into Application
- cmd/api: mount auth handlers; group /me under RequireSession
```

---

### Task 10: Google OAuth — `/auth/google/start`

**Files:**
- Create: `backend/internal/auth/google.go`
- Modify: `backend/internal/auth/handler.go` to mount the Google routes inside `Mount(r)`.

**Interfaces:**
- Produces:
  ```go
  type GoogleConfig struct {
      ClientID, ClientSecret, RedirectURL string
      FrontendBaseURL                     string
  }
  // Set on Handlers via a Configure method or constructor extension.
  ```

- [ ] **Step 1: Add OAuth deps**

```bash
cd backend
go get golang.org/x/oauth2@latest
go get google.golang.org/api/idtoken@latest
```

- [ ] **Step 2: Implement start handler**

File: `backend/internal/auth/google.go`
```go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oauthStateCookie = "rue_oauth_state"
const oauthStateMaxAge = 10 * 60

func (h *Handlers) oauthConfig() (*oauth2.Config, bool) {
	if h.GoogleClientID == "" || h.GoogleClientSecret == "" {
		return nil, false
	}
	return &oauth2.Config{
		ClientID:     h.GoogleClientID,
		ClientSecret: h.GoogleClientSecret,
		RedirectURL:  h.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}, true
}

// googleStart godoc
//
// @Summary  Begin Google OAuth login
// @Tags     auth
// @Success  302 "redirect to Google"
// @Failure  503 {object} httpx.ErrorEnvelope
// @Router   /auth/google/start [get]
func (h *Handlers) googleStart(w http.ResponseWriter, r *http.Request) {
	cfg, ok := h.oauthConfig()
	if !ok {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeNotConfigured, "google oauth not configured", nil)
		return
	}
	stateBytes := make([]byte, 32)
	_, _ = rand.Read(stateBytes)
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   oauthStateMaxAge,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// next return; helper to redirect to the frontend after callback
func (h *Handlers) afterOAuthRedirect() string {
	if h.FrontendBaseURL == "" {
		return "/"
	}
	u, err := url.Parse(h.FrontendBaseURL)
	if err != nil {
		return "/"
	}
	u.Path = "/account"
	return u.String()
}

// stub helper just to silence "imported and not used" for time when only
// callback uses it.
var _ = time.Second
```

- [ ] **Step 3: Extend `Handlers` struct + `Mount`**

In `handler.go`, replace the `Handlers` struct definition with:
```go
type Handlers struct {
	Svc                *Service
	CookieName         string
	CookieDomain       string
	Secure             bool
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendBaseURL    string
}
```

Extend `NewHandlers` to accept a `GoogleConfig` (or extend signature with the 4 new strings — choose the cleaner). Update `cmd/api/main.go` callers accordingly.

In `Mount(r)`, add:
```go
r.Get("/auth/google/start", h.googleStart)
r.Get("/auth/google/callback", h.googleCallback) // implemented in Task 11
```

(No commit yet — bundled with Task 11.)

---

### Task 11: Google OAuth — `/auth/google/callback`

**Files:**
- Modify: `backend/internal/auth/google.go`
- Modify: `backend/internal/auth/service.go` — add `LoginWithGoogle(ctx, provider, providerAccountID, email, name string, ip, ua) (LoginResult, error)`
- Tests: extend `service_test.go`, `handler_test.go`.

**Behavior:** verify state cookie matches `state` query param; exchange code for ID token via `cfg.Exchange`; validate ID token via `idtoken.Validate(ctx, idToken, h.GoogleClientID)`; extract `sub` and `email`; call `Svc.LoginWithGoogle("google", sub, email, name)`; set session cookie; 302 to the frontend `/account` URL.

`LoginWithGoogle` logic:
1. `GetOAuthAccount("google", sub)` — if found, load user by `account.UserID`, list roles, mint session, return.
2. Else `GetUserByEmail(email)`:
   - Found → link: `UpsertOAuthAccount(user.id, "google", sub)`, mint session.
   - Not found → create user (set `email_verified=true` since Google verified it), insert oauth account, add `customer` role, mint session.
3. Everything wrapped in `db.WithTx`.

- [ ] **Step 1: Add `LoginWithGoogle` to `service.go`**

Append:
```go
func (s *Service) LoginWithGoogle(ctx context.Context, providerSub, emailAddr, name string, ip net.IP, ua string) (LoginResult, error) {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	// Fast path: existing oauth_account.
	if acc, err := s.Repo.GetOAuthAccount(ctx, "google", providerSub); err == nil {
		return s.mintSessionForUser(ctx, acc.UserID, ip, ua)
	} else if !errors.Is(err, ErrNotFound) {
		return LoginResult{}, err
	}
	// Slow path: link or create.
	var userID uuid.UUID
	err := db.WithTx(ctx, s.Repo.Pool(), func(tx db.Tx) error {
		q := sqlcq.New(tx)
		user, err := q.GetUserByEmail(ctx, emailAddr)
		if errors.Is(err, pgx.ErrNoRows) {
			user, err = q.CreateUser(ctx, sqlcq.CreateUserParams{
				Email: emailAddr, Name: name, EmailVerified: true,
			})
			if err != nil {
				return err
			}
			if err := q.AddUserRole(ctx, sqlcq.AddUserRoleParams{UserID: user.ID, Role: "customer"}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := q.UpsertOAuthAccount(ctx, sqlcq.UpsertOAuthAccountParams{
			UserID: user.ID, Provider: "google", ProviderAccountID: providerSub,
		}); err != nil {
			return err
		}
		userID = user.ID
		return nil
	})
	if err != nil {
		return LoginResult{}, err
	}
	return s.mintSessionForUser(ctx, userID, ip, ua)
}

func (s *Service) mintSessionForUser(ctx context.Context, userID uuid.UUID, ip net.IP, ua string) (LoginResult, error) {
	roles, err := s.Repo.ListRolesForUser(ctx, userID)
	if err != nil {
		return LoginResult{}, err
	}
	rawToken, _ := NewToken()
	tokenHash := HashToken(rawToken)
	expires := s.Now().Add(s.SessionLifetime)
	_, err = s.Repo.q.CreateSession(ctx, sqlcq.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash[:],
		ExpiresAt: expires,
		Ip:        netToInet(ip),
		UserAgent: ua,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{UserID: userID, Role: primaryRole(roles), SessionToken: rawToken, SessionExpires: expires}, nil
}
```

- [ ] **Step 2: Implement `googleCallback`**

Append to `google.go`:
```go
// idtoken.Validate makes a network call to fetch Google's keyset. For tests,
// we swap this via the IDTokenValidator interface on Handlers.
type IDTokenValidator interface {
	Validate(ctx context.Context, token, audience string) (*Payload, error)
}
type Payload struct {
	Subject string
	Email   string
	Name    string
}

type googleValidator struct{}

func (googleValidator) Validate(ctx context.Context, token, audience string) (*Payload, error) {
	p, err := idtoken.Validate(ctx, token, audience)
	if err != nil {
		return nil, err
	}
	return &Payload{
		Subject: p.Subject,
		Email:   asString(p.Claims["email"]),
		Name:    asString(p.Claims["name"]),
	}, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

// googleCallback godoc
//
// @Summary  Handle Google OAuth callback
// @Tags     auth
// @Success  302
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  503 {object} httpx.ErrorEnvelope
// @Router   /auth/google/callback [get]
func (h *Handlers) googleCallback(w http.ResponseWriter, r *http.Request) {
	cfg, ok := h.oauthConfig()
	if !ok {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeNotConfigured, "google oauth not configured", nil)
		return
	}
	q := r.URL.Query()
	state := q.Get("state")
	code := q.Get("code")
	if state == "" || code == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "missing state or code", nil)
		return
	}
	c, err := r.Cookie(oauthStateCookie)
	if err != nil || c.Value == "" || c.Value != state {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid oauth state", nil)
		return
	}
	// Clear state cookie regardless of success.
	http.SetCookie(w, &http.Cookie{Name: oauthStateCookie, Value: "", Path: "/", Domain: h.CookieDomain, MaxAge: -1, HttpOnly: true, Secure: h.Secure, SameSite: http.SameSiteLaxMode})

	tok, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "oauth exchange", "err", err)
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "code exchange failed", nil)
		return
	}
	idToken, _ := tok.Extra("id_token").(string)
	if idToken == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "missing id_token", nil)
		return
	}
	validator := h.Validator
	if validator == nil {
		validator = googleValidator{}
	}
	payload, err := validator.Validate(r.Context(), idToken, h.GoogleClientID)
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "id token", "err", err)
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid id_token", nil)
		return
	}
	res, err := h.Svc.LoginWithGoogle(r.Context(), payload.Subject, payload.Email, payload.Name, clientIP(r), r.UserAgent())
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "login with google", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "login failed", nil)
		return
	}
	h.setSessionCookie(w, res.SessionToken, res.SessionExpires)
	http.Redirect(w, r, h.afterOAuthRedirect(), http.StatusFound)
}
```

Add field to `Handlers`: `Validator IDTokenValidator`. Initialize to `nil` in `NewHandlers`; tests can set a stub.

Imports for `google.go`: `"google.golang.org/api/idtoken"`.

- [ ] **Step 3: Tests**

Add to `handler_test.go`:
- `/auth/google/start` without `GOOGLE_CLIENT_ID` → 503 `not_configured`.
- `/auth/google/start` with creds → 302 to Google, sets `rue_oauth_state` cookie.
- `/auth/google/callback` with state mismatch → 400.
- `/auth/google/callback` happy path: stub `Validator` returns a payload, stub the OAuth exchange via a mock `oauth2` endpoint or skip the real `cfg.Exchange` call by adding a hook. **Simpler path:** make the entire OAuth callback an integration test against a fake Google — too invasive for v1. Instead test:
  - `Service.LoginWithGoogle` directly (creates a user when none exists; links to existing user by email; subsequent same-sub call returns same user).

- [ ] **Step 4: Commit Bundle 4 (Tasks 10-11)**

```bash
git add -A
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -F /tmp/rue-p3-b4.txt
```

`/tmp/rue-p3-b4.txt`:
```
feat(auth): Google OAuth start + callback

- internal/auth/google.go: /auth/google/start (state cookie + redirect),
  /auth/google/callback (state verify, code exchange, ID token validation
  via google.golang.org/api/idtoken, session mint, redirect to frontend)
- internal/auth/service.go: LoginWithGoogle — fast path via oauth_accounts,
  fallback link by email, else create + role, transactional
- Validator interface for stubbed ID-token verification in tests
```

---

### Task 12: Email verification handlers

**Files:**
- Modify: `backend/internal/auth/handler.go` — mount `/auth/verify-email` (POST) and `/auth/verify-email/resend` (POST, auth required)
- Modify: `backend/internal/auth/service.go` — add `VerifyEmail(ctx, rawToken) error`, `ResendVerification(ctx, userID) error`

- [ ] **Step 1: Service methods**

```go
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return ErrInvalidToken
	}
	h := HashToken(rawToken)
	tok, err := s.Repo.GetUnusedVerificationToken(ctx, h[:], "email_verify")
	if errors.Is(err, ErrNotFound) {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	return db.WithTx(ctx, s.Repo.Pool(), func(tx db.Tx) error {
		q := sqlcq.New(tx)
		if err := q.UpdateUserEmailVerified(ctx, sqlcq.UpdateUserEmailVerifiedParams{
			ID: tok.UserID, EmailVerified: true,
		}); err != nil {
			return err
		}
		return q.MarkVerificationTokenUsed(ctx, tok.ID)
	})
}

func (s *Service) ResendVerification(ctx context.Context, userID uuid.UUID, emailAddr string) error {
	if !s.isAllowlisted(emailAddr) {
		// Non-allowlisted users are already auto-verified at signup; nothing to send.
		return nil
	}
	raw, _ := NewToken()
	h := HashToken(raw)
	_, err := s.Repo.q.CreateVerificationToken(ctx, sqlcq.CreateVerificationTokenParams{
		UserID: userID, Kind: "email_verify", TokenHash: h[:],
		ExpiresAt: s.Now().Add(24 * time.Hour),
	})
	if err != nil {
		return err
	}
	return s.Email.Send(ctx, emailAddr, "verify_email", map[string]any{"token": raw})
}
```

- [ ] **Step 2: Handlers**

In `handler.go`:
```go
type verifyEmailBody struct{ Token string `json:"token"` }

// verifyEmail godoc
//
// @Summary  Verify email by token
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body verifyEmailBody true "Token payload"
// @Success  204
// @Failure  400 {object} httpx.ErrorEnvelope
// @Router   /auth/verify-email [post]
func (h *Handlers) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var body verifyEmailBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil); return
	}
	if err := h.Svc.VerifyEmail(r.Context(), body.Token); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid or expired token", nil); return
		}
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "verify failed", nil); return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resendVerification godoc (auth required)
//
// @Summary  Resend verification email
// @Tags     auth
// @Produce  json
// @Success  204
// @Failure  401 {object} httpx.ErrorEnvelope
// @Router   /auth/verify-email/resend [post]
func (h *Handlers) resendVerification(w http.ResponseWriter, r *http.Request) {
	view, ok := GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil); return
	}
	_ = h.Svc.ResendVerification(r.Context(), view.UserID, view.Email)
	w.WriteHeader(http.StatusNoContent)
}
```

Wire in `main.go` — `/auth/verify-email` is public; `/auth/verify-email/resend` mounts inside the `RequireSession` group.

(No commit yet — bundled with Task 13.)

---

### Task 13: Password reset handlers

**Files:**
- Modify: `backend/internal/auth/service.go` — add `RequestPasswordReset(ctx, email)` (always returns nil), `ConfirmPasswordReset(ctx, token, newPassword)`
- Modify: `backend/internal/auth/handler.go` — mount POST `/auth/password-reset/request`, POST `/auth/password-reset/confirm`

- [ ] **Step 1: Service methods**

```go
func (s *Service) RequestPasswordReset(ctx context.Context, emailAddr string) {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	user, err := s.Repo.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		// No-op — don't leak which emails exist. Log internally.
		s.Log.InfoContext(ctx, "password-reset request for unknown email", "email", emailAddr)
		return
	}
	raw, _ := NewToken()
	h := HashToken(raw)
	if _, err := s.Repo.q.CreateVerificationToken(ctx, sqlcq.CreateVerificationTokenParams{
		UserID: user.ID, Kind: "password_reset", TokenHash: h[:],
		ExpiresAt: s.Now().Add(1 * time.Hour),
	}); err != nil {
		s.Log.ErrorContext(ctx, "password-reset token", "err", err)
		return
	}
	_ = s.Email.Send(ctx, emailAddr, "password_reset", map[string]any{"token": raw})
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidCreds
	}
	if rawToken == "" {
		return ErrInvalidToken
	}
	h := HashToken(rawToken)
	tok, err := s.Repo.GetUnusedVerificationToken(ctx, h[:], "password_reset")
	if errors.Is(err, ErrNotFound) {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	hash, err := Hash(newPassword, s.Params)
	if err != nil {
		return err
	}
	return db.WithTx(ctx, s.Repo.Pool(), func(tx db.Tx) error {
		q := sqlcq.New(tx)
		if err := q.UpsertPasswordCredential(ctx, sqlcq.UpsertPasswordCredentialParams{
			UserID: tok.UserID, PasswordHash: hash,
		}); err != nil {
			return err
		}
		if err := q.MarkVerificationTokenUsed(ctx, tok.ID); err != nil {
			return err
		}
		return q.DeleteSessionsForUser(ctx, tok.UserID)
	})
}
```

- [ ] **Step 2: Handlers**

```go
type prReqBody struct { Email string `json:"email"` }
type prConfBody struct { Token, NewPassword string `json:"token,new_password"` } // implementer: fix json tags
```

`prConfBody` should be:
```go
type prConfBody struct {
    Token       string `json:"token"`
    NewPassword string `json:"new_password"`
}
```

```go
func (h *Handlers) passwordResetRequest(w http.ResponseWriter, r *http.Request) {
	var body prReqBody
	_ = httpx.ReadJSON(r, &body)
	h.Svc.RequestPasswordReset(r.Context(), body.Email)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) passwordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var body prConfBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil); return
	}
	if err := h.Svc.ConfirmPasswordReset(r.Context(), body.Token, body.NewPassword); err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid or expired token", nil)
		case errors.Is(err, ErrInvalidCreds):
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "password too short", nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "reset failed", nil)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

Mount both as public POSTs in `Mount(r)`.

- [ ] **Step 3: Tests**

Add to `handler_test.go` / `service_test.go`:
- Request for unknown email → still 204 (test that the handler returns 204 even with `email: nobody@x.test`).
- Request for known email → 204 + verify a row was inserted into `verification_tokens` with `kind='password_reset'`.
- Confirm with valid token → 204; old session is invalidated (`GetSession(oldToken)` returns ErrNoSession).
- Confirm with reused token → 400.

- [ ] **Step 4: Commit Bundle 5 (Tasks 12-13)**

```bash
git add -A
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -F /tmp/rue-p3-b5.txt
```

`/tmp/rue-p3-b5.txt`:
```
feat(auth): email verify + password reset (handlers, email stubbed)

- service: VerifyEmail, ResendVerification, RequestPasswordReset (no
  enumeration), ConfirmPasswordReset (invalidates all sessions)
- handler: POST /auth/verify-email, /auth/verify-email/resend (auth),
  /auth/password-reset/request, /auth/password-reset/confirm
- Email sending via the LogSender stub from Plan 5
```

---

### Task 14: RBAC integration test matrix

**Files:**
- Create: `backend/internal/auth/rbac_test.go`
- Create: `backend/internal/testsupport/auth.go` — `LoginAs(t, app, role string) (cookie *http.Cookie)` helper

**Purpose:** spec Section 10.3 mandates an integration test matrix asserting:
- Anonymous → every protected route → 401.
- Customer → admin routes → 403.
- Admin → admin routes → 200 (no admin routes exist yet, but the matrix needs to be ready for Plan 7 — for now, mount a stub admin route in the test).

- [ ] **Step 1: testsupport helper**

File: `backend/internal/testsupport/auth.go`
```go
package testsupport

// LoginAs signs up a user with the given role and returns the cookie.
// (Implementer: depends on auth.Service. Accept it as a parameter rather
// than importing the auth package to avoid an import cycle from testsupport.)
```

Defer the actual implementation to a small in-package helper instead — testsupport can't import auth without a cycle. Use a closure passed from the test:

In `rbac_test.go` define `signUp(t, svc, h, email, role) *http.Cookie` directly. Don't lift to testsupport this round.

- [ ] **Step 2: Test matrix**

File: `backend/internal/auth/rbac_test.go`
- Set up the full HTTP stack with chi (matching what `main.go` does), mount: auth handlers (public), `/me` under RequireSession, plus a stub `/admin/ping` under `RequireSession` + `RequireRole("admin")`.
- Sign up a customer (default `customer` role).
- Manually insert a second user via the repository + `q.AddUserRole(...,"admin")` to get an admin.
- Cases:
  - Anonymous → GET /me → 401.
  - Anonymous → GET /admin/ping → 401.
  - Customer cookie → GET /me → 200.
  - Customer cookie → GET /admin/ping → 403.
  - Admin cookie → GET /me → 200 with `role: "admin"`.
  - Admin cookie → GET /admin/ping → 200.

This is the first plan that exercises the full RBAC story. Subsequent plans will reuse this pattern.

(No commit yet — bundled with Task 15.)

---

### Task 15: Regenerate OpenAPI + final drift check

- [ ] **Step 1: Regenerate**

```bash
cd backend
swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency
```

- [ ] **Step 2: Verify all auth routes present**

```bash
for path in /auth/signup /auth/login /auth/logout /auth/session \
            /auth/google/start /auth/google/callback \
            /auth/verify-email /auth/verify-email/resend \
            /auth/password-reset/request /auth/password-reset/confirm \
            /me; do
    grep -c "\"$path\"" docs/swagger.json
done
```
Each should print `1`.

- [ ] **Step 3: Run drift-check + full test suite**

```bash
cd ..
make drift-check
make test
```
Both exit 0.

- [ ] **Step 4: Commit Bundle 6 (Tasks 14-15)**

```bash
git add -A
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -F /tmp/rue-p3-b6.txt
```

`/tmp/rue-p3-b6.txt`:
```
feat(auth): RBAC integration matrix, regenerate openapi

- rbac_test.go: anonymous/customer/admin × /me + stub /admin/ping
- docs: openapi includes all 10 auth routes + /me
```

---

## Verification — end of Plan 3

When all 15 tasks are complete:

- [ ] `make test` exits 0 (auth + service + handler + middleware + email + me + RBAC + everything from Plans 1-2 still green).
- [ ] `make drift-check` exits 0.
- [ ] `make up && make dev` boots the server. The following sequence works end-to-end:
  - `POST /api/v1/auth/signup {"email":"x@y.test","password":"hunter22","name":"X"}` → 201, cookie set, body `{user_id, email, role:"customer", email_verified:true}`.
  - `GET /api/v1/me` with the cookie → 200 with the user details.
  - `POST /api/v1/auth/logout` → 204, cookie cleared.
  - `GET /api/v1/me` after logout → 401.
  - `POST /api/v1/auth/login` with the same creds → 200, new cookie.
  - `POST /api/v1/auth/password-reset/request {"email":"nobody@x.test"}` → 204 (no enumeration).
- [ ] `git log --oneline 8136426..HEAD` shows 6 new commits, all prefixed `feat(auth):`.
- [ ] No `fmt.Sprintf` building SQL anywhere.
- [ ] No raw session token logged anywhere (`grep -rn 'SessionToken\b' backend/ | grep -v test | grep -v _test.go` reveals no stray log lines).

Plan 4 (Cart with guest tokens + merge) picks up from this baseline.

## Self-Review Notes

- **Spec coverage:** This plan implements Section 4.1 identity tables, Section 5.1 auth ownership boundary, Section 5.2 auth endpoints, Section 6.1 auth flows (signup/login/logout/Google/verify/reset), Section 10.3 RBAC layers 1-2 (router group + handler-level MustBeAdmin — layer 3 row-scoping arrives with cart/orders in Plans 4-6), and Section 14 security checklist for the auth surface.
- **Email allowlist behavior:** implemented per spec — non-allowlisted addresses are server-side auto-verified at signup, no email queued. Allowlisted addresses receive the (stubbed) verify email.
- **`netToInet` brittleness:** the `pgtype.Inet` adapter changes across pgx minor versions. If compilation fails, drop IP capture to `pgtype.Inet{Valid: false}` and surface it as a TODO; not worth blocking Plan 3.
- **Service test speed:** uses `TestParams` (`m=8 MiB, t=1, p=1`) — full auth test suite should finish under 30s including testcontainers cold-start.
- **Test coverage gap deferred:** the Google OAuth callback's `cfg.Exchange` round-trip is not directly tested (would require a fake Google authorization endpoint). The `Validator` interface and `LoginWithGoogle` service method are both directly tested; the wire-level callback handler test asserts state-cookie verification but stubs `cfg.Exchange`. Plan 15 may add a Playwright happy-path against Google's OAuth playground if desired.
- **Naming carryover:** the `Handlers` struct gained four Google-related fields (Task 10). The constructor signature changes — every existing caller in `cmd/api/main.go` and tests must update. Implementer should grep for `NewHandlers(` in `internal/auth/` and `cmd/api/` after Task 10 lands.
