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

# ── 4. Disable system suspend ─────────────────────────────────────────────────
# The kiosk manages display sleep via wlopm/DPMS.  Powering off the Wayland
# output (wlopm --off) removes the active seat from logind's view, which can
# trigger systemd-logind's idle-suspend action and put the whole RPi to sleep.
# Mask all sleep targets and set IdleAction=ignore so only the display sleeps.

log "Disabling system suspend..."
systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target \
    2>/dev/null || true

# logind: no idle suspend, ignore power/sleep keys (kiosk has no lid or power btn)
mkdir -p /etc/systemd/logind.conf.d
cat > /etc/systemd/logind.conf.d/kiosk-no-sleep.conf << 'EOF'
[Login]
IdleAction=ignore
IdleActionSec=0
HandleSuspendKey=ignore
HandleHibernateKey=ignore
HandlePowerKey=ignore
HandleLidSwitch=ignore
HandleLidSwitchExternalPower=ignore
HandleLidSwitchDocked=ignore
EOF
systemctl restart systemd-logind 2>/dev/null || true
log "System suspend disabled"

# ── 5. Disable USB autosuspend ────────────────────────────────────────────────
# The USB hub connecting eth0 (smsc95xx) and the touch screen has autosuspend
# enabled by default. When the hub suspends, SSH connectivity and touch input
# are lost. Disable autosuspend globally for all USB devices.

log "Disabling USB autosuspend..."
echo 'options usbcore autosuspend=-1' \
    > /etc/modprobe.d/kiosk-no-usb-autosuspend.conf
# Apply immediately without reboot
for dev in /sys/bus/usb/devices/*/power/control; do
    echo on > "$dev" 2>/dev/null || true
done
log "USB autosuspend disabled"

# ── 6. udev rule: disable ILITEK touchscreen USB autosuspend ─────────────────
# Belt-and-suspenders: keep the per-device udev rule even with global disable.

log "Installing udev rule for ILITEK touchscreen..."
cat > /etc/udev/rules.d/99-ilitek-touch.rules << 'EOF'
ACTION=="add", SUBSYSTEM=="usb", ATTR{idVendor}=="222a", ATTR{power/control}="on"
EOF
udevadm control --reload-rules
log "udev rule installed"

# ── 7. wl-gammarelay binary (display brightness / display sleep) ──────────────
# Extracted from the dashboard image — correct arch guaranteed, no separate download.
# Requires the dashboard image to be present (run 'docker compose up -d' first,
# or 'docker pull audumla/audiot-dashboard:latest' if running standalone).

log "Installing wl-gammarelay..."
if [ -f /usr/local/bin/wl-gammarelay ]; then
    log "  wl-gammarelay already installed — skipping"
else
    _image="audumla/audiot-dashboard:latest"
    if docker image inspect "$_image" >/dev/null 2>&1 || docker pull "$_image" >/dev/null 2>&1; then
        docker run --rm --entrypoint cat "$_image" \
            /opt/audiot-dashboard/wl-gammarelay > /usr/local/bin/wl-gammarelay
        chmod +x /usr/local/bin/wl-gammarelay
        log "  wl-gammarelay installed"
    else
        log "  WARNING: could not pull dashboard image — skipping wl-gammarelay install"
        log "  Run manually: docker run --rm audumla/audiot-dashboard cat /opt/audiot-dashboard/wl-gammarelay > /usr/local/bin/wl-gammarelay"
    fi
fi

# ── 8. Docker group ───────────────────────────────────────────────────────────

usermod -aG docker "$KIOSK_USER" 2>/dev/null || true
log "Added $KIOSK_USER to docker group"

# ── 9. Prometheus config for DietPi (single-host layout) ─────────────────────

PROMETHEUS_CFG="$KIOSK_DIR/config/prometheus/prometheus.yml"
DIETPI_CFG="$KIOSK_DIR/config/prometheus/prometheus.dietpi.yml"
if [ -f "$DIETPI_CFG" ] && [ ! -f "$PROMETHEUS_CFG" ]; then
    cp "$DIETPI_CFG" "$PROMETHEUS_CFG"
    log "Installed prometheus.yml for DietPi single-host layout"
fi

log ""
log "Setup complete."
log ""
log "Start the monitoring stack:"
log "  cd $KIOSK_DIR"
log "  docker compose -f docker-compose.dietpi.yml up -d"
log ""
log "Then reboot to start the kiosk browser:"
log "  sudo reboot"
