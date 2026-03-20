#!/usr/bin/env bash
# AUDiot Monitor — Remote Deployment Tool
#
# Pushes the monitoring stack to a remote host via rsync and executes the
# corresponding management script (install/update).
#
# Usage:
#   ./deploy-remote.sh <host> <component: collector|dashboard> [target_dir]
#
# Example:
#   ./deploy-remote.sh brutusview dashboard /opt/docker/dashboard

set -euo pipefail

HOST="${1:-}"
COMPONENT="${2:-}"
INSTALL_DIR="${3:-/opt/docker/$COMPONENT}"

if [[ -z "$HOST" || -z "$COMPONENT" ]]; then
    echo "Usage: $0 <host> <component> [target_dir]"
    exit 1
fi

if [[ "$COMPONENT" != "collector" && "$COMPONENT" != "dashboard" ]]; then
    echo "Error: component must be 'collector' or 'dashboard'"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_DIR="$SCRIPT_DIR/$COMPONENT/"

info() { echo "[deploy] $*"; }

info "Syncing $COMPONENT to $HOST:$INSTALL_DIR..."
# Ensure target parent exists
ssh "$HOST" "sudo mkdir -p \"$INSTALL_DIR\" && sudo chown \$USER \"$INSTALL_DIR\""

# Sync files
rsync -avz --delete \
    --exclude ".git" \
    --exclude "node_modules" \
    --exclude "*.log" \
    "$SOURCE_DIR" \
    "$HOST:$INSTALL_DIR/"

info "Executing remote update..."
ssh "$HOST" "INSTALL_DIR=\"$INSTALL_DIR\" bash \"$INSTALL_DIR/manage-$COMPONENT.sh\" update"

info "Deployment to $HOST complete."
