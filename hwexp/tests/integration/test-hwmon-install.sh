#!/usr/bin/env bash
# Integration test: install DEB then verify real hwmon adapter reads sensor data.
# Mounts a fake sysfs tree so the test works anywhere without real hardware.
# Usage: test-hwmon-install.sh <path-to-deb> <path-to-repo-root>
set -euo pipefail

DEB_PATH="${1:?Usage: $0 <path-to-deb> <path-to-repo-root>}"
REPO_ROOT="${2:?Usage: $0 <path-to-deb> <path-to-repo-root>}"
PORT=9200
PASS=0
FAIL=0

pass() { echo "[PASS] $1"; PASS=$((PASS+1)); }
fail() { echo "[FAIL] $1"; FAIL=$((FAIL+1)); }

docker_path() {
  local p="$1"
  if echo "$p" | grep -qE '^[A-Za-z]:'; then
    echo "$p" | sed 's|^\([A-Za-z]\):|//\L\1|; s|\\|/|g'
  else
    realpath "$p"
  fi
}

DEB_DOCKER="$(docker_path "$DEB_PATH")"
REPO_DOCKER="$(docker_path "$REPO_ROOT")"

echo "=== hwmon Adapter Integration Test ==="

docker run --rm \
  -v "${DEB_DOCKER}:/tmp/hwexp.deb:ro" \
  -v "${REPO_DOCKER}/hwexp/tests/fixtures/sysfs:/test/hwmon:ro" \
  -v "${REPO_DOCKER}/hwexp/configs/mappings.yaml:/etc/hwexp/mappings.yaml:ro" \
  ubuntu:22.04 \
  bash -c "
    apt-get update -qq && apt-get install -y -qq curl 2>/dev/null

    # Install package
    dpkg -i /tmp/hwexp.deb

    # Write a test config that points hwmon adapter at our fake sysfs
    cat > /tmp/hwexp-test.yaml << 'EOF'
server:
  listen_address: \"0.0.0.0:${PORT}\"
  refresh_interval: 2s
  discovery_interval: 10s
identity:
  host: test-host
  platform: linux
adapters:
  linux_hwmon:
    enabled: true
    settings:
      hwmon_path: /test/hwmon
mapping:
  rules_file: /etc/hwexp/mappings.yaml
  strict_mode: false
debug:
  enable_raw_endpoint: true
EOF

    # Start hwexp with real hwmon adapter
    /usr/bin/hwexp --config /tmp/hwexp-test.yaml > /tmp/hwexp.log 2>&1 &
    HWEXP_PID=\$!

    # Wait for ready
    for i in \$(seq 1 10); do
      sleep 1
      curl -sf http://localhost:${PORT}/healthz > /dev/null 2>&1 && break
    done

    # --- Checks ---
    HEALTH=\$(curl -sf http://localhost:${PORT}/healthz)
    echo \"\$HEALTH\" | grep -q '\"status\":\"ok\"' && echo '[PASS] /healthz ok' || echo '[FAIL] /healthz bad'

    DISCOVERY=\$(curl -sf http://localhost:${PORT}/debug/discovery)
    echo \"\$DISCOVERY\" | grep -q 'amdgpu'    && echo '[PASS] amdgpu device discovered'  || echo '[FAIL] amdgpu not found in discovery'
    echo \"\$DISCOVERY\" | grep -q 'k10temp'   && echo '[PASS] k10temp device discovered' || echo '[FAIL] k10temp not found in discovery'
    echo \"\$DISCOVERY\" | grep -q 'nct6775'   && echo '[PASS] nct6775 device discovered' || echo '[FAIL] nct6775 not found in discovery'

    RAW=\$(curl -sf 'http://localhost:${PORT}/debug/raw')
    echo \"\$RAW\" | grep -q 'temp1_input'     && echo '[PASS] raw temp readings present'  || echo '[FAIL] no raw temp readings'
    echo \"\$RAW\" | grep -q 'fan1_input'      && echo '[PASS] raw fan readings present'   || echo '[FAIL] no raw fan readings'
    echo \"\$RAW\" | grep -q 'power1_input'    && echo '[PASS] raw power readings present' || echo '[FAIL] no raw power readings'

    METRICS=\$(curl -sf http://localhost:${PORT}/metrics)
    echo \"\$METRICS\" | grep -q 'hwexp_up 1'              && echo '[PASS] /metrics hwexp_up=1'           || echo '[FAIL] hwexp_up missing'
    echo \"\$METRICS\" | grep -q 'hw_device_temperature'   && echo '[PASS] temperature metric exported'   || echo '[FAIL] temperature metric missing'

    echo ''
    echo '--- Last 15 lines of hwexp log ---'
    tail -15 /tmp/hwexp.log

    kill \$HWEXP_PID 2>/dev/null || true
  " 2>&1

echo ""
echo "=== Done ==="
