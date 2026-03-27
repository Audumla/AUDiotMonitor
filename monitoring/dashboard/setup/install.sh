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
#   --tag <tag>           Docker image tag (default: latest)
#   --install-dir <path>  Deploy path (default: /opt/docker/services/dashboard)
#   --user <name>         Kiosk user (default: dietpi)

set -euo pipefail

# ── Configuration ─────────────────────────────────────────────────────────────

BRANCH="${AUDIOT_BRANCH:-dietpi}"
RAW_BASE="https://raw.githubusercontent.com/Audumla/AUDiotMonitor/${BRANCH}/monitoring/dashboard"
INSTALL_DIR="${INSTALL_DIR:-/opt/docker/services/dashboard}"
KIOSK_USER="${KIOSK_USER:-dietpi}"
AUDIOT_TAG="${AUDIOT_TAG:-latest}"
SETUP_KIOSK=true

while [[ $# -gt 0 ]]; do
    case "$1" in
        --no-kiosk)    SETUP_KIOSK=false; shift ;;
        --tag)         AUDIOT_TAG="$2"; shift 2 ;;
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

# ── 1. Prerequisites ──────────────────────────────────────────────────────────

log "Installing prerequisites..."
apt-get update -qq
apt-get install -y --no-install-recommends curl ca-certificates

# Docker
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com | sh
fi
usermod -aG docker "$KIOSK_USER" 2>/dev/null || true

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
cat > /etc/udev/rules.d/99-ilitek-touch.rules << 'EOF'
ACTION=="add", SUBSYSTEM=="usb", ATTR{idVendor}=="222a", ATTR{power/control}="on"
EOF
udevadm control --reload-rules

# ── 5. Start monitoring stack ─────────────────────────────────────────────────

log "Pulling images and starting monitoring stack..."
cd "$INSTALL_DIR"
AUDIOT_TAG="$AUDIOT_TAG" docker compose up -d
log "Stack started"

# ── 6. Extract kiosk tools from dashboard image ───────────────────────────────
# wl-gammarelay is baked into the dashboard image at the correct arch.
# Extract it now that the image is already pulled — no separate download needed.

if [ "$SETUP_KIOSK" = "true" ] && [ ! -f /usr/local/bin/wl-gammarelay ]; then
    log "Extracting wl-gammarelay from dashboard image..."
    _image="audumla/audiot-dashboard:${AUDIOT_TAG:-latest}"
    docker run --rm --entrypoint cat "$_image" \
        /opt/audiot-dashboard/wl-gammarelay > /usr/local/bin/wl-gammarelay
    chmod +x /usr/local/bin/wl-gammarelay
    log "  wl-gammarelay installed (/usr/local/bin/wl-gammarelay)"
fi

# ── 5. Kiosk X11 autologin and systemd service ────────────────────────────────

if [ "$SETUP_KIOSK" = "true" ]; then
    log "Configuring kiosk autologin for user: $KIOSK_USER..."

    # getty autologin on tty1
    mkdir -p /etc/systemd/system/getty@tty1.service.d
    cat > /etc/systemd/system/getty@tty1.service.d/autologin.conf << EOF
[Service]
ExecStart=
ExecStart=-/sbin/agetty --autologin $KIOSK_USER --noclear %I \$TERM
EOF

    # .bash_profile: launch X on tty1
    PROFILE="/home/$KIOSK_USER/.bash_profile"
    if ! grep -q 'audiot-kiosk' "$PROFILE" 2>/dev/null; then
        cat >> "$PROFILE" << 'BASHEOF'

# AUDiot kiosk: start X session on tty1 autologin
if [ -z "${DISPLAY:-}" ] && [ "$(tty)" = "/dev/tty1" ]; then
    exec startx "$HOME/.xinitrc" -- :0 vt1
fi
BASHEOF
        chown "$KIOSK_USER:$KIOSK_USER" "$PROFILE" 2>/dev/null || true
    fi

    # Minimal .xinitrc: openbox session (kiosk.sh handles DPMS itself)
    XINITRC="/home/$KIOSK_USER/.xinitrc"
    cat > "$XINITRC" << 'XINITEOF'
#!/bin/sh
xset s off
xset -dpms
xset s noblank
openbox --config-file /dev/null &
exec true
XINITEOF
    chmod +x "$XINITRC"
    chown "$KIOSK_USER:$KIOSK_USER" "$XINITRC" 2>/dev/null || true

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
