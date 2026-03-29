#!/usr/bin/env bash
# AUDiot Collector Layout Installer
#
# Creates a host-owned collector layout with editable hwexp and Prometheus
# config. Default recording rules are shipped with the repo; a system-specific
# custom rules file is generated only if one does not already exist.

set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/opt/docker/services/monitoring}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

info() { echo "[collector-layout] $*"; }

mkdir -p \
    "$INSTALL_DIR/config/hwexp" \
    "$INSTALL_DIR/config/hwexp/components" \
    "$INSTALL_DIR/config/hwexp/local/components" \
    "$INSTALL_DIR/config/hwexp/custom.d" \
    "$INSTALL_DIR/config/prometheus/rules/defaults" \
    "$INSTALL_DIR/config/prometheus/rules/custom" \
    "$INSTALL_DIR/hwexp" \
    "$INSTALL_DIR/prometheus"

copy_if_missing() {
    local src="$1"
    local dst="$2"
    if [ ! -e "$dst" ]; then
        cp "$src" "$dst"
        info "Installed $(realpath --relative-to="$INSTALL_DIR" "$dst" 2>/dev/null || echo "$dst")"
    fi
}

cp "$SCRIPT_DIR/docker-compose.yml" "$INSTALL_DIR/docker-compose.yml"
cp "$SCRIPT_DIR/install-layout.sh" "$INSTALL_DIR/install-layout.sh"
cp "$SCRIPT_DIR/manage-collector.sh" "$INSTALL_DIR/manage-collector.sh"
cp "$SCRIPT_DIR/generate-prometheus-custom-rules.py" "$INSTALL_DIR/generate-prometheus-custom-rules.py"
chmod +x "$INSTALL_DIR/generate-prometheus-custom-rules.py"
chmod +x "$INSTALL_DIR/install-layout.sh"
chmod +x "$INSTALL_DIR/manage-collector.sh"

copy_if_missing "$SCRIPT_DIR/.env.example" "$INSTALL_DIR/.env"
copy_if_missing "$SCRIPT_DIR/config/hwexp/hwexp.yaml" "$INSTALL_DIR/config/hwexp/hwexp.yaml"
copy_if_missing "$SCRIPT_DIR/config/hwexp/mappings.yaml" "$INSTALL_DIR/config/hwexp/mappings.yaml"
copy_if_missing "$SCRIPT_DIR/config/prometheus/prometheus.yml" "$INSTALL_DIR/config/prometheus/prometheus.yml"
copy_if_missing "$SCRIPT_DIR/config/prometheus/rules/defaults/audiot-recording-rules.yml" "$INSTALL_DIR/config/prometheus/rules/defaults/audiot-recording-rules.yml"
copy_if_missing "$SCRIPT_DIR/config/prometheus/rules/custom/system.rules.yml.example" "$INSTALL_DIR/config/prometheus/rules/custom/system.rules.yml.example"

if [ ! -e "$INSTALL_DIR/config/prometheus/rules/custom/system.rules.yml" ]; then
    INSTALL_DIR="$INSTALL_DIR" python3 "$INSTALL_DIR/generate-prometheus-custom-rules.py" --output "$INSTALL_DIR/config/prometheus/rules/custom/system.rules.yml"
fi

info "Collector layout ready in $INSTALL_DIR"
