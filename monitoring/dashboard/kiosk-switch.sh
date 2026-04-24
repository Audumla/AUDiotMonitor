#!/usr/bin/env bash
# kiosk-switch.sh — Switch the active Grafana dashboard in the running kiosk
#
# Usage:
#   kiosk-switch.sh                   # list available dashboards
#   kiosk-switch.sh <uid>             # switch to dashboard by UID
#
# Examples:
#   kiosk-switch.sh audiot-triple-gpu-wide
#   kiosk-switch.sh audiot-panel-1080p
#
# Run on the kiosk machine directly or via SSH:
#   ssh <user>@<dashboard-host> /opt/docker/services/dashboard/kiosk-switch.sh audiot-system-overview

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KIOSK_CONFIG_FILE="${KIOSK_CONFIG_FILE:-$SCRIPT_DIR/config/kiosk.env}"
[ -f "$KIOSK_CONFIG_FILE" ] && . "$KIOSK_CONFIG_FILE"

GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
REFRESH="${KIOSK_REFRESH:-30s}"
USER_ID="$(id -u)"
XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$USER_ID}"
WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-0}"
KIOSK_PROFILE_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/chromium-kiosk"

list_dashboards() {
    echo "Available dashboards:"
    curl -sf "$GRAFANA_URL/api/search?query=&type=dash-db" 2>/dev/null \
        | python3 -c "
import sys, json
items = json.load(sys.stdin)
for d in sorted(items, key=lambda x: x.get('title','')):
    print(f\"  {d['uid']:<40} {d['title']}\")
" 2>/dev/null || echo "  (could not reach Grafana at $GRAFANA_URL)"
}

if [ $# -eq 0 ]; then
    list_dashboards
    exit 0
fi

UID_ARG="$1"

# Verify the dashboard exists
if ! curl -sf "$GRAFANA_URL/api/search?query=" 2>/dev/null | grep -qF "\"uid\":\"$UID_ARG\""; then
    echo "Error: dashboard '$UID_ARG' not found in Grafana" >&2
    echo ""
    list_dashboards
    exit 1
fi

URL="$GRAFANA_URL/d/$UID_ARG/$UID_ARG?orgId=1&kiosk&_dash.hideTimePicker=true&from=now-5m&to=now&timezone=browser&refresh=${REFRESH}"

echo "Switching to: $UID_ARG"
echo "URL: $URL"

# Send the new URL to the running Chromium singleton
WAYLAND_DISPLAY="$WAYLAND_DISPLAY" \
XDG_RUNTIME_DIR="$XDG_RUNTIME_DIR" \
    chromium \
    --ozone-platform=wayland \
    --user-data-dir="$KIOSK_PROFILE_DIR" \
    --app="$URL" \
    2>/dev/null &

echo "Done"
