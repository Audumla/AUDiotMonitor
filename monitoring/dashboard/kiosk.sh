#!/usr/bin/env bash
# AUDiot Kiosk Launcher
#
# Detects the connected screen resolution and opens the best-matching dashboard
# in Chromium fullscreen/kiosk mode. Restarts Chromium automatically on exit.
#
# Usage:
#   ./kiosk.sh [grafana-url]
#
# Environment variables (all optional):
#   GRAFANA_URL      Base URL of Grafana  (default: http://localhost:3000)
#   KIOSK_REFRESH    Dashboard auto-refresh interval  (default: 30s)
#   KIOSK_DASHBOARD  Force a specific dashboard UID — skips auto-detection
#
# Dashboard selection by resolution:
#   Width ≥ 1800, height ≤ 500   →  audiot-panel-1920x440  (ticker / ultra-wide strip)
#   Portrait  (height > width)   →  audiot-panel-portrait
#   Narrow    (width ≤ 800)      →  audiot-panel-portrait
#   Width ≥ 1920, height ≥ 1000  →  audiot-panel-1080p
#   All other                    →  audiot-system-overview

set -euo pipefail

GRAFANA_URL="${GRAFANA_URL:-${1:-http://localhost:3000}}"
REFRESH="${KIOSK_REFRESH:-30s}"

# ── Detect resolution ────────────────────────────────────────────────────────

detect_resolution() {
    # Method 1: xrandr (X11 — most accurate, requires $DISPLAY)
    if command -v xrandr &>/dev/null && [ -n "${DISPLAY:-}" ]; then
        local res
        res=$(xrandr --current 2>/dev/null \
              | awk '/ connected/{p=1} p && /[0-9]+x[0-9]+\*/{match($0,/([0-9]+x[0-9]+)/,a); print a[0]; p=0}' \
              | head -1)
        [ -n "$res" ] && { echo "$res"; return; }
    fi

    # Method 2: Wayland via wlr-randr (Raspberry Pi OS Bookworm / labwc)
    if command -v wlr-randr &>/dev/null; then
        local res
        res=$(wlr-randr 2>/dev/null \
              | awk '/current/{match($0,/([0-9]+x[0-9]+)/,a); print a[0]}' \
              | head -1)
        [ -n "$res" ] && { echo "$res"; return; }
    fi

    # Method 3: framebuffer sysfs — works without X or Wayland, always available on Pi
    if [ -r /sys/class/graphics/fb0/virtual_size ]; then
        tr ',' 'x' < /sys/class/graphics/fb0/virtual_size
        return
    fi

    # Fallback
    echo "1920x1080"
}

# ── Select dashboard UID by resolution ───────────────────────────────────────

select_dashboard() {
    local res="$1"
    local w h
    w=$(echo "$res" | cut -dx -f1)
    h=$(echo "$res" | cut -dx -f2)

    # Ultra-wide / ticker strip  (e.g. 1920×440, 3440×440)
    if [ "$w" -ge 1800 ] && [ "$h" -le 500 ]; then
        echo "audiot-panel-1920x440"
    # Portrait or narrow  (height > width, or width ≤ 800)
    elif [ "$h" -gt "$w" ] || [ "$w" -le 800 ]; then
        echo "audiot-panel-portrait"
    # Standard 1080p and up
    elif [ "$w" -ge 1920 ] && [ "$h" -ge 1000 ]; then
        echo "audiot-panel-1080p"
    # Smaller landscape (e.g. 1366×768, 1280×800)
    else
        echo "audiot-system-overview"
    fi
}

# ── Wait for Grafana to respond ───────────────────────────────────────────────

wait_for_grafana() {
    echo "[kiosk] Waiting for Grafana at $GRAFANA_URL ..."
    local i=0
    while ! curl -sf "$GRAFANA_URL/api/health" > /dev/null 2>&1; do
        i=$((i + 1))
        [ "$i" -ge 72 ] && { echo "[kiosk] Grafana did not respond after 6 min — launching anyway"; return; }
        sleep 5
    done
    echo "[kiosk] Grafana ready"
}

# ── Main ──────────────────────────────────────────────────────────────────────

if [ -n "${KIOSK_DASHBOARD:-}" ]; then
    DASHBOARD_UID="$KIOSK_DASHBOARD"
    echo "[kiosk] Using forced dashboard: $DASHBOARD_UID"
else
    RES=$(detect_resolution)
    DASHBOARD_UID=$(select_dashboard "$RES")
    echo "[kiosk] Detected resolution: $RES  →  dashboard: $DASHBOARD_UID"
fi

KIOSK_URL="$GRAFANA_URL/d/$DASHBOARD_UID?kiosk&refresh=$REFRESH"
echo "[kiosk] URL: $KIOSK_URL"

wait_for_grafana

# Disable screen blanking and power management
xset s off   2>/dev/null || true
xset -dpms   2>/dev/null || true
xset s noblank 2>/dev/null || true

# Find Chromium binary (Debian names it chromium or chromium-browser)
CHROMIUM=""
for bin in chromium-browser chromium google-chrome; do
    command -v "$bin" &>/dev/null && CHROMIUM="$bin" && break
done

if [ -z "$CHROMIUM" ]; then
    echo "[kiosk] ERROR: chromium not found. Install with: sudo apt install chromium-browser"
    exit 1
fi

echo "[kiosk] Launching $CHROMIUM in kiosk mode"

# Remove leftover crash flags that prevent kiosk from starting cleanly
rm -f ~/.config/chromium/Default/Preferences 2>/dev/null || true

# Launch loop — Chromium is restarted automatically on exit
while true; do
    "$CHROMIUM" \
        --kiosk \
        --noerrdialogs \
        --disable-infobars \
        --disable-session-crashed-bubble \
        --disable-restore-session-state \
        --disable-pinch \
        --overscroll-history-navigation=0 \
        --check-for-update-interval=604800 \
        --app="$KIOSK_URL" \
        2>/dev/null || true

    echo "[kiosk] Chromium exited — restarting in 5s"
    sleep 5
done
