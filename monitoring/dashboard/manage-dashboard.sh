#!/usr/bin/env bash
# AUDiot Dashboard Manager
#
# Reproducible entry point for installing, updating, validating, and operating
# the host-owned dashboard stack.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${INSTALL_DIR:-/opt/docker/services/dashboard}"

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
    if systemctl --user is-active audiot-kiosk >/dev/null 2>&1; then
        systemctl --user restart audiot-kiosk
        info "Kiosk service restarted via systemd"
    else
        pkill -f "$INSTALL_DIR/kiosk.sh" || true
        pkill -f 'chromium.*audiot-' || true
        nohup "$INSTALL_DIR/kiosk.sh" >/tmp/audiot-kiosk.log 2>&1 </dev/null &
        info "Kiosk restart requested via background process (no systemd user service found)"
    fi
}

list_dashboards() {
    info "Available dashboards (UIDs):"
    find "$INSTALL_DIR/dashboards" -name "*.json" -type f | while read -r path; do
        uid=$(grep -oP '"uid":\s*"\K[^"]+' "$path" | head -1 || true)
        title=$(grep -oP '"title":\s*"\K[^"]+' "$path" | head -1 || true)
        if [ -n "$uid" ]; then
            printf "  %-30s : %s\n" "$uid" "$title"
        fi
    done
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
        if systemctl --user list-unit-files audiot-kiosk.service >/dev/null 2>&1; then
            echo ""
            systemctl --user status audiot-kiosk --no-pager || true
        fi
        ;;
    logs)
        run_in_install_dir docker compose logs --tail=100 grafana
        ;;
    list-dashboards)
        list_dashboards
        ;;
    set-dashboard)
        shift
        run_in_install_dir ./set-dashboard.sh "$@"
        if [[ " $* " == *" --restart "* ]]; then
            restart_kiosk
        fi
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
  ./manage-dashboard.sh list-dashboards
  ./manage-dashboard.sh set-dashboard <uid> [--restart]
EOF
        exit 1
        ;;
esac
