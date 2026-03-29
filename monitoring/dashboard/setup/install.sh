#!/usr/bin/env bash
# AUDiot DietPi Bootstrap Installer
#
# Installs the complete AUDiot monitoring stack on a fresh DietPi host.
# Pulls all Docker images from Docker Hub — no git clone required.
#
# Run as root:
#   curl -sSL https://raw.githubusercontent.com/Audumla/AUDiotMonitor/dietpi/monitoring/dashboard/setup/install.sh | sudo bash
#
# Options (set as env vars or flags):
#   --no-kiosk            Skip X11 kiosk setup (headless/server mode)
#   --tag <tag>           Set both dashboard and kiosk image tags
#   --dashboard-tag <tag> Dashboard image tag (default: latest)
#   --kiosk-tag <tag>     Kiosk image tag (default: latest)
#   --install-dir <path>  Deploy path (default: /opt/docker/services/dashboard)
#   --user <name>         Kiosk user (default: dietpi)

set -euo pipefail

# ── Configuration ─────────────────────────────────────────────────────────────

BRANCH="${AUDIOT_BRANCH:-dietpi}"
RAW_BASE="https://raw.githubusercontent.com/Audumla/AUDiotMonitor/${BRANCH}/monitoring/dashboard"
INSTALL_DIR="${INSTALL_DIR:-/opt/docker/services/dashboard}"
KIOSK_USER="${KIOSK_USER:-dietpi}"
DASHBOARD_TAG="${DASHBOARD_TAG:-latest}"
KIOSK_TAG="${KIOSK_TAG:-latest}"
SETUP_KIOSK=true

while [[ $# -gt 0 ]]; do
    case "$1" in
        --no-kiosk)    SETUP_KIOSK=false; shift ;;
        --tag)         DASHBOARD_TAG="$2"; KIOSK_TAG="$2"; shift 2 ;;
        --dashboard-tag) DASHBOARD_TAG="$2"; shift 2 ;;
        --kiosk-tag)   KIOSK_TAG="$2"; shift 2 ;;
        --install-dir) INSTALL_DIR="$2"; shift 2 ;;
        --user)        KIOSK_USER="$2"; shift 2 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

log() { echo "$(date '+%H:%M:%S') [audiot-install] $*"; }

fetch() {
    local url="$RAW_BASE/$1"
    local dst="$2"
    mkdir -p "$(dirname "$dst")"
    curl -fsSL "$url" -o "$dst"
    log "  fetched $(basename "$dst")"
}

# Load shared library (fetch it from GitHub so curl-pipe invocations work)
fetch "setup/lib.sh" "/tmp/audiot-lib.sh"
# shellcheck source=/dev/null
. /tmp/audiot-lib.sh

# ── 1. Prerequisites ──────────────────────────────────────────────────────────

log "Installing prerequisites..."
apt-get update -qq
apt-get install -y --no-install-recommends curl ca-certificates

# Docker
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com | sh
fi
add_docker_group "$KIOSK_USER"

if [ "$SETUP_KIOSK" = "true" ]; then
    apt-get install -y --no-install-recommends \
        chromium \
        xserver-xorg-core \
        xserver-xorg-input-all \
        xserver-xorg-video-fbdev \
        xinit \
        openbox \
        x11-xserver-utils
fi

# ── 2. Fetch host-side files from GitHub raw ──────────────────────────────────

log "Fetching config files..."
mkdir -p "$INSTALL_DIR"

fetch "docker-compose.yml"          "$INSTALL_DIR/docker-compose.yml"
fetch ".env.example"                "$INSTALL_DIR/.env.example"
fetch "config/kiosk.env.example"    "$INSTALL_DIR/config/kiosk.env"
fetch "audiot-kiosk.service"        "$INSTALL_DIR/audiot-kiosk.service"

# Kiosk scripts (needed on the host; also baked into the Docker image)
fetch "kiosk.sh"        "$INSTALL_DIR/kiosk.sh"
fetch "kiosk-switch.sh" "$INSTALL_DIR/kiosk-switch.sh"
chmod +x "$INSTALL_DIR/kiosk.sh" "$INSTALL_DIR/kiosk-switch.sh"

# Dashboard JSON files
for profile in profiles/debug profiles/mobile profiles/standard profiles/wide-screens custom; do
    mkdir -p "$INSTALL_DIR/dashboards/$profile"
    # fetch manifest to discover files (graceful — skip if directory is empty)
done
fetch "dashboards/profiles/standard/panel-1080p.json"      "$INSTALL_DIR/dashboards/profiles/standard/panel-1080p.json"
fetch "dashboards/profiles/standard/system-overview.json"   "$INSTALL_DIR/dashboards/profiles/standard/system-overview.json"
fetch "dashboards/profiles/mobile/panel-portrait.json"      "$INSTALL_DIR/dashboards/profiles/mobile/panel-portrait.json"
fetch "dashboards/custom/triple-gpu-wide.json"              "$INSTALL_DIR/dashboards/custom/triple-gpu-wide.json" || true
fetch "dashboards/custom/triple-gpu-combined.json"          "$INSTALL_DIR/dashboards/custom/triple-gpu-combined.json" || true

# Seed .env from example if not already present
if [ ! -f "$INSTALL_DIR/.env" ]; then
    cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
    log "  created .env from .env.example — set PROMETHEUS_URL and HWEXP_URL before starting"
fi

chown -R "$KIOSK_USER:$KIOSK_USER" "$INSTALL_DIR" 2>/dev/null || true

# ── 3. udev rule: ILITEK touchscreen autosuspend ──────────────────────────────

log "Installing udev rule for ILITEK touchscreen..."
install_udev_rule_ilitek

# ── 5. Start monitoring stack ─────────────────────────────────────────────────

log "Pulling images and starting monitoring stack..."
cd "$INSTALL_DIR"
DASHBOARD_TAG="$DASHBOARD_TAG" KIOSK_TAG="$KIOSK_TAG" docker compose up -d
log "Stack started"

# ── 6. Extract kiosk tools from dashboard image ───────────────────────────────
# wl-gammarelay is baked into the dashboard image at the correct arch.
# Extract it now that the image is already pulled — no separate download needed.

if [ "$SETUP_KIOSK" = "true" ]; then
    log "Extracting wl-gammarelay from dashboard image..."
    extract_wl_gammarelay "audumla/audiot-dashboard:${DASHBOARD_TAG:-latest}"
fi

# ── 5. Kiosk X11 autologin and systemd service ────────────────────────────────

if [ "$SETUP_KIOSK" = "true" ]; then
    log "Configuring kiosk autologin for user: $KIOSK_USER..."
    setup_getty_autologin "$KIOSK_USER"
    setup_bash_profile    "$KIOSK_USER"
    setup_xinitrc         "$KIOSK_USER"

    # Install kiosk as a systemd user service
    USER_SERVICE_DIR="/home/$KIOSK_USER/.config/systemd/user"
    mkdir -p "$USER_SERVICE_DIR"
    sed '/^User=/d' "$INSTALL_DIR/audiot-kiosk.service" \
        > "$USER_SERVICE_DIR/audiot-kiosk.service"
    chown -R "$KIOSK_USER:$KIOSK_USER" "/home/$KIOSK_USER/.config" 2>/dev/null || true
    loginctl enable-linger "$KIOSK_USER" 2>/dev/null || true
    sudo -u "$KIOSK_USER" \
        XDG_RUNTIME_DIR="/run/user/$(id -u "$KIOSK_USER")" \
        systemctl --user enable audiot-kiosk.service 2>/dev/null || true
    log "audiot-kiosk.service enabled for $KIOSK_USER"
fi

# ── Done ──────────────────────────────────────────────────────────────────────

log ""
log "Installation complete."
log ""
log "Stack status:  docker compose -C $INSTALL_DIR ps"
log "Grafana:       http://$(hostname -I | awk '{print $1}'):3000  (admin / admin)"
log "Prometheus:    http://$(hostname -I | awk '{print $1}'):9090"
if [ "$SETUP_KIOSK" = "true" ]; then
    log ""
    log "Reboot to launch the kiosk display:"
    log "  sudo reboot"
fi
