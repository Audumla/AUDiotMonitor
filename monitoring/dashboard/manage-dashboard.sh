#!/usr/bin/env bash
# AUDiot Dashboard Manager
#
# Reproducible entry point for installing, updating, validating, and operating
# the host-owned dashboard stack.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${INSTALL_DIR:-/opt/docker/dashboard}"

info() { echo "[dashboard-manage] $*"; }
die() { echo "[dashboard-manage] ERROR: $*" >&2; exit 1; }

run_in_install_dir() {
    (cd "$INSTALL_DIR" && "$@")
}

validate_json_tree() {
    python3 - <<'PY' "$INSTALL_DIR"
import json, sys
from pathlib import Path
root = Path(sys.argv[1])
for path in sorted(root.rglob("*.json")):
    json.loads(path.read_text())
    print(f"OK {path}")
PY
}

validate_dashboard_layout() {
    [ -f "$INSTALL_DIR/docker-compose.yml" ] || die "Missing $INSTALL_DIR/docker-compose.yml"
    [ -f "$INSTALL_DIR/config/kiosk.env" ] || die "Missing $INSTALL_DIR/config/kiosk.env"
    validate_json_tree
    if command -v docker >/dev/null 2>&1; then
        run_in_install_dir docker compose config >/dev/null
    fi
    info "Validation passed"
}

restart_kiosk() {
    pkill -f "$INSTALL_DIR/kiosk.sh" || true
    pkill -f 'chromium.*audiot-' || true
    nohup "$INSTALL_DIR/kiosk.sh" >/tmp/audiot-kiosk.log 2>&1 </dev/null &
    info "Kiosk restart requested"
}

case "${1:-}" in
    install)
        INSTALL_DIR="$INSTALL_DIR" "$SCRIPT_DIR/install-layout.sh"
        ;;
    update)
        INSTALL_DIR="$INSTALL_DIR" "$SCRIPT_DIR/install-layout.sh"
        validate_dashboard_layout
        ;;
    validate)
        validate_dashboard_layout
        ;;
    up)
        run_in_install_dir docker compose up -d
        ;;
    restart-grafana)
        run_in_install_dir docker compose restart grafana
        ;;
    restart-kiosk)
        restart_kiosk
        ;;
    status)
        run_in_install_dir docker compose ps
        ;;
    logs)
        run_in_install_dir docker compose logs --tail=100 grafana
        ;;
    set-dashboard)
        shift
        run_in_install_dir ./set-dashboard.sh "$@"
        ;;
    *)
        cat <<'EOF'
Usage:
  ./manage-dashboard.sh install
  ./manage-dashboard.sh update
  ./manage-dashboard.sh validate
  ./manage-dashboard.sh up
  ./manage-dashboard.sh restart-grafana
  ./manage-dashboard.sh restart-kiosk
  ./manage-dashboard.sh status
  ./manage-dashboard.sh logs
  ./manage-dashboard.sh set-dashboard <args...>
EOF
        exit 1
        ;;
esac
