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
#   KIOSK_CONFIG_FILE        Config file path  (default: ./config/kiosk.env)
#   GRAFANA_URL              Base URL of Grafana  (default: http://localhost:3000)
#   KIOSK_REFRESH            Dashboard auto-refresh interval  (default: 30s)
#   KIOSK_DASHBOARD          Force a specific dashboard UID — skips auto-detection
#   KIOSK_DASHBOARD_DEFAULT  Default dashboard UID when no screen override is set
#   KIOSK_DASHBOARD_ULTRAWIDE / _PORTRAIT / _1080P / _LANDSCAPE
#                           Optional per-screen-class overrides
#   KIOSK_DASHBOARD_FALLBACKS Comma-separated fallback UIDs checked in order
#
# Dashboard selection by resolution:
#   Width ≥ 1800, height ≤ 500   →  audiot-panel-1920x440  (ticker / ultra-wide strip)
#   Portrait  (height > width)   →  audiot-panel-portrait
#   Narrow    (width ≤ 800)      →  audiot-panel-portrait
#   Width ≥ 1920, height ≥ 1000  →  audiot-panel-1080p
#   All other                    →  audiot-system-overview

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KIOSK_CONFIG_FILE="${KIOSK_CONFIG_FILE:-$SCRIPT_DIR/config/kiosk.env}"

if [ -f "$KIOSK_CONFIG_FILE" ]; then
    # shellcheck disable=SC1090
    . "$KIOSK_CONFIG_FILE"
fi

GRAFANA_URL="${GRAFANA_URL:-${1:-http://localhost:3000}}"
REFRESH="${KIOSK_REFRESH:-30s}"
USER_ID="$(id -u)"

# Fill in common desktop session variables when kiosk.sh is launched from a
# stripped-down autostart environment.
XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$USER_ID}"
if [ -z "${WAYLAND_DISPLAY:-}" ] && [ -S "$XDG_RUNTIME_DIR/wayland-0" ]; then
    WAYLAND_DISPLAY="wayland-0"
fi
if [ -z "${DBUS_SESSION_BUS_ADDRESS:-}" ] && [ -S "$XDG_RUNTIME_DIR/bus" ]; then
    DBUS_SESSION_BUS_ADDRESS="unix:path=$XDG_RUNTIME_DIR/bus"
fi
if [ -z "${XAUTHORITY:-}" ] && [ -f "$HOME/.Xauthority" ]; then
    XAUTHORITY="$HOME/.Xauthority"
fi

export XDG_RUNTIME_DIR
[ -n "${WAYLAND_DISPLAY:-}" ] && export WAYLAND_DISPLAY
[ -n "${DBUS_SESSION_BUS_ADDRESS:-}" ] && export DBUS_SESSION_BUS_ADDRESS
[ -n "${XAUTHORITY:-}" ] && export XAUTHORITY

# ── Detect resolution ────────────────────────────────────────────────────────

detect_resolution() {
    # Method 1: xrandr (X11 — most accurate, requires $DISPLAY)
    if command -v xrandr &>/dev/null && [ -n "${DISPLAY:-}" ]; then
        local res
        res=$(xrandr --current 2>/dev/null \
              | awk '/ connected/{p=1} p && /[0-9]+x[0-9]+\*/{if (match($0, /[0-9]+x[0-9]+/)) { print substr($0, RSTART, RLENGTH); p=0 }}' \
              | head -1)
        [ -n "$res" ] && { echo "$res"; return; }
    fi

    # Method 2: Wayland via wlr-randr (Raspberry Pi OS Bookworm / labwc)
    if command -v wlr-randr &>/dev/null; then
        local res
        res=$(wlr-randr 2>/dev/null \
              | awk '/current/{if (match($0, /[0-9]+x[0-9]+/)) print substr($0, RSTART, RLENGTH)}' \
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

screen_class_for_resolution() {
    local res="$1"
    local w h
    w=$(echo "$res" | cut -dx -f1)
    h=$(echo "$res" | cut -dx -f2)

    if [ "$w" -ge 1800 ] && [ "$h" -le 500 ]; then
        echo "ultrawide"
    elif [ "$h" -gt "$w" ] || [ "$w" -le 800 ]; then
        echo "portrait"
    elif [ "$w" -ge 1920 ] && [ "$h" -ge 1000 ]; then
        echo "1080p"
    else
        echo "landscape"
    fi
}

select_dashboard() {
    local res="$1"
    local screen_class

    screen_class="$(screen_class_for_resolution "$res")"

    case "$screen_class" in
        ultrawide)
            echo "${KIOSK_DASHBOARD_ULTRAWIDE:-${KIOSK_DASHBOARD_DEFAULT:-audiot-panel-display}}"
            ;;
        portrait)
            echo "${KIOSK_DASHBOARD_PORTRAIT:-${KIOSK_DASHBOARD_DEFAULT:-audiot-panel-portrait}}"
            ;;
        1080p)
            echo "${KIOSK_DASHBOARD_1080P:-${KIOSK_DASHBOARD_DEFAULT:-audiot-panel-1080p}}"
            ;;
        *)
            echo "${KIOSK_DASHBOARD_LANDSCAPE:-${KIOSK_DASHBOARD_DEFAULT:-audiot-system-overview}}"
            ;;
    esac
}

# Return success if the dashboard UID is present in Grafana's search index.
dashboard_exists() {
    local uid="$1"
    curl -sf "$GRAFANA_URL/api/search?query=" 2>/dev/null \
        | grep -F "\"uid\":\"$uid\"" > /dev/null
}

try_fallback_dashboards() {
    local fallback uid
    IFS=',' read -r -a fallback <<< "${KIOSK_DASHBOARD_FALLBACKS:-audiot-panel-display,audiot-system-overview}"
    for uid in "${fallback[@]}"; do
        uid="$(echo "$uid" | xargs)"
        [ -z "$uid" ] && continue
        if dashboard_exists "$uid"; then
            echo "$uid"
            return
        fi
    done
}

copy_example_dashboard_if_missing() {
    local example_dir="$SCRIPT_DIR/examples/custom"
    local custom_dir="$SCRIPT_DIR/dashboards/custom"

    [ -d "$example_dir" ] || return 0
    mkdir -p "$custom_dir"

    while IFS= read -r src; do
        local name dst
        name="$(basename "$src")"
        dst="$custom_dir/$name"
        [ -f "$dst" ] || cp "$src" "$dst"
    done < <(find "$example_dir" -maxdepth 1 -type f -name '*.json' | sort)
}

ensure_supporting_layout() {
    copy_example_dashboard_if_missing
}

# Ensure the chosen UID exists; otherwise fall back to a known-good option.
resolve_dashboard_uid() {
    local preferred="$1"

    if dashboard_exists "$preferred"; then
        echo "$preferred"
        return
    fi

    if [ -n "${KIOSK_DASHBOARD_DEFAULT:-}" ] && dashboard_exists "$KIOSK_DASHBOARD_DEFAULT"; then
        echo "$KIOSK_DASHBOARD_DEFAULT"
        return
    fi

    if try_fallback_dashboards; then
        return
    fi

    echo "$preferred"
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
    DASHBOARD_UID="$(select_dashboard "$RES")"
    echo "[kiosk] Detected resolution: $RES"
fi

wait_for_grafana
ensure_supporting_layout
DASHBOARD_UID="$(resolve_dashboard_uid "$DASHBOARD_UID")"

if [ -z "${KIOSK_DASHBOARD:-}" ]; then
    echo "[kiosk] Detected resolution: $RES  →  dashboard: $DASHBOARD_UID"
fi

# The user's provided URL is the ground truth. The slug is the same as the UID.
# Parameters:
# - kiosk: Hides the side navigation panel.
# - _dash.hideTimePicker=true: Hides the top time picker and refresh bar.
KIOSK_URL="$GRAFANA_URL/d/$DASHBOARD_UID/$DASHBOARD_UID?orgId=1&kiosk&_dash.hideTimePicker=true&from=now-5m&to=now"
echo "[kiosk] URL: $KIOSK_URL"

# Configure screen blanking and power management
if [ -n "${KIOSK_IDLE_TIMEOUT:-}" ] && [ "${KIOSK_IDLE_TIMEOUT}" -gt 0 ] 2>/dev/null; then
    echo "[kiosk] Enabling screen blanking (timeout: ${KIOSK_IDLE_TIMEOUT}s)"
    xset s on 2>/dev/null || true
    xset +dpms 2>/dev/null || true
    xset s blank 2>/dev/null || true
    xset dpms "${KIOSK_IDLE_TIMEOUT}" "${KIOSK_IDLE_TIMEOUT}" "${KIOSK_IDLE_TIMEOUT}" 2>/dev/null || true
else
    echo "[kiosk] Disabling screen blanking (always on)"
    xset s off   2>/dev/null || true
    xset -dpms   2>/dev/null || true
    xset s noblank 2>/dev/null || true
fi

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

# Use a dedicated profile so kiosk mode does not reuse or fight with a
# desktop Chromium session.
KIOSK_PROFILE_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/chromium-kiosk"
mkdir -p "$KIOSK_PROFILE_DIR"

# Clear stale Chromium single-instance markers in the kiosk profile.
if [ -L "$KIOSK_PROFILE_DIR/SingletonLock" ]; then
    lock_target=$(readlink "$KIOSK_PROFILE_DIR/SingletonLock" 2>/dev/null || true)
    lock_pid="${lock_target##*-}"
    if ! [ -n "$lock_pid" ] || ! [ "$lock_pid" -eq "$lock_pid" ] 2>/dev/null || ! kill -0 "$lock_pid" 2>/dev/null; then
        rm -f \
            "$KIOSK_PROFILE_DIR/SingletonLock" \
            "$KIOSK_PROFILE_DIR/SingletonCookie" \
            "$KIOSK_PROFILE_DIR/SingletonSocket"
    fi
fi

# Remove leftover crash flags that prevent kiosk from starting cleanly
rm -f "$KIOSK_PROFILE_DIR/Default/Preferences" 2>/dev/null || true

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
        --no-first-run \
        --no-default-browser-check \
        --password-store=basic \
        --user-data-dir="$KIOSK_PROFILE_DIR" \
        --app="$KIOSK_URL" \
        2>/dev/null || true

    echo "[kiosk] Chromium exited — restarting in 5s"
    sleep 5
done
