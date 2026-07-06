#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────
# Rue Cosmetics — deploy script
# Usage: bash /opt/rue/deploy.sh
# Requires: /opt/rue/repo  (cloned git repo)
#           /opt/rue/shared/.env  (secrets)
# ──────────────────────────────────────────────

# ── Ensure Go is installed ──
if ! command -v go &>/dev/null; then
  echo "→ Go not found. Installing..."
  GO_VER=$(grep '^go ' /opt/rue/repo/backend/go.mod | awk '{print $2}')
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l)  ARCH="armv6l" ;;
  esac
  wget -q "https://go.dev/dl/go${GO_VER}.linux-${ARCH}.tar.gz" -O /tmp/go.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
  export PATH="$PATH:/usr/local/go/bin"
  echo "   ✓ Go ${GO_VER} installed for ${ARCH}"
fi

# ── Ensure Node.js is installed (for frontend build) ──
if ! command -v node &>/dev/null; then
  echo "→ Node.js not found. Installing LTS via NodeSource..."
  curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
  sudo apt-get install -y nodejs
  echo "   ✓ Node.js $(node -v) installed"
fi

# ── Ensure npm dependencies are cached ──
if [ ! -d /opt/rue/repo/frontend/node_modules ]; then
  echo "→ Installing frontend dependencies..."
  cd /opt/rue/repo/frontend
  npm ci
fi

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
