#!/usr/bin/env bash
# AUDiot Kiosk Launcher
#
# Detects the connected screen resolution and opens the best-matching dashboard
# in Chromium fullscreen/kiosk mode. Restarts Chromium automatically on exit.
#
# Supports two compositor backends (auto-detected):
#   Wayland  — wlopm for DPMS, wlr-randr for resolution (e.g. RPi Desktop / labwc)
#   X11      — xset dpms for DPMS, xrandr for resolution (e.g. DietPi + openbox)
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

# ── Display backend detection ─────────────────────────────────────────────────
# Prefer Wayland if a socket exists; fall back to X11.

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
if [ -z "${DISPLAY:-}" ] && [ -z "${WAYLAND_DISPLAY:-}" ]; then
    DISPLAY=":0"
fi

export XDG_RUNTIME_DIR
[ -n "${WAYLAND_DISPLAY:-}" ] && export WAYLAND_DISPLAY
[ -n "${DBUS_SESSION_BUS_ADDRESS:-}" ] && export DBUS_SESSION_BUS_ADDRESS
[ -n "${XAUTHORITY:-}" ] && export XAUTHORITY
[ -n "${DISPLAY:-}" ] && export DISPLAY

# Select backend: "wayland" if wlopm is available and WAYLAND_DISPLAY is set;
# otherwise "x11" (requires xset).
if [ -n "${WAYLAND_DISPLAY:-}" ] && command -v wlopm &>/dev/null; then
    _BACKEND="wayland"
else
    _BACKEND="x11"
fi

# ── Helpers ───────────────────────────────────────────────────────────────────

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [kiosk] $*"
}

log "Display backend: $_BACKEND (DISPLAY=${DISPLAY:-} WAYLAND_DISPLAY=${WAYLAND_DISPLAY:-})"

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
# Touch/input events on /dev/input are monitored directly so the display wakes
# even when the compositor does not route touch through its idle-notify protocol.
#
# Wayland backend: wlopm --on/--off '*'
# X11 backend:     xset dpms force on/off
#
# KIOSK_DPMS_STANDBY  : seconds of inactivity before display dims/powers off
# KIOSK_DPMS_OFF      : seconds of inactivity before display fully powers off
# KIOSK_IDLE_TIMEOUT  : shorthand when standby == off (0 = always on)
# KIOSK_TOUCH_DEVICE  : override input device path (auto-detected if blank)

find_input_device() {
    # Print the first readable touch/pointer input device path; always returns 0.
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

probe_dimmer() {
    # Probe available dimming tools in priority order; set _DIMMER_READY and
    # the dim/restore command strings used by swayidle callbacks.
    #
    # Priority:
    #   1. wl-gammarelay  — Wayland wlr-gamma-control (installed via Docker image)
    #   2. ddcutil        — DDC/CI brightness over HDMI I2C (monitor must support it)
    #
    # Sets on success: _DIMMER_READY=true, _DIMMER_CMD_DIM, _DIMMER_CMD_RESTORE
    # Logs reason and leaves _DIMMER_READY=false when no tool works.
    _DIMMER_READY=false
    _DIMMER_CMD_DIM=""
    _DIMMER_CMD_RESTORE=""

    [ "$_BACKEND" = "wayland" ] || return 0

    local brightness="${KIOSK_DIM_BRIGHTNESS:-0.2}"
    # ddcutil uses 0-100; convert float to integer percentage
    local brightness_pct
    brightness_pct=$(awk "BEGIN { printf \"%d\", $brightness * 100 }")

    # ── Option 1: wl-gammarelay (Wayland gamma control) ───────────────────────
    # Built into the Docker kiosk image via CI; not installed on bare DietPi.
    if command -v wl-gammarelay &>/dev/null && command -v busctl &>/dev/null; then
        pkill -f wl-gammarelay 2>/dev/null || true
        WAYLAND_DISPLAY="$_DPMS_WD" XDG_RUNTIME_DIR="$_DPMS_XR" \
            wl-gammarelay >/dev/null 2>&1 &
        _relay_pid=$!
        sleep 1
        local _dbus_addr="unix:path=${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/bus"
        if DBUS_SESSION_BUS_ADDRESS="$_dbus_addr" busctl --user set-property \
                rs.wl-gammarelay / rs.wl.gammarelay Brightness d 0.5 >/dev/null 2>&1; then
            DBUS_SESSION_BUS_ADDRESS="$_dbus_addr" busctl --user set-property \
                rs.wl-gammarelay / rs.wl.gammarelay Brightness d 1.0 >/dev/null 2>&1 || true
            log "Dimmer: wl-gammarelay OK (pid $_relay_pid) — brightness control enabled"
            _DIMMER_CMD_DIM="DBUS_SESSION_BUS_ADDRESS='$_dbus_addr' busctl --user set-property rs.wl-gammarelay / rs.wl.gammarelay Brightness d $brightness 2>/dev/null || true"
            _DIMMER_CMD_RESTORE="DBUS_SESSION_BUS_ADDRESS='$_dbus_addr' busctl --user set-property rs.wl-gammarelay / rs.wl.gammarelay Brightness d 1.0 2>/dev/null || true"
            _DIMMER_CMD_BLANK="DBUS_SESSION_BUS_ADDRESS='$_dbus_addr' busctl --user set-property rs.wl-gammarelay / rs.wl.gammarelay Brightness d 0.0 2>/dev/null || true"
            _DIMMER_READY=true
            return 0
        else
            log "Dimmer: wl-gammarelay test failed (wlr-gamma-control not supported?)"
            kill "$_relay_pid" 2>/dev/null || true
        fi
    else
        log "Dimmer: wl-gammarelay not installed (expected in Docker kiosk image)"
    fi

    # ── Option 2: ddcutil (DDC/CI brightness via HDMI I2C) ────────────────────
    # Works when the connected monitor supports DDC/CI on its HDMI input.
    if command -v ddcutil &>/dev/null; then
        if ddcutil getvcp 10 --brief --noverify 2>/dev/null | grep -q '^VCP 10'; then
            log "Dimmer: ddcutil OK (DDC/CI VCP 10) — dimming enabled at ${brightness_pct}%"
            _DIMMER_CMD_DIM="ddcutil setvcp 10 $brightness_pct 2>/dev/null || true"
            _DIMMER_CMD_RESTORE="ddcutil setvcp 10 100 2>/dev/null || true"
            _DIMMER_READY=true
            return 0
        else
            log "Dimmer: ddcutil present but VCP 10 (brightness) not supported by display"
        fi
    fi

    log "Dimmer: no working dimmer found — display will power off without dimming"
}

_dpms_enabled=false
if [ -n "${KIOSK_IDLE_TIMEOUT:-}" ] && [ "${KIOSK_IDLE_TIMEOUT}" -gt 0 ] 2>/dev/null; then
    _dpms_enabled=true
elif [ -n "${KIOSK_DPMS_STANDBY:-}${KIOSK_DPMS_OFF:-}" ]; then
    _dpms_enabled=true
fi

pkill -f swayidle 2>/dev/null || true
pkill -f xautolock 2>/dev/null || true
pkill -f 'swaylock -f -c 000000' 2>/dev/null || true
pkill -f 'dd if=/dev/input' 2>/dev/null || true

_DPMS_LOG="/tmp/kiosk-dpms-${USER}.log"

if [ "$_dpms_enabled" = "true" ]; then
    dim_t="${KIOSK_DPMS_STANDBY:-${KIOSK_IDLE_TIMEOUT:-600}}"
    off_t="${KIOSK_DPMS_OFF:-${KIOSK_IDLE_TIMEOUT:-1800}}"
    log "Screen power management: dim=${dim_t}s off=${off_t}s (backend: $_BACKEND)"

    _DPMS_WD="${WAYLAND_DISPLAY:-wayland-0}"
    _DPMS_XR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"

    if [ "$_BACKEND" = "wayland" ] && command -v swayidle &>/dev/null && command -v wlopm &>/dev/null; then
        # Wayland: swayidle handles the idle timeout and powers the display off via wlopm.
        # However, labwc stops routing input events through its idle-notify protocol once
        # the output is powered off, so swayidle's resume hook may never fire on touch or
        # mouse input.  A supplemental input watcher (below) reads /dev/input directly and
        # calls wlopm --on to guarantee the display wakes on any input event.
        # Probe for dimming support — cascades through wl-gammarelay then ddcutil.
        probe_dimmer

        # Prefer swaylock for blanking: overlays a black surface without cutting the
        # HDMI signal, so the monitor shows black instead of "no signal".
        # Fall back to wlopm --off if swaylock is unavailable.
        _CMD_OFF="WAYLAND_DISPLAY='$_DPMS_WD' XDG_RUNTIME_DIR='$_DPMS_XR' wlopm --off '*' 2>/dev/null || true; echo \"\$(date '+%Y-%m-%d %H:%M:%S') display OFF\" >> '$_DPMS_LOG'"
        _CMD_ON="WAYLAND_DISPLAY='$_DPMS_WD' XDG_RUNTIME_DIR='$_DPMS_XR' wlopm --on '*' 2>/dev/null || true; echo \"\$(date '+%Y-%m-%d %H:%M:%S') display ON\" >> '$_DPMS_LOG'"
        _BLANKER="wlopm"
        _CMD_DIM="$_DIMMER_CMD_DIM; echo \"\$(date '+%Y-%m-%d %H:%M:%S') display DIM\" >> '$_DPMS_LOG'"
        _CMD_RESTORE="$_DIMMER_CMD_RESTORE"

        # wl-gammarelay: use brightness=0 as blanker instead of wlopm --off.
        # Keeps HDMI signal alive so monitor shows black rather than "no signal".
        if [ "$_DIMMER_READY" = "true" ]; then
            _CMD_OFF="$_DIMMER_CMD_BLANK; echo \"\$(date '+%Y-%m-%d %H:%M:%S') display BLANK\" >> '$_DPMS_LOG'"
            _CMD_ON="$_DIMMER_CMD_RESTORE; echo \"\$(date '+%Y-%m-%d %H:%M:%S') display UNBLANK\" >> '$_DPMS_LOG'"
            _BLANKER="wl-gammarelay"
        fi

        if [ "$_DIMMER_READY" = "true" ] && [ "$dim_t" -lt "$off_t" ] 2>/dev/null; then
            log "Screen power management: dim=${dim_t}s off=${off_t}s brightness=${KIOSK_DIM_BRIGHTNESS:-0.2} (swayidle+dimmer+$_BLANKER)"
            WAYLAND_DISPLAY="$_DPMS_WD" XDG_RUNTIME_DIR="$_DPMS_XR" \
            swayidle -w \
                timeout "$dim_t" "$_CMD_DIM" \
                timeout "$off_t" "$_CMD_OFF" \
                resume          "$_CMD_ON; $_CMD_RESTORE" \
                &
        else
            log "Screen power management: off=${off_t}s (swayidle+$_BLANKER)"
            WAYLAND_DISPLAY="$_DPMS_WD" XDG_RUNTIME_DIR="$_DPMS_XR" \
            swayidle -w \
                timeout "$off_t" "$_CMD_OFF" \
                resume          "$_CMD_ON" \
                &
        fi
        log "swayidle started (pid $!)"

        # Supplemental input watcher: reads /dev/input directly so any input device
        # wakes the display even when labwc's idle-notify protocol stops after wlopm --off.
        # Watches ALL readable /dev/input/event* devices (covers touch, mouse, keyboard)
        # plus /dev/input/mice (kernel aggregate, catches hot-plugged mice automatically).
        # Also restores brightness after wake in case wl-gammarelay dimmed the screen.
        # A supervisor loop re-scans every 30 s so devices that appear after startup
        # (e.g. USB touch reconnect) are picked up without a kiosk restart.
        _WAKE_WD="$_DPMS_WD"
        _WAKE_XR="$_DPMS_XR"
        # On wake: restore brightness (if dimmer was used) and dismiss swaylock blanker
        _WAKE_RESTORE="${_DIMMER_CMD_RESTORE:-}"

        _start_wake_watcher() {
            local dev="$1" bs="${2:-24}"
            [ -r "$dev" ] || return 1
            (
                while true; do
                    [ -r "$dev" ] || { sleep 5; continue; }
                    if timeout 10 dd if="$dev" bs="$bs" count=1 >/dev/null 2>&1; then
                        WAYLAND_DISPLAY="$_WAKE_WD" XDG_RUNTIME_DIR="$_WAKE_XR" \
                            wlopm --on '*' 2>/dev/null || true
                        [ -n "$_WAKE_RESTORE" ] && eval "$_WAKE_RESTORE" 2>/dev/null || true
                        echo "$(date '+%Y-%m-%d %H:%M:%S') wake[$dev]: input, display ON" >> "$_DPMS_LOG"
                    fi
                done
            ) &
        }

        # Supervisor: starts one watcher per event device, re-checks every 30 s for new ones
        (
            declare -A _watched
            while true; do
                for dev in /dev/input/event*; do
                    [ -r "$dev" ] || continue
                    [ "${_watched[$dev]+set}" = "set" ] && continue
                    _start_wake_watcher "$dev" 24
                    _watched[$dev]=$!
                    echo "$(date '+%Y-%m-%d %H:%M:%S') wake-supervisor: watching $dev (pid $!)" >> "$_DPMS_LOG"
                done
                sleep 30
            done
        ) &
        log "Wayland wake-watcher supervisor started (pid $!)"

        # /dev/input/mice aggregates all mice (existing and hot-plugged); 3-byte PS/2 packets
        if [ -r /dev/input/mice ]; then
            _start_wake_watcher /dev/input/mice 3
            log "Wayland wake-watcher: mice aggregate /dev/input/mice (pid $!)"
        fi

    elif [ "$_BACKEND" = "x11" ] && command -v xset &>/dev/null; then
        INPUT_DEV="${KIOSK_TOUCH_DEVICE:-$(find_input_device 2>/dev/null)}"
        if [ -n "$INPUT_DEV" ] && [ -r "$INPUT_DEV" ]; then
            log "Input-DPMS: watching $INPUT_DEV (off after ${off_t}s, wake on touch)"
            # X11: poll /dev/input directly since X11 DPMS timers don't respond to touch-only events.
            _DPMS_DP="${DISPLAY:-:0}"
            _DPMS_DEV="$INPUT_DEV"
            _DPMS_DIM="$dim_t"
            _DPMS_OFF="$off_t"
            # Enable DPMS in the X server so force on/off commands work.
            # .xinitrc disables DPMS (xset -dpms) to hand control to this script;
            # we must re-enable it here before using 'xset dpms force'.
            # Set X11's own timers to 0 so the X server never fires them independently.
            DISPLAY="$_DPMS_DP" xset +dpms 2>/dev/null || true
            DISPLAY="$_DPMS_DP" xset dpms 0 0 0 2>/dev/null || true
            (
                _state=on
                _last=$(date +%s)
                _dpms_on() {
                    DISPLAY="$_DPMS_DP" xset +dpms 2>/dev/null || true
                    DISPLAY="$_DPMS_DP" xset dpms force on 2>/dev/null || true
                    echo "$(date '+%Y-%m-%d %H:%M:%S') display ON" >> "$_DPMS_LOG"
                }
                _dpms_off() {
                    DISPLAY="$_DPMS_DP" xset dpms force off 2>/dev/null || true
                    echo "$(date '+%Y-%m-%d %H:%M:%S') display OFF" >> "$_DPMS_LOG"
                }
                echo "$(date '+%Y-%m-%d %H:%M:%S') Input-DPMS started: dev=$_DPMS_DEV off=${_DPMS_OFF}s" >> "$_DPMS_LOG"
                while true; do
                    if [ -z "$_DPMS_DEV" ] || [ ! -r "$_DPMS_DEV" ]; then
                        _new_dev=$(find_input_device 2>/dev/null)
                        if [ -n "$_new_dev" ] && [ "$_new_dev" != "$_DPMS_DEV" ]; then
                            _DPMS_DEV="$_new_dev"
                            echo "$(date '+%Y-%m-%d %H:%M:%S') Input-DPMS: device -> $_DPMS_DEV" >> "$_DPMS_LOG"
                        fi
                        if [ -z "$_DPMS_DEV" ] || [ ! -r "$_DPMS_DEV" ]; then
                            sleep 5; continue
                        fi
                    fi
                    if timeout 5 dd if="$_DPMS_DEV" bs=24 count=1 >/dev/null 2>&1; then
                        _last=$(date +%s)
                        _dpms_on
                        _state=on
                    else
                        _now=$(date +%s)
                        _idle=$(( _now - _last ))
                        if [ "$_state" != "off" ] && [ "$_idle" -ge "$_DPMS_OFF" ]; then
                            _dpms_off; _state=off
                        elif [ "$_state" = "on" ] && [ "$_idle" -ge "$_DPMS_DIM" ]; then
                            _dpms_off; _state=dim
                        fi
                    fi
                done
            ) &
            log "Input-DPMS daemon started (pid $!)"
        else
            # X11 fallback: configure DPMS timers in the X server
            log "No input device found; using X11 DPMS timers (standby=${dim_t}s off=${off_t}s)"
            DISPLAY="${DISPLAY:-:0}" xset +dpms 2>/dev/null || true
            DISPLAY="${DISPLAY:-:0}" xset dpms "$dim_t" "$dim_t" "$off_t" 2>/dev/null || true
        fi
    fi
else
    log "Screen power management disabled (always on)"
    if [ "$_BACKEND" = "wayland" ] && command -v wlopm &>/dev/null; then
        WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-0}" XDG_RUNTIME_DIR="$XDG_RUNTIME_DIR" \
            wlopm --on '*' 2>/dev/null || true
    elif [ "$_BACKEND" = "x11" ] && command -v xset &>/dev/null; then
        DISPLAY="${DISPLAY:-:0}" xset -dpms 2>/dev/null || true
        DISPLAY="${DISPLAY:-:0}" xset s off 2>/dev/null || true
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

# Select Chromium ozone platform flag based on backend
if [ "$_BACKEND" = "wayland" ]; then
    _OZONE_FLAG="--ozone-platform=wayland"
else
    _OZONE_FLAG="--ozone-platform=x11"
fi

log "Launching $CHROMIUM in kiosk mode ($_OZONE_FLAG)"

while true; do
    "$CHROMIUM" \
        $_OZONE_FLAG \
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
