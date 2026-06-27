# Backend Foundation Implementation Plan (Plan 1 of 15)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a bootable Go HTTP server at `backend/cmd/api` that wires Postgres (via pgx/v5 + sqlc + goose), a typed Config, structured logging, the chi router with cross-cutting middleware (request ID, recovery, CORS, error envelope), a `/healthz` endpoint that pings the DB, and an OpenAPI generation pipeline whose output is committed and drift-checked. No business endpoints yet — that's Plan 2 onwards.

**Architecture:** Three-layer backend (handler → service → repository), built around a `db.Pool` and an `app.Application` DI container. All HTTP machinery isolated under `internal/httpx`. OpenAPI is the single source of truth at the wire boundary; `swaggo/swag` emits `backend/docs/openapi.json` which a later plan feeds into Orval for the frontend.

**Tech Stack:** Go 1.22+, chi v5, pgx/v5 + pgxpool, sqlc (with `pgx/v5` driver), goose v3, `kelseyhightower/envconfig`, `swaggo/swag` + `swaggo/http-swagger`, `httprate`, `slog` (stdlib), `air` for hot reload, `testcontainers-go` for integration tests, Postgres 16.

## Global Constraints

- **Go module path:** `github.com/oti-adjei/ruecosmetics`. Substitute this verbatim in `go.mod`, all imports, and all generated code.
- **Go version floor:** Go 1.22 (`go 1.22` in `go.mod`).
- **Project working directory:** `casestud/ruecosmetics/`. All paths in this plan are relative to that directory unless prefixed with `backend/` (which is relative to `casestud/ruecosmetics/backend/`).
- **Money type:** every monetary column is `BIGINT` in pesewas. Plan 1 has no money columns yet but the convention is established by `sqlc.yaml` overrides for `BIGINT → int64`.
- **UUID generation:** `gen_random_uuid()` via the `pgcrypto` extension. The first migration enables it.
- **Error response envelope:** `{"error": {"code": "...", "message": "...", "fields": {...}?}}`. No other top-level shape is allowed for non-2xx responses.
- **CORS origins:** read from `CORS_ORIGINS` env (comma-separated). No hardcoded origins.
- **Git commit identity (user preference):** `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit …` — every commit command in this plan uses this flag.
- **Existing state:** the repo at `casestud/ruecosmetics/` already contains `docs/superpowers/{specs,plans}/...` and has two existing commits. There is no `backend/` directory yet; Task 1 creates it.

## File Structure (created by this plan)

```
casestud/ruecosmetics/
├── .gitignore                          # repo-level: node_modules, dist, .env, *.log, .air, tmp
├── docker-compose.yml                  # postgres + mailpit services
├── Makefile                            # delegates: dev, test, openapi, drift-check
├── README.md                           # quickstart + prerequisites
└── backend/
    ├── .air.toml                       # hot-reload config
    ├── .env.example                    # all env vars used by Config (no secrets)
    ├── Makefile                        # backend-specific targets
    ├── go.mod                          # module declaration
    ├── go.sum                          # (generated)
    ├── sqlc.yaml                       # sqlc config (pgx/v5 driver)
    ├── cmd/api/main.go                 # entrypoint: build cfg → app → server → ListenAndServe
    ├── internal/
    │   ├── app/
    │   │   └── app.go                  # Application struct (Config, Pool, Logger), New()
    │   ├── config/
    │   │   ├── config.go               # typed Config + LoadFromEnv()
    │   │   └── config_test.go
    │   ├── db/
    │   │   ├── db.go                   # NewPool(), WithTx() helper
    │   │   ├── db_test.go              # integration: pool connects, WithTx commits/rolls back
    │   │   └── sqlc/                   # sqlc-generated code (do not edit by hand)
    │   │       ├── db.go               # (generated)
    │   │       ├── models.go           # (generated)
    │   │       └── health.sql.go       # (generated)
    │   ├── health/
    │   │   ├── handler.go              # GET /healthz + swaggo annotation
    │   │   └── handler_test.go         # integration: returns 200 + ok body
    │   └── httpx/
    │       ├── error.go                # ErrorEnvelope, ErrorCode constants, WriteError
    │       ├── error_test.go
    │       ├── json.go                 # ReadJSON, WriteJSON helpers
    │       ├── json_test.go
    │       ├── middleware.go           # RequestID + Recovery
    │       ├── middleware_test.go
    │       ├── cors.go                 # CORS from allowlist
    │       └── cors_test.go
    ├── migrations/
    │   └── 00001_enable_pgcrypto.sql   # CREATE EXTENSION + a tiny health table
    ├── queries/
    │   └── health.sql                  # sqlc input: HealthPing query
    └── docs/
        ├── docs.go                     # (generated by swag)
        ├── swagger.json                # (generated by swag)
        └── swagger.yaml                # (generated by swag)
```

Why a "tiny health table"? Plan 1 needs a real sqlc query to prove the codegen pipeline works end-to-end; an empty schema gives sqlc nothing to generate. The health table is one row, used only to validate connectivity and code generation; it's harmless if it stays.

---

### Task 1: Create repo skeleton, go.mod, .gitignore, docker-compose

**Files:**
- Create: `casestud/ruecosmetics/.gitignore`
- Create: `casestud/ruecosmetics/docker-compose.yml`
- Create: `casestud/ruecosmetics/backend/go.mod`
- Create: `casestud/ruecosmetics/backend/.env.example`
- Create: `casestud/ruecosmetics/backend/.air.toml`

**Interfaces:**
- Consumes: nothing (first task).
- Produces:
  - `docker compose up -d postgres` → Postgres 16 listening on `localhost:5432`, DB `ruecosmetics`, user `rue`, password `rue_dev`.
  - `docker compose up -d mailpit` → SMTP on `:1025`, UI on `:8025`.
  - Go module `github.com/oti-adjei/ruecosmetics` declared.

- [ ] **Step 1: Create `.gitignore`**

File: `casestud/ruecosmetics/.gitignore`
```
# binaries
backend/cmd/api/api
backend/tmp/
*.exe

# env
.env
.env.local
*.env

# editors / OS
.DS_Store
.idea/
.vscode/

# node (for Plan 9+)
node_modules/
frontend/dist/

# go build cache (already in $GOPATH but harmless)
*.test
*.out

# air
.air.toml.bak
```

- [ ] **Step 2: Create `docker-compose.yml`**

File: `casestud/ruecosmetics/docker-compose.yml`
```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: rue_postgres
    environment:
      POSTGRES_DB: ruecosmetics
      POSTGRES_USER: rue
      POSTGRES_PASSWORD: rue_dev
    ports:
      - "5432:5432"
    volumes:
      - rue_pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U rue -d ruecosmetics"]
      interval: 2s
      timeout: 3s
      retries: 10

  mailpit:
    image: axllent/mailpit:latest
    container_name: rue_mailpit
    ports:
      - "1025:1025"
      - "8025:8025"

volumes:
  rue_pg_data:
```

- [ ] **Step 3: Initialize Go module**

Run from `casestud/ruecosmetics/backend/`:
```bash
go mod init github.com/oti-adjei/ruecosmetics
```

Verify `backend/go.mod` contains exactly:
```
module github.com/oti-adjei/ruecosmetics

go 1.22
```

- [ ] **Step 4: Create `backend/.env.example`**

File: `casestud/ruecosmetics/backend/.env.example`
```
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://rue:rue_dev@localhost:5432/ruecosmetics?sslmode=disable

# CORS (comma-separated)
CORS_ORIGINS=http://localhost:5173

# Logging
LOG_LEVEL=debug
```

- [ ] **Step 5: Create `backend/.air.toml`**

File: `casestud/ruecosmetics/backend/.air.toml`
```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/api ./cmd/api"
bin = "./tmp/api"
include_ext = ["go"]
exclude_dir = ["tmp", "docs", "internal/db/sqlc"]
delay = 200
stop_on_error = true

[log]
time = true
```

- [ ] **Step 6: Smoke-test docker-compose**

```bash
cd casestud/ruecosmetics
docker compose up -d postgres
docker compose ps
```

Expected: `rue_postgres` is `Up (healthy)` within ~10 seconds. Then:
```bash
docker exec rue_postgres psql -U rue -d ruecosmetics -c 'select 1'
```
Expected output contains `1 row`.

Stop the container:
```bash
docker compose down
```

- [ ] **Step 7: Commit**

```bash
cd casestud/ruecosmetics
git add .gitignore docker-compose.yml backend/go.mod backend/.env.example backend/.air.toml
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "chore: scaffold backend module and dev infrastructure"
```

---

### Task 2: Typed Config with envconfig

**Files:**
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/config/config_test.go`

**Interfaces:**
- Produces: `config.Config` struct, `config.Load() (*Config, error)`.
  ```go
  type Config struct {
      Port         int
      Env          string         // "development" | "production"
      DatabaseURL  string
      CORSOrigins  []string
      LogLevel     string         // "debug" | "info" | "warn" | "error"
  }
  ```
- Consumed by: Task 10 (`app.Application`), Task 12 (`cmd/api/main.go`).

- [ ] **Step 1: Add envconfig dependency**

```bash
cd casestud/ruecosmetics/backend
go get github.com/kelseyhightower/envconfig@v1.4.0
```

- [ ] **Step 2: Write failing test**

File: `backend/internal/config/config_test.go`
```go
package config_test

import (
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/config"
)

func TestLoadParsesEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost:5432/z")
	t.Setenv("CORS_ORIGINS", "https://a.com,https://b.com")
	t.Setenv("LOG_LEVEL", "info")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want production", cfg.Env)
	}
	if cfg.DatabaseURL != "postgres://x:y@localhost:5432/z" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if len(cfg.CORSOrigins) != 2 || cfg.CORSOrigins[0] != "https://a.com" || cfg.CORSOrigins[1] != "https://b.com" {
		t.Errorf("CORSOrigins = %v", cfg.CORSOrigins)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q", cfg.LogLevel)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	// envconfig treats empty as zero-value unless `required:"true"` is set.
	t.Setenv("DATABASE_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is empty")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd casestud/ruecosmetics/backend
go test ./internal/config/...
```
Expected: compile error — `config` package does not exist.

- [ ] **Step 4: Implement Config**

File: `backend/internal/config/config.go`
```go
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port        int      `envconfig:"PORT" default:"8080"`
	Env         string   `envconfig:"ENV" default:"development"`
	DatabaseURL string   `envconfig:"DATABASE_URL" required:"true"`
	CORSOrigins []string `envconfig:"CORS_ORIGINS" default:"http://localhost:5173"`
	LogLevel    string   `envconfig:"LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return &cfg, nil
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./internal/config/... -v
```
Expected: both tests PASS.

- [ ] **Step 6: Commit**

```bash
cd casestud/ruecosmetics
git add backend/go.mod backend/go.sum backend/internal/config/
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(config): typed Config loaded via envconfig"
```

---

### Task 3: Postgres pool wiring (pgx/v5 + WithTx helper)

**Files:**
- Create: `backend/internal/db/db.go`
- Create: `backend/internal/db/db_test.go`

**Interfaces:**
- Produces:
  ```go
  type Pool = *pgxpool.Pool

  func NewPool(ctx context.Context, databaseURL string) (Pool, error)
  func WithTx(ctx context.Context, pool Pool, fn func(pgx.Tx) error) error
  ```
- Consumed by: Task 5 (sqlc generated code uses `*pgxpool.Pool`), Task 10 (`app.Application.DB`), Task 11 (health handler pings via Pool).

- [ ] **Step 1: Add pgx dependency**

```bash
cd casestud/ruecosmetics/backend
go get github.com/jackc/pgx/v5@v5.7.1
go get github.com/jackc/pgx/v5/pgxpool@v5.7.1
go get github.com/testcontainers/testcontainers-go@v0.34.0
go get github.com/testcontainers/testcontainers-go/modules/postgres@v0.34.0
```

- [ ] **Step 2: Write failing integration test**

File: `backend/internal/db/db_test.go`
```go
package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go"
)

func startPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ruetest"),
		postgres.WithUsername("rue"),
		postgres.WithPassword("rue_dev"),
		testcontainers.WithWaitStrategy(postgres.DefaultWaitStrategy()),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	url, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	return url, func() { _ = pg.Terminate(ctx) }
}

func TestNewPoolConnects(t *testing.T) {
	url, stop := startPostgres(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestWithTxCommits(t *testing.T) {
	url, stop := startPostgres(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `CREATE TABLE t (v int)`)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	err = db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO t (v) VALUES (1)`)
		return err
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM t`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Errorf("count = %d, want 1", n)
	}
}

func TestWithTxRollsBackOnError(t *testing.T) {
	url, stop := startPostgres(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()
	_, _ = pool.Exec(ctx, `CREATE TABLE t (v int)`)

	sentinel := errors.New("boom")
	err = db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		_, _ = tx.Exec(ctx, `INSERT INTO t (v) VALUES (1)`)
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
	var n int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM t`).Scan(&n)
	if n != 0 {
		t.Errorf("count = %d, want 0 (rolled back)", n)
	}
}
```

Also add `import "errors"` to the test file (the Step 2 listing already implies it; verify it's in the import block when pasting).

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/db/... -v
```
Expected: compile error — `db` package not found.

- [ ] **Step 4: Implement `db.go`**

File: `backend/internal/db/db.go`
```go
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool = *pgxpool.Pool

func NewPool(ctx context.Context, databaseURL string) (Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

func WithTx(ctx context.Context, pool Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/db/... -v -timeout=120s
```
Expected: all three tests PASS. (Docker must be running locally — testcontainers will spin Postgres images.)

- [ ] **Step 6: Commit**

```bash
cd casestud/ruecosmetics
git add backend/go.mod backend/go.sum backend/internal/db/
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(db): pgx/v5 pool and WithTx helper with integration tests"
```

---

### Task 4: goose migrations setup + first migration

**Files:**
- Create: `backend/migrations/00001_enable_pgcrypto.sql`
- Modify: `backend/Makefile` (created in Task 13; for now create a stub)
- Create: `backend/internal/db/migrate_test.go` (proves migrations apply)

**Interfaces:**
- Produces: a usable `goose -dir migrations postgres "$DATABASE_URL" up` flow. Migration 00001 enables `pgcrypto` and creates table `health_marker (id uuid primary key default gen_random_uuid(), created_at timestamptz not null default now())`.
- Consumed by: Task 5 (sqlc query targets `health_marker`).

- [ ] **Step 1: Install goose CLI dependency for tests**

```bash
cd casestud/ruecosmetics/backend
go get github.com/pressly/goose/v3@v3.22.1
```

- [ ] **Step 2: Write the migration**

File: `backend/migrations/00001_enable_pgcrypto.sql`
```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE health_marker (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO health_marker DEFAULT VALUES;

-- +goose Down
DROP TABLE health_marker;
DROP EXTENSION IF EXISTS pgcrypto;
```

- [ ] **Step 3: Write failing test**

File: `backend/internal/db/migrate_test.go`
```go
package db_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/pressly/goose/v3"
)

func TestMigrationsApply(t *testing.T) {
	url, stop := startPostgres(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()

	migDir, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	if err := goose.UpContext(ctx, sqlDB, migDir); err != nil {
		t.Fatalf("up: %v", err)
	}

	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM health_marker`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n < 1 {
		t.Errorf("health_marker rows = %d, want >= 1", n)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/db/... -run TestMigrationsApply -v -timeout=120s
```
Expected: PASS — applies the single migration, finds at least one row in `health_marker`.

- [ ] **Step 5: Commit**

```bash
cd casestud/ruecosmetics
git add backend/go.mod backend/go.sum backend/migrations/ backend/internal/db/migrate_test.go
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(db): first goose migration (pgcrypto + health_marker)"
```

---

### Task 5: sqlc setup + first generated query

**Files:**
- Create: `backend/sqlc.yaml`
- Create: `backend/queries/health.sql`
- Create: `backend/internal/db/sqlc/` (generated content checked in)
- Create: `backend/internal/db/sqlc_test.go` (uses the generated Queries against a migrated DB)

**Interfaces:**
- Produces: `package sqlc` with type `Queries` and method `CountHealthMarkers(ctx, db DBTX) (int64, error)`.
- Consumed by: Task 11 (health handler optionally uses it; can also just `pool.Ping`).

- [ ] **Step 1: Install sqlc**

If not already on PATH:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0
```
Verify: `sqlc version` prints `v1.27.0` or compatible.

- [ ] **Step 2: Write `sqlc.yaml`**

File: `backend/sqlc.yaml`
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries"
    schema: "migrations"
    gen:
      go:
        package: "sqlc"
        out: "internal/db/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: false
        emit_pointers_for_null_types: true
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
```

- [ ] **Step 3: Add uuid dep**

```bash
cd casestud/ruecosmetics/backend
go get github.com/google/uuid@v1.6.0
```

- [ ] **Step 4: Write the query**

File: `backend/queries/health.sql`
```sql
-- name: CountHealthMarkers :one
SELECT count(*) FROM health_marker;
```

- [ ] **Step 5: Generate**

```bash
cd casestud/ruecosmetics/backend
sqlc generate
```
Expected: creates `internal/db/sqlc/db.go`, `internal/db/sqlc/models.go`, `internal/db/sqlc/health.sql.go`. No errors.

- [ ] **Step 6: Write integration test**

File: `backend/internal/db/sqlc_test.go`
```go
package db_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/pressly/goose/v3"
)

func TestSqlcCountHealthMarkers(t *testing.T) {
	url, stop := startPostgres(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sqlDB, _ := sql.Open("pgx", url)
	defer sqlDB.Close()
	_ = goose.SetDialect("postgres")
	migDir, _ := filepath.Abs("../../migrations")
	if err := goose.UpContext(ctx, sqlDB, migDir); err != nil {
		t.Fatalf("up: %v", err)
	}

	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	q := sqlcq.New(pool)
	n, err := q.CountHealthMarkers(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n < 1 {
		t.Errorf("n = %d, want >= 1", n)
	}
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./internal/db/... -v -timeout=180s
```
Expected: all `db_test` tests PASS, including the new sqlc one.

- [ ] **Step 8: Commit**

```bash
cd casestud/ruecosmetics
git add backend/sqlc.yaml backend/queries/ backend/internal/db/sqlc/ backend/internal/db/sqlc_test.go backend/go.mod backend/go.sum
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(db): sqlc pipeline with first generated query"
```

---

### Task 6: httpx — error envelope + JSON helpers

**Files:**
- Create: `backend/internal/httpx/error.go`
- Create: `backend/internal/httpx/error_test.go`
- Create: `backend/internal/httpx/json.go`
- Create: `backend/internal/httpx/json_test.go`

**Interfaces:**
- Produces:
  ```go
  // error.go
  type ErrorEnvelope struct {
      Error ErrorBody `json:"error"`
  }
  type ErrorBody struct {
      Code    string            `json:"code"`
      Message string            `json:"message"`
      Fields  map[string]string `json:"fields,omitempty"`
  }
  const (
      CodeBadRequest    = "bad_request"
      CodeUnauthorized  = "unauthorized"
      CodeForbidden     = "forbidden"
      CodeNotFound      = "not_found"
      CodeConflict      = "conflict"
      CodeInternal      = "internal_error"
      CodeValidation    = "validation_failed"
  )
  func WriteError(w http.ResponseWriter, status int, code, message string, fields map[string]string)

  // json.go
  func ReadJSON(r *http.Request, dst any) error
  func WriteJSON(w http.ResponseWriter, status int, body any)
  ```
- Consumed by: every handler in this and future plans.

- [ ] **Step 1: Write failing tests**

File: `backend/internal/httpx/error_test.go`
```go
package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func TestWriteErrorShape(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(rec, http.StatusBadRequest, httpx.CodeBadRequest, "boom", map[string]string{"x": "missing"})
	if rec.Code != 400 {
		t.Errorf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}
	var env httpx.ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.Code != "bad_request" || env.Error.Message != "boom" {
		t.Errorf("envelope = %+v", env)
	}
	if env.Error.Fields["x"] != "missing" {
		t.Errorf("fields = %+v", env.Error.Fields)
	}
}

func TestWriteErrorOmitsFieldsWhenNil(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(rec, http.StatusInternalServerError, httpx.CodeInternal, "oops", nil)
	if got := rec.Body.String(); got == "" {
		t.Fatal("empty body")
	}
	// fields should not appear in JSON when nil
	if got := rec.Body.String(); contains(got, "fields") {
		t.Errorf("expected no fields key, got %s", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && len(s) > 0 && indexOf(s, sub) >= 0))
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
```

File: `backend/internal/httpx/json_test.go`
```go
package httpx_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func TestReadJSONOK(t *testing.T) {
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	var v struct{ Name string }
	if err := httpx.ReadJSON(req, &v); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if v.Name != "x" {
		t.Errorf("Name = %q", v.Name)
	}
}

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"x","extra":1}`))
	var v struct{ Name string }
	if err := httpx.ReadJSON(req, &v); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestWriteJSONSetsHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteJSON(rec, http.StatusOK, map[string]string{"ok": "yes"})
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("ct = %q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), `"ok":"yes"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd casestud/ruecosmetics/backend
go test ./internal/httpx/...
```
Expected: compile error — package doesn't exist.

- [ ] **Step 3: Implement `error.go`**

File: `backend/internal/httpx/error.go`
```go
package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

const (
	CodeBadRequest   = "bad_request"
	CodeUnauthorized = "unauthorized"
	CodeForbidden    = "forbidden"
	CodeNotFound     = "not_found"
	CodeConflict     = "conflict"
	CodeInternal     = "internal_error"
	CodeValidation   = "validation_failed"
)

func WriteError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Error: ErrorBody{Code: code, Message: message, Fields: fields},
	})
}
```

- [ ] **Step 4: Implement `json.go`**

File: `backend/internal/httpx/json.go`
```go
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxBody = 1 << 20 // 1 MiB

func ReadJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var syn *json.SyntaxError
		switch {
		case errors.As(err, &syn):
			return fmt.Errorf("malformed JSON at byte %d", syn.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("empty body")
		default:
			return err
		}
	}
	if dec.More() {
		return errors.New("body must contain a single JSON value")
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/httpx/... -v
```
Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
cd casestud/ruecosmetics
git add backend/internal/httpx/error.go backend/internal/httpx/error_test.go backend/internal/httpx/json.go backend/internal/httpx/json_test.go
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(httpx): error envelope and JSON helpers"
```

---

### Task 7: httpx — request ID + recovery middleware

**Files:**
- Create: `backend/internal/httpx/middleware.go`
- Create: `backend/internal/httpx/middleware_test.go`

**Interfaces:**
- Produces:
  ```go
  type ctxKey int
  const RequestIDKey ctxKey = 1
  func RequestID(next http.Handler) http.Handler   // sets/propagates X-Request-Id
  func Recovery(logger *slog.Logger) func(http.Handler) http.Handler   // recovers panics, writes CodeInternal envelope
  func GetRequestID(ctx context.Context) string
  ```
- Consumed by: Task 12 (`cmd/api/main.go` router wiring).

- [ ] **Step 1: Add chi dependency**

```bash
cd casestud/ruecosmetics/backend
go get github.com/go-chi/chi/v5@v5.1.0
```

- [ ] **Step 2: Write failing tests**

File: `backend/internal/httpx/middleware_test.go`
```go
package httpx_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func TestRequestIDPropagatesIncoming(t *testing.T) {
	var seen string
	h := httpx.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = httpx.GetRequestID(r.Context())
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "rid-abc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if seen != "rid-abc" {
		t.Errorf("seen = %q", seen)
	}
	if rec.Header().Get("X-Request-Id") != "rid-abc" {
		t.Errorf("response header = %q", rec.Header().Get("X-Request-Id"))
	}
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	var seen string
	h := httpx.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = httpx.GetRequestID(r.Context())
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if seen == "" {
		t.Fatal("expected generated request id")
	}
	if rec.Header().Get("X-Request-Id") != seen {
		t.Errorf("response header mismatch: %q vs %q", rec.Header().Get("X-Request-Id"), seen)
	}
}

func TestRecoveryReturnsEnvelope(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := httpx.Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 500 {
		t.Errorf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./internal/httpx/... -run "RequestID|Recovery" -v
```
Expected: compile error — symbols don't exist yet.

- [ ] **Step 4: Implement middleware**

File: `backend/internal/httpx/middleware.go`
```go
package httpx

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey int

const RequestIDKey ctxKey = 1

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", rid)
		ctx := context.WithValue(r.Context(), RequestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						slog.Any("panic", rec),
						slog.String("request_id", GetRequestID(r.Context())),
						slog.String("path", r.URL.Path),
					)
					WriteError(w, http.StatusInternalServerError, CodeInternal, "internal server error", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/httpx/... -v
```
Expected: all middleware tests PASS, prior httpx tests still PASS.

- [ ] **Step 6: Commit**

```bash
cd casestud/ruecosmetics
git add backend/internal/httpx/middleware.go backend/internal/httpx/middleware_test.go backend/go.mod backend/go.sum
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(httpx): request id + panic recovery middleware"
```

---

### Task 8: httpx — CORS middleware

**Files:**
- Create: `backend/internal/httpx/cors.go`
- Create: `backend/internal/httpx/cors_test.go`

**Interfaces:**
- Produces:
  ```go
  func CORS(allowedOrigins []string) func(http.Handler) http.Handler
  ```
  Behavior: if request's `Origin` header matches an entry, set `Access-Control-Allow-Origin: <origin>`, `Access-Control-Allow-Credentials: true`, `Access-Control-Allow-Methods: GET,POST,PATCH,DELETE,OPTIONS`, `Access-Control-Allow-Headers: Content-Type,X-Request-Id`, `Vary: Origin`. For `OPTIONS` requests, write 204 immediately. Non-matching origins get no CORS headers (browser blocks).
- Consumed by: Task 12.

- [ ] **Step 1: Write failing tests**

File: `backend/internal/httpx/cors_test.go`
```go
package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func TestCORSAllowsListedOrigin(t *testing.T) {
	h := httpx.CORS([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("credentials = %q", got)
	}
}

func TestCORSRejectsUnlistedOrigin(t *testing.T) {
	h := httpx.CORS([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no ACAO, got %q", got)
	}
}

func TestCORSPreflightReturns204(t *testing.T) {
	h := httpx.CORS([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run on preflight")
	}))
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("code = %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/httpx/... -run CORS -v
```
Expected: compile error.

- [ ] **Step 3: Implement CORS**

File: `backend/internal/httpx/cors.go`
```go
package httpx

import "net/http"

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if _, ok := allowed[origin]; ok && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Request-Id")
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/httpx/... -v
```
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
cd casestud/ruecosmetics
git add backend/internal/httpx/cors.go backend/internal/httpx/cors_test.go
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(httpx): CORS middleware with origin allowlist"
```

---

### Task 9: Structured logger via slog

**Files:**
- Create: `backend/internal/app/logger.go`
- Create: `backend/internal/app/logger_test.go`

**Interfaces:**
- Produces:
  ```go
  func NewLogger(level, env string) *slog.Logger
  ```
  Behavior: `env=="development"` → text handler to stdout. Else JSON handler. Level parsed from `level` (debug/info/warn/error; default info on unknown).
- Consumed by: Task 10, Task 12.

- [ ] **Step 1: Write failing test**

File: `backend/internal/app/logger_test.go`
```go
package app_test

import (
	"log/slog"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/app"
)

func TestNewLoggerLevels(t *testing.T) {
	cases := []struct {
		in    string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"garbage", slog.LevelInfo},
	}
	for _, c := range cases {
		l := app.NewLogger(c.in, "development")
		if !l.Enabled(nil, c.want) {
			t.Errorf("level %s: not enabled at want %v", c.in, c.want)
		}
	}
}

func TestNewLoggerNotNil(t *testing.T) {
	if app.NewLogger("info", "production") == nil {
		t.Fatal("nil logger")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test ./internal/app/...
```
Expected: package not found.

- [ ] **Step 3: Implement**

File: `backend/internal/app/logger.go`
```go
package app

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(level, env string) *slog.Logger {
	lvl := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: lvl}
	if env == "development" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/app/... -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd casestud/ruecosmetics
git add backend/internal/app/logger.go backend/internal/app/logger_test.go
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(app): structured logger via slog"
```

---

### Task 10: Application struct (DI container)

**Files:**
- Create: `backend/internal/app/app.go`
- Create: `backend/internal/app/app_test.go`

**Interfaces:**
- Produces:
  ```go
  type Application struct {
      Config *config.Config
      Pool   db.Pool
      Logger *slog.Logger
  }
  func New(ctx context.Context, cfg *config.Config) (*Application, error)
  func (a *Application) Close()
  ```
- Consumed by: Task 11 (handler reads from app), Task 12 (main builds Application).

- [ ] **Step 1: Write failing integration test**

File: `backend/internal/app/app_test.go`
```go
package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestApplicationNewWiresPool(t *testing.T) {
	ctx := context.Background()
	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ruetest"),
		postgres.WithUsername("rue"),
		postgres.WithPassword("rue_dev"),
		testcontainers.WithWaitStrategy(postgres.DefaultWaitStrategy()),
	)
	if err != nil {
		t.Fatalf("pg: %v", err)
	}
	defer pg.Terminate(ctx)
	url, _ := pg.ConnectionString(ctx, "sslmode=disable")

	cfg := &config.Config{
		Port: 0, Env: "development",
		DatabaseURL: url,
		CORSOrigins: []string{"http://localhost:5173"},
		LogLevel:    "debug",
	}
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	a, err := app.New(c, cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Close()
	if a.Pool == nil || a.Logger == nil || a.Config == nil {
		t.Errorf("nil field on Application")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test ./internal/app/... -run TestApplicationNew
```
Expected: undefined `app.New`.

- [ ] **Step 3: Implement**

File: `backend/internal/app/app.go`
```go
package app

import (
	"context"

	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"log/slog"
)

type Application struct {
	Config *config.Config
	Pool   db.Pool
	Logger *slog.Logger
}

func New(ctx context.Context, cfg *config.Config) (*Application, error) {
	logger := NewLogger(cfg.LogLevel, cfg.Env)
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return &Application{Config: cfg, Pool: pool, Logger: logger}, nil
}

func (a *Application) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/app/... -v -timeout=120s
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd casestud/ruecosmetics
git add backend/internal/app/app.go backend/internal/app/app_test.go
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(app): Application DI container"
```

---

### Task 11: Health handler

**Files:**
- Create: `backend/internal/health/handler.go`
- Create: `backend/internal/health/handler_test.go`

**Interfaces:**
- Produces:
  ```go
  func Handler(app *app.Application) http.HandlerFunc
  ```
  Behavior: ping the pool with a 2s timeout; on success return `200 {"status":"ok","db":"ok"}`; on failure return `503` envelope with `CodeInternal`. Swaggo annotation comment included.
- Consumed by: Task 12 (router mounts at `/healthz`).

- [ ] **Step 1: Write failing test**

File: `backend/internal/health/handler_test.go`
```go
package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/health"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestHealthOK(t *testing.T) {
	ctx := context.Background()
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("ruetest"), postgres.WithUsername("rue"), postgres.WithPassword("rue_dev"),
		testcontainers.WithWaitStrategy(postgres.DefaultWaitStrategy()),
	)
	if err != nil {
		t.Fatalf("pg: %v", err)
	}
	defer pg.Terminate(ctx)
	url, _ := pg.ConnectionString(ctx, "sslmode=disable")
	cfg := &config.Config{Env: "development", DatabaseURL: url, CORSOrigins: []string{"http://localhost:5173"}, LogLevel: "debug"}
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	a, err := app.New(c, cfg)
	if err != nil {
		t.Fatalf("app: %v", err)
	}
	defer a.Close()

	rec := httptest.NewRecorder()
	health.Handler(a)(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body struct{ Status, DB string }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Status != "ok" || body.DB != "ok" {
		t.Errorf("body = %+v", body)
	}
}

func TestHealthDownReturns503(t *testing.T) {
	// closed pool → ping fails
	ctx := context.Background()
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("ruetest"), postgres.WithUsername("rue"), postgres.WithPassword("rue_dev"),
		testcontainers.WithWaitStrategy(postgres.DefaultWaitStrategy()),
	)
	if err != nil {
		t.Fatalf("pg: %v", err)
	}
	url, _ := pg.ConnectionString(ctx, "sslmode=disable")
	cfg := &config.Config{Env: "development", DatabaseURL: url, CORSOrigins: []string{"http://localhost:5173"}, LogLevel: "debug"}
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	a, err := app.New(c, cfg)
	if err != nil {
		t.Fatalf("app: %v", err)
	}
	a.Pool.Close()
	pg.Terminate(ctx)

	rec := httptest.NewRecorder()
	health.Handler(a)(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("code = %d, want 503", rec.Code)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test ./internal/health/...
```
Expected: package not found.

- [ ] **Step 3: Implement**

File: `backend/internal/health/handler.go`
```go
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

type response struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// Handler godoc
//
// @Summary      Health check
// @Description  Verifies the service is up and the database is reachable.
// @Tags         meta
// @Produce      json
// @Success      200 {object} health.response
// @Failure      503 {object} httpx.ErrorEnvelope
// @Router       /healthz [get]
func Handler(a *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.Pool.Ping(ctx); err != nil {
			a.Logger.WarnContext(ctx, "healthz: db ping failed", "err", err)
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeInternal, "db unavailable", nil)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, response{Status: "ok", DB: "ok"})
	}
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/health/... -v -timeout=120s
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd casestud/ruecosmetics
git add backend/internal/health/
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(health): /healthz handler with db ping"
```

---

### Task 12: `cmd/api/main.go` — wire everything + graceful shutdown

**Files:**
- Create: `backend/cmd/api/main.go`
- Create: `backend/cmd/api/main_test.go` (smoke test)

**Interfaces:**
- Produces: a binary that boots, listens on `Config.Port`, mounts middleware in order (Recovery → RequestID → CORS), serves `GET /healthz`, and shuts down cleanly on SIGINT/SIGTERM with a 10s drain.
- Consumed by: nothing in Plan 1 (this is the leaf).

- [ ] **Step 1: Write the smoke test**

File: `backend/cmd/api/main_test.go`
```go
package main_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestServerBootsAndHealthzReturnsOK(t *testing.T) {
	ctx := context.Background()
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("ruetest"), postgres.WithUsername("rue"), postgres.WithPassword("rue_dev"),
		testcontainers.WithWaitStrategy(postgres.DefaultWaitStrategy()),
	)
	if err != nil {
		t.Fatalf("pg: %v", err)
	}
	defer pg.Terminate(ctx)
	url, _ := pg.ConnectionString(ctx, "sslmode=disable")

	wd, _ := os.Getwd()
	root := filepath.Join(wd, "..", "..")
	bin := filepath.Join(t.TempDir(), "api")
	build := exec.Command("go", "build", "-o", bin, "./cmd/api")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"PORT=18080",
		"ENV=development",
		"DATABASE_URL="+url,
		"CORS_ORIGINS=http://localhost:5173",
		"LOG_LEVEL=debug",
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	deadline := time.Now().Add(10 * time.Second)
	var resp *http.Response
	for time.Now().Before(deadline) {
		resp, err = http.Get("http://127.0.0.1:18080/healthz")
		if err == nil && resp.StatusCode == 200 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil || resp == nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("healthz code = %d", resp.StatusCode)
		}
		t.Fatalf("healthz never returned 200: %v", err)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
cd casestud/ruecosmetics/backend
go test ./cmd/api/... -timeout=180s
```
Expected: build error — `cmd/api/main.go` doesn't exist.

- [ ] **Step 3: Implement main**

File: `backend/cmd/api/main.go`
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/health"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer a.Close()

	r := chi.NewRouter()
	r.Use(httpx.Recovery(a.Logger))
	r.Use(httpx.RequestID)
	r.Use(httpx.CORS(cfg.CORSOrigins))

	r.Get("/healthz", health.Handler(a))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		a.Logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		a.Logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	shutdownCtx, scancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer scancel()
	return srv.Shutdown(shutdownCtx)
}
```

- [ ] **Step 4: Run smoke test**

```bash
cd casestud/ruecosmetics/backend
go test ./cmd/api/... -v -timeout=180s
```
Expected: PASS — binary builds, boots, `/healthz` returns 200, process killed at end.

- [ ] **Step 5: Verify entire test suite passes**

```bash
go test ./... -timeout=300s
```
Expected: all tests PASS (config + db + httpx + app + health + cmd/api).

- [ ] **Step 6: Commit**

```bash
cd casestud/ruecosmetics
git add backend/cmd/api/
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(api): wire server, middleware, /healthz with graceful shutdown"
```

---

### Task 13: swaggo OpenAPI generation + Makefile + drift check

**Files:**
- Create: `backend/Makefile`
- Create: `casestud/ruecosmetics/Makefile` (top-level passthrough)
- Modify: `backend/cmd/api/main.go` — add general API info comment block
- Create: `backend/docs/docs.go`, `backend/docs/swagger.json`, `backend/docs/swagger.yaml` (generated, committed)
- Create: `backend/scripts/drift-check.sh` (CI helper)

**Interfaces:**
- Produces:
  - `make -C backend openapi` regenerates `backend/docs/*`.
  - `make -C backend drift-check` re-runs generation and fails if the working tree is dirty.
  - `make -C backend test` runs `go test ./...`.
  - `make -C backend dev` runs `air`.
- Consumed by: future plans (Orval in Plan 9 reads `backend/docs/swagger.json`).

- [ ] **Step 1: Install swag CLI**

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
```
Verify: `swag --version`.

- [ ] **Step 2: Add swag annotation block in main.go**

Modify the top of `backend/cmd/api/main.go` — insert above the `package main` line is **not** valid Go; the annotations go directly above `func main()`. Replace the existing `func main()` block with:

```go
// @title           Rue Cosmetics API
// @version         0.1.0
// @description     E-commerce backend for the Rue Cosmetics case study.
// @BasePath        /api/v1
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Generate**

```bash
cd casestud/ruecosmetics/backend
swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency
```
Expected: creates `backend/docs/docs.go`, `backend/docs/swagger.json`, `backend/docs/swagger.yaml`. The `/healthz` route appears in `swagger.json` under `paths./healthz.get`.

Verify:
```bash
grep -c '"/healthz"' docs/swagger.json
```
Expected: `1`.

- [ ] **Step 4: Create `backend/Makefile`**

File: `casestud/ruecosmetics/backend/Makefile`
```make
.PHONY: dev test openapi drift-check sqlc tidy

dev:
	air

test:
	go test ./... -timeout=300s

openapi:
	swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency

drift-check:
	$(MAKE) openapi
	@git diff --exit-code docs/ || (echo "ERROR: openapi docs are stale. Run 'make openapi' and commit." && exit 1)
	@git diff --exit-code internal/db/sqlc/ || (echo "ERROR: sqlc output is stale. Run 'make sqlc' and commit." && exit 1)

sqlc:
	sqlc generate

tidy:
	go mod tidy
```

- [ ] **Step 5: Create top-level Makefile**

File: `casestud/ruecosmetics/Makefile`
```make
.PHONY: dev test openapi drift-check up down

up:
	docker compose up -d postgres mailpit

down:
	docker compose down

dev: up
	$(MAKE) -C backend dev

test:
	$(MAKE) -C backend test

openapi:
	$(MAKE) -C backend openapi

drift-check:
	$(MAKE) -C backend drift-check
```

- [ ] **Step 6: Verify drift check passes**

```bash
cd casestud/ruecosmetics
make drift-check
```
Expected: no output diff; exits 0.

- [ ] **Step 7: Verify drift check fails on edit**

Temporarily edit `backend/docs/swagger.json` (add a space), then:
```bash
make drift-check
```
Expected: prints "ERROR: openapi docs are stale.", exits non-zero. Revert the edit (`git checkout backend/docs/swagger.json`) and re-run to confirm green.

- [ ] **Step 8: Commit**

```bash
cd casestud/ruecosmetics
git add backend/cmd/api/main.go backend/docs/ backend/Makefile Makefile
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "build: openapi generation pipeline with drift check"
```

---

### Task 14: README quickstart

**Files:**
- Create: `casestud/ruecosmetics/README.md`

**Interfaces:**
- Consumed by: humans cloning the repo.

- [ ] **Step 1: Write README**

File: `casestud/ruecosmetics/README.md`
```markdown
# Rue Cosmetics

E-commerce case study. Go backend + React frontend, deployed behind Caddy on Hetzner.

See `docs/superpowers/specs/2026-06-27-rue-cosmetics-design.md` for the design spec, and `docs/superpowers/plans/` for the implementation plans.

## Prerequisites

- Go 1.22+
- Docker (for Postgres and Mailpit)
- `make`
- `sqlc` (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0`)
- `swag` (`go install github.com/swaggo/swag/cmd/swag@v1.16.4`)
- `air` (`go install github.com/air-verse/air@latest`) — for backend hot reload
- `goose` (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Quickstart (backend, Plan 1)

```bash
make up                       # start postgres + mailpit
cp backend/.env.example backend/.env
goose -dir backend/migrations postgres "$(grep DATABASE_URL backend/.env | cut -d= -f2-)" up
make dev                      # air on :8080
curl http://localhost:8080/healthz
# {"status":"ok","db":"ok"}
```

## Tests

```bash
make test                     # runs all Go tests; requires Docker for testcontainers
```

## OpenAPI

```bash
make openapi                  # regenerate backend/docs/{docs.go,swagger.json,swagger.yaml}
make drift-check              # fails CI if openapi or sqlc output is stale
```

## Directory layout

See the spec, Section 3.2.
```

- [ ] **Step 2: Commit**

```bash
cd casestud/ruecosmetics
git add README.md
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "docs: add quickstart README"
```

---

## Verification — end of Plan 1

When all 14 tasks are complete:

- [ ] `make test` exits 0 (all unit + integration + smoke tests pass).
- [ ] `make drift-check` exits 0 on a clean working tree.
- [ ] `make up && make dev` boots the server; `curl :8080/healthz` returns `{"status":"ok","db":"ok"}`.
- [ ] `git log --oneline` shows ~14 atomic commits.
- [ ] No `// TODO` or `// XXX` left in the code outside generated files.

Plan 2 (Catalog + shipping) picks up from this baseline: it adds migrations for `categories`, `brands`, `products`, sqlc queries, repository code, handlers, and the `shipping.Service` that reads `seed/config/shipping_config.json` at startup.

## Self-Review Notes

- **Spec coverage:** This plan implements Section 3.1/3.2 (architecture skeleton), Section 3.3 (local dev), Section 5 cross-cutting (error envelope, CORS, request ID), Section 10.1 (type safety toolchain — OpenAPI pipeline started, drift check active; Orval comes in Plan 9), Section 10.2 (sqlc pipeline established — the grep CI rule for raw SQL is added in Plan 2 once there are services worth grepping), Section 13 (slog, /healthz). All later sections are deferred to subsequent plans by design.
- **Type consistency:** `app.Application` has fields `Config`, `Pool`, `Logger`; all tasks that reference them use those exact names. `db.Pool` is a type alias for `*pgxpool.Pool` (Task 3) and is used consistently. The error codes in `httpx` (Task 6) are referenced by name in Task 7 (`Recovery` uses `CodeInternal`) and Task 11 (`/healthz`).
- **Placeholder scan:** every code block contains complete, runnable code. No "TBD", "TODO", or "similar to" references.
- **One nuance to flag:** the `contains`/`indexOf` helpers in Task 6's `error_test.go` are hand-rolled to avoid importing `strings` twice; replace with `strings.Contains` if a reviewer prefers. The test still passes either way.
