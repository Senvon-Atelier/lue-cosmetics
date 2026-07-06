#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────
# Rue Cosmetics — deploy script
# Usage: bash /opt/rue/deploy.sh
# Requires: /opt/rue/repo  (cloned git repo)
#           /opt/rue/shared/.env  (secrets)
# ──────────────────────────────────────────────

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RELEASE="/opt/rue/releases/$TIMESTAMP"
REPO="/opt/rue/repo"
SHARED="/opt/rue/shared"

echo "→ Creating release directory: $TIMESTAMP"
mkdir -p "$RELEASE"

# ── Build Go binary ──
echo "→ Building Go binary..."
cd "$REPO/backend"
go build -o "$RELEASE/api" ./cmd/api
echo "   ✓ binary: $RELEASE/api"

# ── Build frontend ──
echo "→ Building frontend..."
cd "$REPO/frontend"
npm ci --silent
npm run build --silent
cp -r dist "$RELEASE/frontend"
echo "   ✓ frontend: $RELEASE/frontend"

# ── Link env ──
ln -sf "$SHARED/.env" "$RELEASE/.env"
echo "   ✓ .env linked"

# ── Copy shipping config if it exists ──
if [ -f "$REPO/backend/config/shipping_config.json" ]; then
  cp "$REPO/backend/config/shipping_config.json" "$RELEASE/shipping_config.json"
  echo "   ✓ shipping config copied"
fi

# ── Swap current / previous ──
echo "→ Swapping symlinks..."
[ -L /opt/rue/current ] && mv /opt/rue/current /opt/rue/previous || true
ln -sfn "$RELEASE" /opt/rue/current

# ── Restart service ──
echo "→ Restarting rue-api service..."
systemctl daemon-reload
systemctl restart rue-api
echo "   ✓ service restarted"

# ── Prune old releases (keep last 5) ──
echo "→ Pruning old releases (keeping last 5)..."
ls -t /opt/rue/releases/ 2>/dev/null | tail -n +6 | while read -r old; do
  rm -rf "/opt/rue/releases/$old"
  echo "   ✗ removed: $old"
done

echo ""
echo "──────────────────────────────────────────────"
echo "✅  Deploy complete: $TIMESTAMP"
echo "   current → $RELEASE"
echo "   rollback target → $(readlink /opt/rue/previous 2>/dev/null || echo 'none')"
echo "──────────────────────────────────────────────"
