# Rue Cosmetics — Deploy Guide

## Server

**Hetzner CAX11** — 4 GB RAM, 2 vCPU, 40 GB SSD.

Services already running on this machine:
- Caddy (reverse proxy)
- Docker (with Postgres + Mailpit)
- Dockge, Glances, Uptime Kuma, Bugsink

---

## One-Time Setup

### 1. Clone the repo

```bash
mkdir -p /opt/rue/{releases,shared,repo}
git clone <repo-url> /opt/rue/repo
```

### 2. Configure environment

```bash
cp /opt/rue/repo/deploy/.env.prod /opt/rue/shared/.env
nano /opt/rue/shared/.env
```

Fill in at minimum:
- `DATABASE_URL` — postgres connection string
- `CORS_ORIGINS` — your frontend domain
- `FRONTEND_BASE_URL` — same as above
- (Paystack, Resend, Google OAuth are optional)

### 3. Set up systemd service

```bash
cp /opt/rue/repo/deploy/rue-api.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable rue-api
```

### 4. Install Caddy snippet

Append contents of `deploy/Caddyfile.rue` to your existing Caddyfile, then:

```bash
caddy reload
```

### 5. Ensure Docker services are running

```bash
docker compose up -d postgres mailpit   # or just 'postgres' if Mailpit isn't needed
```

---

## Deploying

```bash
cd /opt/rue/repo
git pull
bash /opt/rue/deploy.sh
```

The script will:
1. Build the Go binary
2. Build the frontend (`npm ci && npm run build`)
3. Link the `.env` file
4. Swap `current` → `previous`, new release → `current`
5. Restart the systemd service (migrations run automatically on startup)
6. Prune old releases (keeps last 5)

### Verify

```bash
systemctl status rue-api
journalctl -u rue-api -n 20 --no-pager
curl http://localhost:8080/healthz
```

---

## Rolling Back

```bash
# Check what we'd roll back to
readlink /opt/rue/previous

# Rollback
systemctl stop rue-api
ln -sfn $(readlink /opt/rue/previous) /opt/rue/current
systemctl start rue-api

# Verify
systemctl status rue-api
```

**Important:** Do NOT roll back Postgres migrations. The schema is designed to be forward-only (additive changes only). The old binary simply ignores columns it doesn't know about.

---

## Directory Layout

```
/opt/rue/
├── releases/
│   ├── 20260706-120000/         # each deploy = timestamped dir
│   │   ├── api                  # Go binary
│   │   ├── frontend/            # built frontend (dist/)
│   │   └── .env → ../../shared/.env
│   ├── 20260705-100000/
│   └── 20260704-080000/
├── current → releases/20260706-120000/   # symlink (active)
├── previous → releases/20260705-100000/  # symlink (rollback target)
├── shared/
│   └── .env                     # real secrets, never overwritten
├── repo/                        # git clone
│   ├── backend/
│   └── frontend/
├── deploy.sh                    # copy of deploy/deploy.sh
└── rue-api.service              # systemd unit (symlinked or copied)
```

---

## Postgres — Migrations are Additive Only

| ✅ Allowed | ❌ Not Allowed (without phased deploy) |
|---|---|
| `ADD COLUMN` | `DROP COLUMN` |
| `CREATE TABLE` | `DROP TABLE` |
| `CREATE INDEX` | `ALTER COLUMN` (rename/change type) |
| `ALTER TABLE ADD CONSTRAINT` | renaming anything |

This guarantees that a rollback never breaks — the old binary just ignores new columns/tables.

---

## Useful Commands

```bash
# View recent logs
journalctl -u rue-api -n 50 -f

# Check if the API is responding
curl -s http://localhost:8080/healthz

# Manually restart
systemctl restart rue-api

# List all releases
ls -l /opt/rue/releases/

# Check disk usage of releases
du -sh /opt/rue/releases/*
```
