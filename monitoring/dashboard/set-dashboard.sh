#!/usr/bin/env bash
# AUDiot Dashboard Selector
#
# Updates config/kiosk.env so the kiosk uses a chosen dashboard UID.
#
# Usage:
#   ./set-dashboard.sh list
#   ./set-dashboard.sh force <uid>
#   ./set-dashboard.sh clear-force
#   ./set-dashboard.sh set ultrawide <uid>
#   ./set-dashboard.sh set portrait <uid>
#   ./set-dashboard.sh set 1080p <uid>
#   ./set-dashboard.sh set landscape <uid>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${KIOSK_CONFIG_FILE:-$SCRIPT_DIR/config/kiosk.env}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"

info() { echo "[dashboard] $*"; }
die() { echo "[dashboard] ERROR: $*" >&2; exit 1; }

restart_kiosk() {
    if systemctl --user is-active audiot-kiosk >/dev/null 2>&1; then
        systemctl --user restart audiot-kiosk
        info "Kiosk service restarted via systemd."
    else
        info "No systemd user service found."
        pkill -f "kiosk.sh" || true
        pkill -f 'chromium.*audiot-' || true
        info "Kiosk process stopped."
        info "To restart it, please run './kiosk-install.sh' on the host machine."
    fi
}

[ -f "$CONFIG_FILE" ] || die "Missing config file: $CONFIG_FILE"

list_dashboards() {
    if command -v jq >/dev/null 2>&1; then
        curl -sf "$GRAFANA_URL/api/search?query=" \
            | jq -r '.[] | select(.type=="dash-db") | "\(.uid)\t\(.title)"'
    else
        curl -sf "$GRAFANA_URL/api/search?query=" \
            | sed 's/},{/}\n{/g' \
            | grep '"type":"dash-db"' \
            | sed -n 's/.*"uid":"\([^"]*\)".*"title":"\([^"]*\)".*/\1\t\2/p'
    fi
}

dashboard_exists() {
    local uid="$1"
    list_dashboards | awk -F '\t' -v uid="$uid" '$1 == uid { found=1 } END { exit(found ? 0 : 1) }'
}

set_key() {
    local key="$1"
    local value="$2"
    if grep -q "^${key}=" "$CONFIG_FILE"; then
        sed -i "s|^${key}=.*|${key}=${value}|" "$CONFIG_FILE"
    else
        printf '%s=%s\n' "$key" "$value" >> "$CONFIG_FILE"
    fi
}

case "${1:-}" in
    list)
        list_dashboards
        ;;
    force)
        uid="${2:-}"
        [ -n "$uid" ] || die "Usage: ./set-dashboard.sh force <uid> [--restart]"
        dashboard_exists "$uid" || die "Dashboard UID not found in Grafana: $uid"
        set_key "KIOSK_DASHBOARD" "$uid"
        info "Forced dashboard set to $uid"
        [[ " $* " == *" --restart "* ]] && restart_kiosk
        ;;
    clear-force)
        set_key "KIOSK_DASHBOARD" ""
        info "Forced dashboard cleared"
        [[ " $* " == *" --restart "* ]] && restart_kiosk
        ;;
    set)
        class="${2:-}"
        uid="${3:-}"
        [ -n "$class" ] && [ -n "$uid" ] || die "Usage: ./set-dashboard.sh set <ultrawide|portrait|1080p|landscape> <uid> [--restart]"
        dashboard_exists "$uid" || die "Dashboard UID not found in Grafana: $uid"
        case "$class" in
            ultrawide) key="KIOSK_DASHBOARD_ULTRAWIDE" ;;
            portrait) key="KIOSK_DASHBOARD_PORTRAIT" ;;
            1080p) key="KIOSK_DASHBOARD_1080P" ;;
            landscape) key="KIOSK_DASHBOARD_LANDSCAPE" ;;
            *) die "Unknown class: $class" ;;
        esac
        set_key "$key" "$uid"
        info "$key set to $uid"
        [[ " $* " == *" --restart "* ]] && restart_kiosk
        ;;
    *)
        cat <<'EOF'
Usage:
  ./set-dashboard.sh list
  ./set-dashboard.sh force <uid> [--restart]
  ./set-dashboard.sh clear-force [--restart]
  ./set-dashboard.sh set <class> <uid> [--restart]
EOF
        exit 1
        ;;
esac
