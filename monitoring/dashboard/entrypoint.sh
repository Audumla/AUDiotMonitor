#!/usr/bin/env bash
# AUDiot Dashboard Image Entrypoint
#
# Ensures the dashboard directory in the volume is initialized with templates
# and a 'custom' directory before starting Grafana.

set -euo pipefail

DASHBOARD_VOL="/var/lib/grafana/dashboards"
FACTORY_LIB="/opt/audiot-dashboard/dashboards"

info() { echo "[audiot-init] $*"; }

# 1. Create the base layout if it's a fresh volume
mkdir -p \
    "$DASHBOARD_VOL/profiles/standard" \
    "$DASHBOARD_VOL/profiles/wide-screens" \
    "$DASHBOARD_VOL/profiles/mobile" \
    "$DASHBOARD_VOL/profiles/debug" \
    "$DASHBOARD_VOL/custom"

# 2. Copy "factory" versions of dashboards if missing (don't overwrite user edits)
if [ -d "$FACTORY_LIB" ]; then
    info "Initializing dashboard profiles from image library..."
    # -n: do not overwrite an existing file
    # -r: recursive
    cp -rn "$FACTORY_LIB/profiles/"* "$DASHBOARD_VOL/profiles/"
fi

# 3. Create placeholder for user custom dashboards if totally empty
if [ -z "$(ls -A "$DASHBOARD_VOL/custom" 2>/dev/null || true)" ] && [ -d "/opt/audiot-dashboard/examples/custom" ]; then
    info "Initializing example custom dashboards..."
    cp -rn /opt/audiot-dashboard/examples/custom/* "$DASHBOARD_VOL/custom/"
fi

# 4. Copy kiosk.sh into the volume root so host operators can find it easily.
cp /opt/audiot-dashboard/kiosk.sh "$DASHBOARD_VOL/kiosk.sh"
chmod +x "$DASHBOARD_VOL/kiosk.sh"

# 5. Fix permissions so Grafana can write/read
# This is needed if the volume was mounted as root
chown -R grafana:grafana "$DASHBOARD_VOL"

info "Initialization complete. Starting Grafana..."

# Execute the original Grafana entrypoint
exec /run.sh "$@"
