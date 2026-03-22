#!/usr/bin/env bash
# AUDiot Kiosk Launcher
#
# Detects the connected screen resolution and opens the best-matching dashboard
# in Chromium fullscreen/kiosk mode. Restarts Chromium automatically on exit.
#
# Usage:
#   ./kiosk.sh [grafana-url]

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

# Environment recovery for display access
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

# ── Helpers ───────────────────────────────────────────────────────────────────

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [kiosk] $*"
}

# ── Resolution Detection ──────────────────────────────────────────────────────

detect_resolution() {
    if command -v xrandr &>/dev/null && [ -n "${DISPLAY:-}" ]; then
        local res
        res=$(xrandr --current 2>/dev/null | awk '/ connected/{p=1} p && /[0-9]+x[0-9]+\*/{if (match($0, /[0-9]+x[0-9]+/)) { print substr($0, RSTART, RLENGTH); p=0 }}' | head -1)
        [ -n "$res" ] && { echo "$res"; return; }
    fi
    if command -v wlr-randr &>/dev/null; then
        local res
        res=$(wlr-randr 2>/dev/null | awk '/current/{if (match($0, /[0-9]+x[0-9]+/)) print substr($0, RSTART, RLENGTH)}' | head -1)
        [ -n "$res" ] && { echo "$res"; return; }
    fi
    if [ -r /sys/class/graphics/fb0/virtual_size ]; then
        tr ',' 'x' < /sys/class/graphics/fb0/virtual_size
        return
    fi
    echo "1920x1080"
}

# ── Dashboard Selection ───────────────────────────────────────────────────────

screen_class_for_resolution() {
    local res="$1"
    local w h
    w=$(echo "$res" | cut -dx -f1)
    h=$(echo "$res" | cut -dx -f2)
    if [ "$w" -ge 1800 ] && [ "$h" -le 500 ]; then echo "ultrawide"
    elif [ "$h" -gt "$w" ] || [ "$w" -le 800 ]; then echo "portrait"
    elif [ "$w" -ge 1920 ] && [ "$h" -ge 1000 ]; then echo "1080p"
    else echo "landscape"; fi
}

select_dashboard() {
    local res="$1"
    local class=$(screen_class_for_resolution "$res")
    case "$class" in
        ultrawide) echo "${KIOSK_DASHBOARD_ULTRAWIDE:-${KIOSK_DASHBOARD_DEFAULT:-audiot-panel-display}}" ;;
        portrait)  echo "${KIOSK_DASHBOARD_PORTRAIT:-${KIOSK_DASHBOARD_DEFAULT:-audiot-panel-portrait}}" ;;
        1080p)     echo "${KIOSK_DASHBOARD_1080P:-${KIOSK_DASHBOARD_DEFAULT:-audiot-panel-1080p}}" ;;
        *)         echo "${KIOSK_DASHBOARD_LANDSCAPE:-${KIOSK_DASHBOARD_DEFAULT:-audiot-system-overview}}" ;;
    esac
}

dashboard_exists() {
    local uid="$1"
    curl -sf "$GRAFANA_URL/api/search?query=" 2>/dev/null | grep -F "\"uid\":\"$uid\"" > /dev/null
}

try_fallback_dashboards() {
    local fallback uid
    IFS=',' read -r -a fallback <<< "${KIOSK_DASHBOARD_FALLBACKS:-audiot-panel-display,audiot-system-overview}"
    for uid in "${fallback[@]}"; do
        uid="$(echo "$uid" | xargs)"
        [ -z "$uid" ] && continue
        dashboard_exists "$uid" && { echo "$uid"; return 0; }
    done
    return 1
}

resolve_dashboard_uid() {
    local preferred="$1"
    dashboard_exists "$preferred" && { echo "$preferred"; return; }
    [ -n "${KIOSK_DASHBOARD_DEFAULT:-}" ] && dashboard_exists "$KIOSK_DASHBOARD_DEFAULT" && { echo "$KIOSK_DASHBOARD_DEFAULT"; return; }
    try_fallback_dashboards && return
    echo "$preferred"
}

# ── Main ──────────────────────────────────────────────────────────────────────

if [ -n "${KIOSK_DASHBOARD:-}" ]; then
    DASHBOARD_UID="$KIOSK_DASHBOARD"
    log "Using forced dashboard: $DASHBOARD_UID"
else
    RES=$(detect_resolution)
    DASHBOARD_UID=$(select_dashboard "$RES")
    DASHBOARD_UID=$(resolve_dashboard_uid "$DASHBOARD_UID")
    log "Detected resolution: $RES  →  dashboard: $DASHBOARD_UID"
fi

# Build kiosk URL — uses $REFRESH from kiosk.env (default 5s)
KIOSK_URL="$GRAFANA_URL/d/$DASHBOARD_UID/$DASHBOARD_UID?orgId=1&kiosk&_dash.hideTimePicker=true&from=now-5m&to=now&timezone=browser&refresh=${REFRESH:-5s}"
log "URL: $KIOSK_URL"

# ── Screen power management (DPMS) ────────────────────────────────────────────
# Works on Wayland via wlopm (direct output power control).
# Touch/input events on /dev/input are monitored directly so the display wakes
# even when the Wayland compositor does not route touch through idle-notify.
#
# KIOSK_DPMS_STANDBY  : seconds of inactivity before display powers off (dim)
# KIOSK_DPMS_OFF      : seconds of inactivity before display fully powers off
# KIOSK_IDLE_TIMEOUT  : shorthand when standby == off (0 = always on)
# KIOSK_TOUCH_DEVICE  : override input device path (auto-detected if blank)

find_input_device() {
    # Print the first readable touch/pointer input device path; always returns 0.
    # (Caller checks for empty output — returning non-zero would trigger set -e.)
    for sysname in /sys/class/input/event*/device/name; do
        [ -f "$sysname" ] || continue
        name=$(cat "$sysname" 2>/dev/null)
        case "$name" in
            *[Tt]ouch*|*TOUCH*|ILITEK*|eGalax*|Goodix*|Wacom*|wacom*|Mouse*|mouse*|pointer*)
                ev="${sysname%/device/name}"
                dev="/dev/input/${ev##*/}"
                [ -r "$dev" ] && echo "$dev" && return 0
                ;;
        esac
    done
    return 0
}

_dpms_enabled=false
if [ -n "${KIOSK_IDLE_TIMEOUT:-}" ] && [ "${KIOSK_IDLE_TIMEOUT}" -gt 0 ] 2>/dev/null; then
    _dpms_enabled=true
elif [ -n "${KIOSK_DPMS_STANDBY:-}${KIOSK_DPMS_OFF:-}" ]; then
    _dpms_enabled=true
fi

pkill -f swayidle 2>/dev/null || true
pkill -f 'dd if=/dev/input' 2>/dev/null || true

_DPMS_LOG="/tmp/kiosk-dpms-${USER}.log"

if [ "$_dpms_enabled" = "true" ]; then
    dim_t="${KIOSK_DPMS_STANDBY:-${KIOSK_IDLE_TIMEOUT:-600}}"
    off_t="${KIOSK_DPMS_OFF:-${KIOSK_IDLE_TIMEOUT:-1800}}"
    log "Screen power management: dim=${dim_t}s off=${off_t}s"

    INPUT_DEV="${KIOSK_TOUCH_DEVICE:-$(find_input_device 2>/dev/null)}"

    if [ -n "$INPUT_DEV" ] && [ -r "$INPUT_DEV" ] && command -v wlopm &>/dev/null; then
        log "Input-DPMS: watching $INPUT_DEV (off after ${off_t}s, wake on touch)"
        # Direct input monitor: polls device in 5s windows to track idle time.
        # Turns display off after off_t seconds of no input; wakes immediately on touch.
        # WAYLAND_DISPLAY and XDG_RUNTIME_DIR are hardcoded at fork time for reliability.
        _DPMS_WD="${WAYLAND_DISPLAY:-wayland-0}"
        _DPMS_XR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"
        _DPMS_DEV="$INPUT_DEV"
        _DPMS_DIM="$dim_t"
        _DPMS_OFF="$off_t"
        (
            _state=on
            _last=$(date +%s)
            _wlopm_on()  {
                WAYLAND_DISPLAY="$_DPMS_WD" XDG_RUNTIME_DIR="$_DPMS_XR" wlopm --on  '*' 2>/dev/null || true
                echo "$(date '+%Y-%m-%d %H:%M:%S') display ON" >> "$_DPMS_LOG"
            }
            _wlopm_off() {
                WAYLAND_DISPLAY="$_DPMS_WD" XDG_RUNTIME_DIR="$_DPMS_XR" wlopm --off '*' 2>/dev/null || true
                echo "$(date '+%Y-%m-%d %H:%M:%S') display OFF" >> "$_DPMS_LOG"
            }
            echo "$(date '+%Y-%m-%d %H:%M:%S') Input-DPMS started: dev=$_DPMS_DEV WD=$_DPMS_WD XR=$_DPMS_XR off=${_DPMS_OFF}s" >> "$_DPMS_LOG"
            while true; do
                if timeout 5 dd if="$_DPMS_DEV" bs=24 count=1 >/dev/null 2>&1; then
                    # Input received — always wake display (idempotent if already on)
                    _last=$(date +%s)
                    _wlopm_on
                    _state=on
                else
                    # 5-second window with no input — check idle duration
                    _now=$(date +%s)
                    _idle=$(( _now - _last ))
                    if [ "$_state" != "off" ] && [ "$_idle" -ge "$_DPMS_OFF" ]; then
                        _wlopm_off
                        _state=off
                    elif [ "$_state" = "on" ] && [ "$_idle" -ge "$_DPMS_DIM" ]; then
                        _wlopm_off
                        _state=dim
                    fi
                fi
            done
        ) &
        log "Input-DPMS daemon started (pid $!)"

    elif command -v wlopm &>/dev/null; then
        # No readable input device — fall back to swayidle for timer-based off
        log "No input device found; using swayidle for DPMS (wake requires physical compositor activity)"
        if command -v swayidle &>/dev/null; then
            swayidle -w \
                timeout "$dim_t" "WAYLAND_DISPLAY=${WAYLAND_DISPLAY:-wayland-0} XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-/run/user/$(id -u)} wlopm --off '*' 2>/dev/null || true" \
                resume           "WAYLAND_DISPLAY=${WAYLAND_DISPLAY:-wayland-0} XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-/run/user/$(id -u)} wlopm --on  '*' 2>/dev/null || true" \
                &
            log "swayidle started (pid $!)"
        fi
    fi
else
    log "Screen power management disabled (always on)"
    if command -v wlopm &>/dev/null; then
        WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-0}" XDG_RUNTIME_DIR="$XDG_RUNTIME_DIR" \
            wlopm --on '*' 2>/dev/null || true
    fi
fi

# Find Chromium binary
CHROMIUM=""
for bin in chromium-browser chromium google-chrome; do
    command -v "$bin" &>/dev/null && CHROMIUM="$bin" && break
done
[ -z "$CHROMIUM" ] && { log "ERROR: chromium not found"; exit 1; }

# Profile and lock management
KIOSK_PROFILE_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/chromium-kiosk"
mkdir -p "$KIOSK_PROFILE_DIR"
rm -f "$KIOSK_PROFILE_DIR/SingletonLock" 2>/dev/null || true
rm -f "$KIOSK_PROFILE_DIR/Default/Preferences" 2>/dev/null || true

# Wait for Grafana to be ready before launching browser (avoids blank/error page on cold start)
log "Waiting for Grafana to be ready..."
for i in $(seq 1 60); do
    if curl -sf "$GRAFANA_URL/api/health" | grep -q '"database": "ok"' 2>/dev/null; then
        log "Grafana ready after ${i}s"
        break
    fi
    sleep 1
done

log "Launching $CHROMIUM in kiosk mode"

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
        --remote-debugging-port=9222 \
        --remote-allow-origins=http://localhost:9222 \
        --user-data-dir="$KIOSK_PROFILE_DIR" \
        --app="$KIOSK_URL" \
        2>/dev/null || true
    log "Chromium exited — restarting in 5s"
    sleep 5
done
