#!/usr/bin/env bash
# Cross-distro compatibility matrix test
# Installs hwexp on each emitter distro, then scrapes from each client distro.
# Usage: test-cross-distro.sh <path-to-deb> <path-to-rpm>
set -euo pipefail

DEB_PATH="${1:?Usage: $0 <path-to-deb> <path-to-rpm>}"
RPM_PATH="${2:?Usage: $0 <path-to-deb> <path-to-rpm>}"

NETWORK="hwexp-cross-test"
PORT=9200
TOTAL_PASS=0
TOTAL_FAIL=0

# Convert path to Docker-compatible mount format (handles Windows Git Bash paths)
docker_path() {
  echo "$1" | sed 's|^\([A-Za-z]\):|//\L\1|; s|\\|/|g'
}

DEB_DOCKER="$(docker_path "$DEB_PATH")"
RPM_DOCKER="$(docker_path "$RPM_PATH")"

# --- Emitter definitions: name, image, package, install command ---
# Format: "name|image|pkg_docker_path|install_cmd"
EMITTERS=(
  "ubuntu-22.04|ubuntu:22.04|${DEB_DOCKER}|apt-get update -qq && apt-get install -y -qq curl && dpkg -i /tmp/pkg"
  "ubuntu-24.04|ubuntu:24.04|${DEB_DOCKER}|apt-get update -qq && apt-get install -y -qq curl && dpkg -i /tmp/pkg"
  "debian-12|debian:12|${DEB_DOCKER}|apt-get update -qq && apt-get install -y -qq curl && dpkg -i /tmp/pkg"
  "fedora-latest|fedora:latest|${RPM_DOCKER}|dnf install -y curl && rpm -ivh /tmp/pkg"
  "rockylinux-9|rockylinux:9|${RPM_DOCKER}|rpm -ivh /tmp/pkg"
  "opensuse-tumbleweed|opensuse/tumbleweed|${RPM_DOCKER}|rpm -ivh /tmp/pkg"
)

# --- Client definitions: name, image, sh command prefix to ensure curl available ---
CLIENTS=(
  "ubuntu-22.04|ubuntu:22.04|apt-get update -qq && apt-get install -y -qq curl 2>/dev/null && curl"
  "alpine|alpine:latest|apk add --no-cache curl -q 2>/dev/null && curl"
  "fedora-latest|fedora:latest|curl"
  "debian-12|debian:12|apt-get update -qq && apt-get install -y -qq curl 2>/dev/null && curl"
)

pass()  { echo "  [PASS] $1"; TOTAL_PASS=$((TOTAL_PASS+1)); }
fail()  { echo "  [FAIL] $1"; TOTAL_FAIL=$((TOTAL_FAIL+1)); }
header(){ echo ""; echo ">>> $1"; }

cleanup_all() {
  docker rm -f hwexp-emitter 2>/dev/null || true
  docker network rm "$NETWORK" 2>/dev/null || true
}
trap cleanup_all EXIT

# Create shared test network once
docker network rm "$NETWORK" 2>/dev/null || true
docker network create "$NETWORK" > /dev/null

echo "========================================"
echo " hwexp Cross-Distro Compatibility Matrix"
echo "========================================"
echo "Emitters: ${#EMITTERS[@]}  Clients: ${#CLIENTS[@]}"

for emitter_def in "${EMITTERS[@]}"; do
  IFS='|' read -r E_NAME E_IMAGE E_PKG E_INSTALL <<< "$emitter_def"
  header "EMITTER: $E_NAME ($E_IMAGE)"

  # Remove any previous emitter
  docker rm -f hwexp-emitter 2>/dev/null || true

  # Start emitter
  docker run -d \
    --name hwexp-emitter \
    --network "$NETWORK" \
    -v "${E_PKG}:/tmp/pkg:ro" \
    "$E_IMAGE" \
    bash -c "
      ${E_INSTALL} &&
      /usr/bin/hwexp --config /etc/hwexp/hwexp.yaml --fixture /etc/hwexp/sample_hwmon.json
    " > /dev/null

  # Wait for it to be ready (poll with alpine/curl — lightweight)
  READY=0
  for i in $(seq 1 15); do
    sleep 2
    if docker run --rm --network "$NETWORK" alpine:latest \
        sh -c "apk add --no-cache curl -q 2>/dev/null && curl -sf http://hwexp-emitter:${PORT}/healthz" > /dev/null 2>&1; then
      READY=1
      break
    fi
  done

  if [ "$READY" -eq 0 ]; then
    fail "$E_NAME: emitter failed to start"
    echo "  --- emitter logs ---"
    docker logs hwexp-emitter 2>&1 | tail -10 | sed 's/^/  /'
    continue
  fi
  pass "$E_NAME: emitter started and listening on :$PORT"

  # Run each client against this emitter
  for client_def in "${CLIENTS[@]}"; do
    IFS='|' read -r C_NAME C_IMAGE C_CURL_CMD <<< "$client_def"

    echo "  -- client: $C_NAME"

    # Run client checks
    RESULT=$(docker run --rm --network "$NETWORK" "$C_IMAGE" \
      sh -c "${C_CURL_CMD} -sf http://hwexp-emitter:${PORT}/healthz" 2>/dev/null || echo "FAIL")

    if echo "$RESULT" | grep -q '"status":"ok"'; then
      pass "$E_NAME → $C_NAME: /healthz ok"
    else
      fail "$E_NAME → $C_NAME: /healthz failed ($RESULT)"
      continue
    fi

    METRICS=$(docker run --rm --network "$NETWORK" "$C_IMAGE" \
      sh -c "${C_CURL_CMD} -sf http://hwexp-emitter:${PORT}/metrics" 2>/dev/null || echo "FAIL")

    if echo "$METRICS" | grep -q 'hwexp_up 1'; then
      pass "$E_NAME → $C_NAME: /metrics ok (hwexp_up=1)"
    else
      fail "$E_NAME → $C_NAME: /metrics failed"
    fi

    if echo "$METRICS" | grep -q 'hw_device_temperature'; then
      pass "$E_NAME → $C_NAME: /metrics has temperature data"
    else
      fail "$E_NAME → $C_NAME: /metrics missing temperature data"
    fi
  done

  docker rm -f hwexp-emitter > /dev/null 2>&1 || true
done

echo ""
echo "========================================"
echo " RESULTS: $TOTAL_PASS passed, $TOTAL_FAIL failed"
echo "========================================"
[ "$TOTAL_FAIL" -eq 0 ] && exit 0 || exit 1
