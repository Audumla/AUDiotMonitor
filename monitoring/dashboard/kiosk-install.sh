#!/usr/bin/env bash
# AUDiot Kiosk Installer
#
# Sets up the kiosk launcher to start automatically when the desktop session
# starts. Works on Raspberry Pi OS and Debian with LXDE, labwc, or GNOME.
#
# Usage (run as the display user, not root):
#   GRAFANA_URL=http://localhost:3000 ./kiosk-install.sh
#
# Options via environment:
#   GRAFANA_URL      Grafana URL  (default: http://localhost:3000)
#   INSTALL_DIR      Where to copy kiosk assets  (default: /opt/docker/dashboard)

set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/opt/docker/dashboard}"
KIOSK_SCRIPT="$INSTALL_DIR/kiosk.sh"
KIOSK_ENV="$INSTALL_DIR/config/kiosk.env"
AUTOSTART_DIR="$HOME/.config/autostart"
SERVICE_FILE="$HOME/.config/systemd/user/audiot-kiosk.service"

# ── Helpers ───────────────────────────────────────────────────────────────────

info()  { echo "[install] $*"; }
ok()    { echo "[install] ✓ $*"; }
warn()  { echo "[install] ! $*"; }

# ── Install Chromium if missing ───────────────────────────────────────────────

if ! command -v chromium-browser &>/dev/null && ! command -v chromium &>/dev/null; then
    info "Chromium not found — installing..."
    sudo apt-get update -qq
    sudo apt-get install -y chromium-browser
    ok "Chromium installed"
else
    ok "Chromium already installed"
fi

# ── Copy kiosk assets to install directory ───────────────────────────────────

if [ "$(realpath "$0")" != "$(realpath "$KIOSK_SCRIPT")" ]; then
    sudo mkdir -p "$INSTALL_DIR"
    sudo cp "$(dirname "$0")/install-layout.sh" "$INSTALL_DIR/install-layout.sh"
    sudo cp "$(dirname "$0")/set-dashboard.sh" "$INSTALL_DIR/set-dashboard.sh"
    sudo chmod +x "$INSTALL_DIR/install-layout.sh"
    sudo chmod +x "$INSTALL_DIR/set-dashboard.sh"
    sudo "$INSTALL_DIR/install-layout.sh"
    sudo cp "$(dirname "$0")/kiosk.sh" "$KIOSK_SCRIPT"
    sudo chmod +x "$KIOSK_SCRIPT"
    ok "kiosk.sh installed to $KIOSK_SCRIPT"
else
    ok "kiosk.sh already in place"
fi

if [ ! -f "$KIOSK_ENV" ] && [ -f "$(dirname "$0")/config/kiosk.env.example" ]; then
    sudo mkdir -p "$INSTALL_DIR/config"
    sudo cp "$(dirname "$0")/config/kiosk.env.example" "$KIOSK_ENV"
    ok "Installed kiosk config: $KIOSK_ENV"
fi

# ── Method 1: systemd user service (preferred on Debian bookworm) ─────────────

if systemctl --user &>/dev/null 2>&1; then
    mkdir -p "$(dirname "$SERVICE_FILE")"
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=AUDiot Grafana Kiosk
After=graphical-session.target network-online.target
Wants=graphical-session.target

[Service]
Type=simple
Environment=DISPLAY=:0
Environment=XAUTHORITY=%h/.Xauthority
ExecStart=$KIOSK_SCRIPT
Restart=always
RestartSec=10

[Install]
WantedBy=graphical-session.target
EOF

    systemctl --user daemon-reload
    systemctl --user enable audiot-kiosk.service
    systemctl --user restart audiot-kiosk.service 2>/dev/null || true
    ok "Installed as systemd user service: audiot-kiosk.service"
    info "  Manage with: systemctl --user {start|stop|status|restart} audiot-kiosk"

# ── Method 2: XDG autostart .desktop (LXDE / Openbox / other desktop) ────────

else
    mkdir -p "$AUTOSTART_DIR"
    cat > "$AUTOSTART_DIR/audiot-kiosk.desktop" << EOF
[Desktop Entry]
Type=Application
Name=AUDiot Kiosk
Comment=Launch AUDiot Grafana dashboard in kiosk mode
Exec=$KIOSK_SCRIPT
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
EOF
    ok "Installed XDG autostart entry: $AUTOSTART_DIR/audiot-kiosk.desktop"
    info "  The kiosk will start on next desktop login"
fi

# ── Summary ───────────────────────────────────────────────────────────────────

echo ""
echo "════════════════════════════════════════"
echo " AUDiot Kiosk installed"
echo "════════════════════════════════════════"
echo " Config file : $KIOSK_ENV"
echo " Grafana URL : edit in $KIOSK_ENV"
echo " Dashboard   : edit in $KIOSK_ENV"
echo ""
echo " To start now without rebooting:"
echo "   $KIOSK_SCRIPT &"
echo ""
echo " To uninstall:"
echo "   systemctl --user disable --now audiot-kiosk.service 2>/dev/null"
echo "   rm -f $AUTOSTART_DIR/audiot-kiosk.desktop"
echo "════════════════════════════════════════"
