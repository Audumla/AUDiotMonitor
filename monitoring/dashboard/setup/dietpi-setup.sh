#!/usr/bin/env bash
# AUDiot DietPi Kiosk Setup
#
# Configures a minimal DietPi installation to run the AUDiot kiosk:
#   - Installs required packages (chromium, docker, x11 tools)
#   - Configures autologin to X11 console session
#   - Installs and enables the systemd kiosk service
#   - Applies udev rule to prevent ILITEK touchscreen USB autosuspend
#
# Run as root on a fresh DietPi install:
#   sudo bash dietpi-setup.sh
#
# Prerequisites: DietPi with network configured, project deployed to
#   /opt/docker/services/dashboard/

set -euo pipefail

KIOSK_USER="${KIOSK_USER:-dietpi}"
KIOSK_DIR="${KIOSK_DIR:-/opt/docker/services/dashboard}"
SERVICE_FILE="$KIOSK_DIR/audiot-kiosk.service"

log() { echo "[dietpi-setup] $*"; }

# ── 1. Packages ───────────────────────────────────────────────────────────────

log "Installing packages..."
apt-get update -qq
apt-get install -y --no-install-recommends \
    chromium \
    xserver-xorg-core \
    xserver-xorg-input-all \
    xserver-xorg-video-fbdev \
    xinit \
    openbox \
    x11-xserver-utils \
    curl \
    docker.io \
    docker-compose-plugin

# ── 2. Autologin to X11 console session ──────────────────────────────────────
# DietPi uses /DietPi/config/dietpi.txt for autologin settings.
# This configures the dietpi user to autologin and start X.

log "Configuring autologin..."

# getty autologin
mkdir -p /etc/systemd/system/getty@tty1.service.d
cat > /etc/systemd/system/getty@tty1.service.d/autologin.conf << EOF
[Service]
ExecStart=
ExecStart=-/sbin/agetty --autologin $KIOSK_USER --noclear %I \$TERM
EOF

# Start X session on login — add to .bash_profile if not headless
PROFILE="/home/$KIOSK_USER/.bash_profile"
if ! grep -q 'startx' "$PROFILE" 2>/dev/null; then
    cat >> "$PROFILE" << 'EOF'

# Start X session on tty1 autologin (AUDiot kiosk)
if [ -z "${DISPLAY:-}" ] && [ "$(tty)" = "/dev/tty1" ]; then
    exec startx /home/dietpi/.xinitrc -- :0 vt1
fi
EOF
    log "Added startx to $PROFILE"
fi

# Minimal .xinitrc: openbox + DPMS disable (kiosk.sh handles DPMS itself)
XINITRC="/home/$KIOSK_USER/.xinitrc"
cat > "$XINITRC" << 'EOF'
#!/bin/sh
# Minimal X session for AUDiot kiosk
xset s off
xset -dpms
xset s noblank
openbox --config-file /dev/null &
exec true
EOF
chmod +x "$XINITRC"
chown "$KIOSK_USER:$KIOSK_USER" "$XINITRC" "$PROFILE" 2>/dev/null || true

# ── 3. Systemd kiosk service ──────────────────────────────────────────────────

log "Installing kiosk service..."
# Patch service file to reference the correct user
sed "s/User=dietpi/User=$KIOSK_USER/" "$SERVICE_FILE" \
    > /etc/systemd/system/audiot-kiosk.service

systemctl daemon-reload
systemctl enable audiot-kiosk.service
log "audiot-kiosk.service enabled"

# ── 4. udev rule: disable ILITEK touchscreen USB autosuspend ─────────────────

log "Installing udev rule for ILITEK touchscreen..."
cat > /etc/udev/rules.d/99-ilitek-touch.rules << 'EOF'
ACTION=="add", SUBSYSTEM=="usb", ATTR{idVendor}=="222a", ATTR{power/control}="on"
EOF
udevadm control --reload-rules
log "udev rule installed"

# ── 5. Docker group ───────────────────────────────────────────────────────────

usermod -aG docker "$KIOSK_USER" 2>/dev/null || true
log "Added $KIOSK_USER to docker group"

log ""
log "Setup complete. Reboot to start the kiosk."
log "  sudo reboot"
