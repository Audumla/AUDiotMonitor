#!/usr/bin/env bash
# Two-container smoke test: emitter (hwexp) + client (curl scraper)
# Verifies a remote client can scrape metrics from an installed hwexp instance.
# Usage: test-network-smoke.sh <path-to-deb>
set -euo pipefail

DEB_PATH="${1:?Usage: $0 <path-to-deb>}"
NETWORK="hwexp-test-net"
EMITTER="hwexp-emitter"
PORT=9200
PASS=0
FAIL=0

pass() { echo "[PASS] $1"; PASS=$((PASS+1)); }
fail() { echo "[FAIL] $1"; FAIL=$((FAIL+1)); }

# Convert path to Docker-compatible absolute bind-mount path.
# On Windows/Git Bash: H:\foo -> //h/foo  (Git Bash strips one slash, Docker sees /h/foo)
# On Linux: resolves to absolute path via realpath
docker_path() {
  local p="$1"
  if echo "$p" | grep -qE '^[A-Za-z]:'; then
    echo "$p" | sed 's|^\([A-Za-z]\):|//\L\1|; s|\\|/|g'
  else
    realpath "$p"
  fi
}

DEB_DOCKER="$(docker_path "$DEB_PATH")"

cleanup() {
  echo "--- Cleaning up ---"
  docker rm -f "$EMITTER" 2>/dev/null || true
  docker network rm "$NETWORK" 2>/dev/null || true
}
trap cleanup EXIT

echo "=== Two-Container Network Smoke Test ==="
echo "DEB: $DEB_DOCKER"

# --- 1. Create isolated Docker network ---
docker network create "$NETWORK"
pass "created test network $NETWORK"

# --- 2. Start emitter container (Ubuntu, install DEB, run hwexp) ---
docker run -d \
  --name "$EMITTER" \
  --network "$NETWORK" \
  -v "${DEB_DOCKER}:/tmp/hwexp.deb:ro" \
  ubuntu:22.04 \
  bash -c "
    apt-get update -qq &&
    apt-get install -y -qq curl 2>/dev/null &&
    dpkg -i /tmp/hwexp.deb &&
    /usr/bin/hwexp \
      --config /etc/hwexp/hwexp.yaml \
      --fixture /etc/hwexp/sample_hwmon.json
  "
pass "emitter container started"

# --- 3. Wait for emitter to be ready (poll from client side) ---
echo "Waiting for emitter to be ready..."
READY=0
for i in $(seq 1 15); do
  sleep 2
  if docker run --rm --network "$NETWORK" curlimages/curl:latest \
      -sf "http://$EMITTER:$PORT/healthz" > /dev/null 2>&1; then
    READY=1
    break
  fi
  echo "  attempt $i/15..."
done

if [ "$READY" -eq 0 ]; then
  fail "emitter did not become reachable within 30s"
  echo "--- emitter logs ---"
  docker logs "$EMITTER"
  exit 1
fi
pass "emitter reachable from client container"

# --- 4. Client scrapes each endpoint from a separate container ---
run_client() {
  docker run --rm --network "$NETWORK" curlimages/curl:latest \
    -sf "http://$EMITTER:$PORT/$1"
}

HEALTH=$(run_client "healthz")
echo "$HEALTH" | grep -q '"status":"ok"'          && pass "client: /healthz ok"          || fail "client: /healthz bad: $HEALTH"

VERSION=$(run_client "version")
echo "$VERSION" | grep -q '"exporter_version"'    && pass "client: /version ok"          || fail "client: /version bad: $VERSION"

METRICS=$(run_client "metrics")
echo "$METRICS" | grep -qE 'hwexp_up(\{[^}]*\})? 1' && pass "client: /metrics hwexp_up 1"  || fail "client: /metrics missing hwexp_up"
TEMP_READY=0
for i in $(seq 1 10); do
  METRICS=$(run_client "metrics")
  if echo "$METRICS" | grep -q 'hw_device_temperature'; then
    TEMP_READY=1
    break
  fi
  sleep 1
done
[ "$TEMP_READY" -eq 1 ] && pass "client: /metrics has temp" || fail "client: /metrics missing temperature"

DISCOVERY=$(run_client "debug/discovery")
echo "$DISCOVERY" | grep -q '"device_count"'      && pass "client: /debug/discovery ok"  || fail "client: /debug/discovery bad: $DISCOVERY"

# --- Summary ---
echo ""
echo "--- emitter logs ---"
docker logs "$EMITTER" 2>&1 | tail -20

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
