#!/usr/bin/env bash
# AUDiot DietPi Bootstrap Installer
#
# Downloads and installs the complete AUDiot monitoring stack on a fresh DietPi host.
# Handles prerequisites, layout deployment, Docker stack startup, and kiosk setup.
#
# Run as root:
#   curl -sSL https://raw.githubusercontent.com/Audumla/AUDiotMonitor/dietpi/monitoring/dashboard/setup/install.sh | sudo bash
#
# Or with options:
#   sudo bash install.sh [--no-kiosk] [--branch dietpi] [--install-dir /opt/docker/services/dashboard]

set -euo pipefail

# ── Configuration ─────────────────────────────────────────────────────────────

REPO_URL="${AUDIOT_REPO:-https://github.com/Audumla/AUDiotMonitor.git}"
BRANCH="${AUDIOT_BRANCH:-dietpi}"
INSTALL_DIR="${INSTALL_DIR:-/opt/docker/services/dashboard}"
KIOSK_USER="${KIOSK_USER:-dietpi}"
SETUP_KIOSK=true

while [[ $# -gt 0 ]]; do
    case "$1" in
        --no-kiosk)    SETUP_KIOSK=false; shift ;;
        --branch)      BRANCH="$2"; shift 2 ;;
        --install-dir) INSTALL_DIR="$2"; shift 2 ;;
        --user)        KIOSK_USER="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

CLONE_DIR="/tmp/audiot-install-$$"

log()  { echo "$(date '+%H:%M:%S') [audiot-install] $*"; }
die()  { echo "ERROR: $*" >&2; exit 1; }

# ── 1. Prerequisites ──────────────────────────────────────────────────────────

log "Installing prerequisites..."
apt-get update -qq
apt-get install -y --no-install-recommends \
    git \
    curl \
    ca-certificates \
    docker.io \
    docker-compose-plugin

# Add kiosk user to docker group
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

# ── 2. Clone repository ───────────────────────────────────────────────────────

log "Cloning $REPO_URL (branch: $BRANCH)..."
rm -rf "$CLONE_DIR"
git clone --depth=1 --branch "$BRANCH" "$REPO_URL" "$CLONE_DIR"
SOURCE_DIR="$CLONE_DIR/monitoring/dashboard"

# ── 3. Deploy layout ──────────────────────────────────────────────────────────

log "Deploying layout to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
INSTALL_DIR="$INSTALL_DIR" bash "$SOURCE_DIR/install-layout.sh"

# DietPi uses the combined compose file as the default
cp "$SOURCE_DIR/docker-compose.dietpi.yml" "$INSTALL_DIR/docker-compose.yml"
log "Installed docker-compose.yml (DietPi combined stack)"

# Deploy prometheus config for DietPi single-host layout
PROM_CFG="$INSTALL_DIR/config/prometheus/prometheus.yml"
mkdir -p "$(dirname "$PROM_CFG")"
if [ ! -f "$PROM_CFG" ]; then
    cp "$SOURCE_DIR/config/prometheus/prometheus.dietpi.yml" "$PROM_CFG"
    log "Installed prometheus.yml"
fi

# Copy kiosk service file
cp "$SOURCE_DIR/audiot-kiosk.service" "$INSTALL_DIR/audiot-kiosk.service"

# Fix ownership of install directory
chown -R "$KIOSK_USER:$KIOSK_USER" "$INSTALL_DIR" 2>/dev/null || true

# ── 4. udev rule: ILITEK touchscreen autosuspend ──────────────────────────────

log "Installing udev rule for ILITEK touchscreen..."
cat > /etc/udev/rules.d/99-ilitek-touch.rules << 'EOF'
ACTION=="add", SUBSYSTEM=="usb", ATTR{idVendor}=="222a", ATTR{power/control}="on"
EOF
udevadm control --reload-rules

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
        log "Updated $PROFILE"
    fi

    # Minimal .xinitrc: openbox + kiosk service wakes into it
    XINITRC="/home/$KIOSK_USER/.xinitrc"
    cat > "$XINITRC" << 'XINITEOF'
#!/bin/sh
# AUDiot kiosk X session
xset s off
xset -dpms
xset s noblank
openbox --config-file /dev/null &
exec true
XINITEOF
    chmod +x "$XINITRC"
    chown "$KIOSK_USER:$KIOSK_USER" "$XINITRC" 2>/dev/null || true

    # Install and enable kiosk systemd user service
    # Use user service so it starts after the X session is up
    SERVICE_SRC="$INSTALL_DIR/audiot-kiosk.service"
    USER_SERVICE_DIR="/home/$KIOSK_USER/.config/systemd/user"
    mkdir -p "$USER_SERVICE_DIR"
    # Patch User= out of service file (not needed for user service)
    sed '/^User=/d' "$SERVICE_SRC" > "$USER_SERVICE_DIR/audiot-kiosk.service"
    chown -R "$KIOSK_USER:$KIOSK_USER" "/home/$KIOSK_USER/.config" 2>/dev/null || true
    # Enable lingering so user services start at boot
    loginctl enable-linger "$KIOSK_USER" 2>/dev/null || true
    # Enable the service for the kiosk user
    sudo -u "$KIOSK_USER" XDG_RUNTIME_DIR="/run/user/$(id -u "$KIOSK_USER")" \
        systemctl --user enable audiot-kiosk.service 2>/dev/null || true
    log "audiot-kiosk.service installed as user service for $KIOSK_USER"
fi

# ── 6. Start monitoring stack ─────────────────────────────────────────────────

log "Starting monitoring stack..."
cd "$INSTALL_DIR"
docker compose up -d
log "Stack started"

# ── 7. Cleanup ────────────────────────────────────────────────────────────────

rm -rf "$CLONE_DIR"

log ""
log "Installation complete."
log ""
log "Monitoring stack: docker compose -f $INSTALL_DIR/docker-compose.yml ps"
log "Grafana:          http://localhost:3000  (admin / admin)"
log "Prometheus:       http://localhost:9090"
if [ "$SETUP_KIOSK" = "true" ]; then
    log ""
    log "Reboot to start the kiosk display:"
    log "  sudo reboot"
fi
