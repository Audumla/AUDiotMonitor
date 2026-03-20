#!/usr/bin/env bash
# AUDiot Kiosk Installer
#
# Sets up the kiosk launcher to start automatically when the desktop session
# starts. This script's only job is to configure the autostart method. It
# assumes all necessary files (kiosk.sh, etc.) are already in place.
#
# Usage (run as the display user, not root):
#   ./kiosk-install.sh

set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/opt/docker/dashboard}"
KIOSK_SCRIPT="$INSTALL_DIR/kiosk.sh"
KIOSK_ENV="$INSTALL_DIR/config/kiosk.env"
AUTOSTART_DIR="$HOME/.config/autostart"
SERVICE_FILE="$HOME/.config/systemd/user/audiot-kiosk.service"

info()  { echo "[install] $*"; }
ok()    { echo "[install] ✓ $*"; }

# ── Method 1: systemd user service (preferred) ────────────────────────
if command -v systemctl >/dev/null 2>&1 && systemctl --user &>/dev/null 2>&1; then
    info "Setting up systemd user service..."
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
    info "  Manage with: systemctl --user {start|stop|status} audiot-kiosk"

# ── Method 2: XDG autostart .desktop (fallback) ───────────────────────
else
    info "Setting up XDG autostart..."
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
    info "  The kiosk will start on the next desktop login."
fi

echo ""
echo "════════════════════════════════════════"
echo " AUDiot Kiosk setup complete."
echo "════════════════════════════════════════"
echo " To uninstall, run:"
echo "   systemctl --user disable --now audiot-kiosk.service 2>/dev/null"
echo "   rm -f $AUTOSTART_DIR/audiot-kiosk.desktop"
echo "════════════════════════════════════════"
