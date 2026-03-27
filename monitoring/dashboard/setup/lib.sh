#!/usr/bin/env bash
# AUDiot kiosk setup shared library
#
# Source from setup scripts:
#   . "$(dirname "$0")/lib.sh"
#
# Or fetch and source in one-shot installers:
#   curl -fsSL "$RAW_BASE/setup/lib.sh" -o /tmp/audiot-lib.sh
#   . /tmp/audiot-lib.sh

# setup_getty_autologin <user>
# Configures getty tty1 to autologin as <user>.
setup_getty_autologin() {
    local user="$1"
    mkdir -p /etc/systemd/system/getty@tty1.service.d
    cat > /etc/systemd/system/getty@tty1.service.d/autologin.conf << EOF
[Service]
ExecStart=
ExecStart=-/sbin/agetty --autologin $user --noclear %I \$TERM
EOF
}

# setup_bash_profile <user>
# Appends X session launcher to ~/.bash_profile (idempotent via audiot-kiosk marker).
setup_bash_profile() {
    local user="$1"
    local profile="/home/$user/.bash_profile"
    if ! grep -q 'audiot-kiosk' "$profile" 2>/dev/null; then
        cat >> "$profile" << 'BASHEOF'

# audiot-kiosk: start X session on tty1 autologin
if [ -z "${DISPLAY:-}" ] && [ "$(tty)" = "/dev/tty1" ]; then
    exec startx "$HOME/.xinitrc" -- :0 vt1
fi
BASHEOF
        chown "$user:$user" "$profile" 2>/dev/null || true
    fi
}

# setup_xinitrc <user>
# Writes a minimal openbox .xinitrc. kiosk.sh handles DPMS separately.
setup_xinitrc() {
    local user="$1"
    local xinitrc="/home/$user/.xinitrc"
    cat > "$xinitrc" << 'XINITEOF'
#!/bin/sh
# Minimal X session for AUDiot kiosk
xset s off
xset -dpms
xset s noblank
openbox --config-file /dev/null &
exec true
XINITEOF
    chmod +x "$xinitrc"
    chown "$user:$user" "$xinitrc" 2>/dev/null || true
}

# install_udev_rule_ilitek
# Prevents ILITEK touchscreen USB autosuspend.
install_udev_rule_ilitek() {
    cat > /etc/udev/rules.d/99-ilitek-touch.rules << 'EOF'
ACTION=="add", SUBSYSTEM=="usb", ATTR{idVendor}=="222a", ATTR{power/control}="on"
EOF
    udevadm control --reload-rules
}

# extract_wl_gammarelay <image>
# Extracts /opt/audiot-dashboard/wl-gammarelay from a Docker image into
# /usr/local/bin/wl-gammarelay. Pulls the image if not already present.
extract_wl_gammarelay() {
    local image="$1"
    if [ -f /usr/local/bin/wl-gammarelay ]; then
        echo "[audiot-lib] wl-gammarelay already installed — skipping"
        return 0
    fi
    if docker image inspect "$image" >/dev/null 2>&1 || docker pull "$image" >/dev/null 2>&1; then
        docker run --rm --entrypoint cat "$image" \
            /opt/audiot-dashboard/wl-gammarelay > /usr/local/bin/wl-gammarelay
        chmod +x /usr/local/bin/wl-gammarelay
        echo "[audiot-lib] wl-gammarelay installed from $image"
    else
        echo "[audiot-lib] WARNING: could not pull $image — skipping wl-gammarelay install" >&2
        echo "[audiot-lib]   Retry: docker run --rm $image cat /opt/audiot-dashboard/wl-gammarelay > /usr/local/bin/wl-gammarelay" >&2
        return 1
    fi
}

# add_docker_group <user>
add_docker_group() {
    usermod -aG docker "$1" 2>/dev/null || true
}
