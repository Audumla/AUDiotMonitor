#!/usr/bin/env bash
# AUDiot Dashboard Layout Installer
#
# Creates a bind-mounted dashboard layout in a target directory and copies the
# default Grafana config, kiosk config, and starter dashboards if missing.

set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/opt/docker/services/dashboard}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

info() { echo "[layout] $*"; }

mkdir -p \
    "$INSTALL_DIR/config/grafana/provisioning/datasources" \
    "$INSTALL_DIR/config/grafana/provisioning/dashboards" \
    "$INSTALL_DIR/dashboards/profiles/standard" \
    "$INSTALL_DIR/dashboards/profiles/wide-screens" \
    "$INSTALL_DIR/dashboards/profiles/mobile" \
    "$INSTALL_DIR/dashboards/profiles/debug" \
    "$INSTALL_DIR/dashboards/custom"

copy_if_missing() {
    local src="$1"
    local dst="$2"
    if [ ! -e "$dst" ]; then
        cp "$src" "$dst"
        info "Installed $(realpath --relative-to="$INSTALL_DIR" "$dst" 2>/dev/null || echo "$dst")"
    fi
}

copy_if_exists() {
    local src="$1"
    local dst="$2"
    if [ -e "$src" ]; then
        cp "$src" "$dst"
    fi
}

cp "$SCRIPT_DIR/docker-compose.yml" "$INSTALL_DIR/docker-compose.yml"
cp "$SCRIPT_DIR/kiosk.sh" "$INSTALL_DIR/kiosk.sh"
copy_if_exists "$SCRIPT_DIR/Dockerfile.kiosk" "$INSTALL_DIR/Dockerfile.kiosk"
copy_if_exists "$SCRIPT_DIR/.env.example" "$INSTALL_DIR/.env.example"
cp "$SCRIPT_DIR/install-layout.sh" "$INSTALL_DIR/install-layout.sh"
copy_if_exists "$SCRIPT_DIR/manage-dashboard.sh" "$INSTALL_DIR/manage-dashboard.sh"
copy_if_exists "$SCRIPT_DIR/set-dashboard.sh" "$INSTALL_DIR/set-dashboard.sh"
chmod +x "$INSTALL_DIR/kiosk.sh"
chmod +x "$INSTALL_DIR/install-layout.sh"
if [ -e "$INSTALL_DIR/manage-dashboard.sh" ]; then chmod +x "$INSTALL_DIR/manage-dashboard.sh"; fi
if [ -e "$INSTALL_DIR/set-dashboard.sh" ]; then chmod +x "$INSTALL_DIR/set-dashboard.sh"; fi

copy_if_missing "$SCRIPT_DIR/kiosk-install.sh" "$INSTALL_DIR/kiosk-install.sh"
chmod +x "$INSTALL_DIR/kiosk-install.sh"

copy_if_missing "$SCRIPT_DIR/config/grafana/grafana.ini" "$INSTALL_DIR/config/grafana/grafana.ini"
copy_if_missing "$SCRIPT_DIR/config/grafana/provisioning/datasources/prometheus.yaml" "$INSTALL_DIR/config/grafana/provisioning/datasources/prometheus.yaml"
copy_if_missing "$SCRIPT_DIR/config/grafana/provisioning/datasources/infinity.yaml" "$INSTALL_DIR/config/grafana/provisioning/datasources/infinity.yaml"
copy_if_missing "$SCRIPT_DIR/config/grafana/provisioning/dashboards/dashboards.yaml" "$INSTALL_DIR/config/grafana/provisioning/dashboards/dashboards.yaml"
copy_if_missing "$SCRIPT_DIR/config/kiosk.env.example" "$INSTALL_DIR/config/kiosk.env"

while IFS= read -r src; do
    rel="${src#$SCRIPT_DIR/}"
    dst="$INSTALL_DIR/$rel"
    copy_if_missing "$src" "$dst"
done < <(find "$SCRIPT_DIR/dashboards/profiles" -type f | sort)

while IFS= read -r src; do
    dst="$INSTALL_DIR/dashboards/custom/$(basename "$src")"
    copy_if_missing "$src" "$dst"
done < <(find "$SCRIPT_DIR/examples/custom" -type f | sort)

info "Dashboard layout ready in $INSTALL_DIR"
