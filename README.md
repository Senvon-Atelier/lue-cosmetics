# Rue Cosmetics

E-commerce case study. Go backend + React frontend, deployed behind Caddy on Hetzner.

See `docs/superpowers/specs/2026-06-27-rue-cosmetics-design.md` for the design spec, and `docs/superpowers/plans/` for the implementation plans.

## Prerequisites

- Go 1.25.8+
- Node 20+ with pnpm 9+
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

## Quickstart (frontend)

```bash
cd frontend
pnpm install
pnpm dev                      # Vite on :5173, proxies /api/* to :8080
```

Regenerate the API client after backend contract changes:

```bash
make openapi && cd frontend && pnpm orval
```

## Tests

```bash
make test                     # runs all Go tests; requires Docker for testcontainers
make seed-run                 # requires DATABASE_URL exported in the shell (does not read backend/.env)
cd frontend && pnpm vitest run
```

## OpenAPI

```bash
make openapi                  # regenerate backend/docs/{docs.go,swagger.json,swagger.yaml}
make drift-check              # fails CI if openapi or sqlc output is stale
```

## Directory layout

See the spec, Section 3.2.
