#!/usr/bin/env bash
# AUDiot Monitor — one-command deploy
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Audumla/AUDiotMonitor/main/deploy.sh | bash
#
# Optional environment variables:
#   HWEXP_HOST        — hostname label on all hw_device_* metrics (default: $HOSTNAME)
#   PROMETHEUS_URL    — Prometheus address for the dashboard stack
#                       (default: http://localhost:9090, fine when both stacks run together)
#   INSTALL_DIR       — where to put the compose files (default: ~/audiot)
#   DEPLOY            — "collector", "dashboard", or "both" (default: both)

set -euo pipefail

BASE="https://raw.githubusercontent.com/Audumla/AUDiotMonitor/main"
INSTALL_DIR="${INSTALL_DIR:-$HOME/audiot}"
DEPLOY="${DEPLOY:-both}"

echo "==> AUDiot Monitor deploy"
echo "    Install dir : $INSTALL_DIR"
echo "    Deploying   : $DEPLOY"

mkdir -p "$INSTALL_DIR/collector" "$INSTALL_DIR/dashboard"

# ── Download compose files ────────────────────────────────────────────────

if [[ "$DEPLOY" == "collector" || "$DEPLOY" == "both" ]]; then
    echo "==> Downloading collector stack..."
    curl -fsSL "$BASE/monitoring/collector/docker-compose.yml" \
        -o "$INSTALL_DIR/collector/docker-compose.yml"
fi

if [[ "$DEPLOY" == "dashboard" || "$DEPLOY" == "both" ]]; then
    echo "==> Downloading dashboard stack..."
    curl -fsSL "$BASE/monitoring/dashboard/docker-compose.yml" \
        -o "$INSTALL_DIR/dashboard/docker-compose.yml"
fi

# ── Start stacks ──────────────────────────────────────────────────────────

if [[ "$DEPLOY" == "collector" || "$DEPLOY" == "both" ]]; then
    echo "==> Starting collector stack..."
    HWEXP_HOST="${HWEXP_HOST:-${HOSTNAME:-localhost}}" \
        docker compose -f "$INSTALL_DIR/collector/docker-compose.yml" up -d
fi

if [[ "$DEPLOY" == "dashboard" || "$DEPLOY" == "both" ]]; then
    echo "==> Starting dashboard stack..."
    PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}" \
        docker compose -f "$INSTALL_DIR/dashboard/docker-compose.yml" up -d
fi

# ── Done ──────────────────────────────────────────────────────────────────

echo ""
echo "==> Done!"
if [[ "$DEPLOY" == "collector" || "$DEPLOY" == "both" ]]; then
    echo "    Hardware metrics : http://localhost:9200/metrics"
    echo "    Prometheus       : http://localhost:9090"
fi
if [[ "$DEPLOY" == "dashboard" || "$DEPLOY" == "both" ]]; then
    echo "    Grafana          : http://localhost:3000  (admin / admin)"
fi
echo ""
echo "    To stop:   docker compose -f $INSTALL_DIR/collector/docker-compose.yml down"
echo "               docker compose -f $INSTALL_DIR/dashboard/docker-compose.yml down"
