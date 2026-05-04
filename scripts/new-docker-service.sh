#!/usr/bin/env bash
# new-docker-service - Add a new service to an existing compose.yaml
#
# Usage:
#   new-docker-service.sh <group> <image-name>
#
# Environment:
#   DOCKER_USER       - Linux user for ownership (default: $SUDO_USER or current user)
#   DOCKER_GROUP      - Linux group for ownership (default: same as DOCKER_USER)
#   BASE_DIR          - Base directory for compose projects (default: /srv/docker)
#   TZ                - Timezone for containers (default: UTC)
#   PUID              - User ID for containers (default: 1000)
#   PGID              - Group ID for containers (default: 1000)
#
# Example:
#   new-docker-service.sh media linuxserver/plex

set -e

if [[ -n "${DOCKER_USER:-}" ]]; then
    : # DOCKER_USER already set
elif [[ -n "${SUDO_USER:-}" ]]; then
    DOCKER_USER="$SUDO_USER"
else
    DOCKER_USER="$(whoami)"
fi

if [[ -n "${DOCKER_GROUP:-}" ]]; then
    : # DOCKER_GROUP already set
else
    DOCKER_GROUP="${DOCKER_USER}"
fi

BASE_DIR="${BASE_DIR:-/srv/docker}"
TZ="${TZ:-UTC}"
PUID="${PUID:-1000}"
PGID="${PGID:-1000}"

if [ $# -ne 2 ]; then
    echo "Usage: $0 <group> <image-name>"
    echo "Example: $0 media linuxserver/plex"
    exit 1
fi

GROUP_NAME="$1"
IMAGE="$2"
SERVICE_NAME=$(basename "$IMAGE" | cut -d':' -f1 | tr '/:' '__')
GROUP_DIR="$BASE_DIR/$GROUP_NAME"
CONFIG_BASE="$GROUP_DIR/config/$SERVICE_NAME"
COMPOSE_FILE="$GROUP_DIR/compose.yaml"
NETWORK_NAME="${GROUP_NAME}_network"

# Create config and data directories
mkdir -p "$CONFIG_BASE/config"
mkdir -p "$CONFIG_BASE/data"
chown -R "${DOCKER_USER}:${DOCKER_GROUP}" "$CONFIG_BASE"

# Create compose.yaml if it doesn't exist
if [ ! -f "$COMPOSE_FILE" ]; then
    cat > "$COMPOSE_FILE" <<EOF
services:

networks:
  ${NETWORK_NAME}:
    driver: bridge
EOF
    chown "${DOCKER_USER}:${DOCKER_GROUP}" "$COMPOSE_FILE"
fi

# Check for existing service entry
if grep -q "  ${SERVICE_NAME}:" "$COMPOSE_FILE"; then
    echo "Service '${SERVICE_NAME}' already exists in $COMPOSE_FILE"
    exit 1
fi

# Insert new service block after 'services:' key
TEMP_FILE=$(mktemp)
awk -v service="$SERVICE_NAME" -v image="$IMAGE" -v group="$GROUP_NAME" \
    -v tz="$TZ" -v puid="$PUID" -v pgid="$PGID" '
  BEGIN { added = 0 }
  /^services:/ {
    print
    print "  " service ":"
    print "    container_name: " service
    print "    image: " image
    print "    restart: unless-stopped"
    print "    volumes:"
    print "      # - ./config/" service "/config:/<container-config-path>"
    print "      # - ./config/" service "/data:/<container-data-path>"
    print "    environment:"
    print "      TZ: \"" tz "\""
    print "      PUID: \"" puid "\""
    print "      PGID: \"" pgid "\""
    print "    networks:"
    print "      - " group "_network"
    print ""
    added = 1
    next
  }
  { print }
' "$COMPOSE_FILE" > "$TEMP_FILE"

mv "$TEMP_FILE" "$COMPOSE_FILE"
chown "${DOCKER_USER}:${DOCKER_GROUP}" "$COMPOSE_FILE"

echo "Service '${SERVICE_NAME}' added with config at $CONFIG_BASE"
