#!/usr/bin/env bash
# Integration test: RPM package install on Fedora
# Usage: test-rpm-install.sh <path-to-rpm>
set -euo pipefail

RPM_PATH="${1:?Usage: $0 <path-to-rpm>}"
RPM_FILE=$(basename "$RPM_PATH")
PORT=9200
PASS=0
FAIL=0

pass() { echo "[PASS] $1"; PASS=$((PASS+1)); }
fail() { echo "[FAIL] $1"; FAIL=$((FAIL+1)); }

echo "=== RPM Install Integration Test ==="
echo "Package: $RPM_FILE"

# --- 1. File verification after install ---
dnf install -y "$RPM_PATH"

[ -x /usr/bin/hwexp ]                                  && pass "binary installed at /usr/bin/hwexp"        || fail "binary missing"
[ -f /etc/hwexp/hwexp.yaml ]                           && pass "config installed at /etc/hwexp/hwexp.yaml"  || fail "config missing"
[ -f /etc/hwexp/mappings.yaml ]                        && pass "mappings installed"                         || fail "mappings missing"
[ -f /etc/hwexp/sample_hwmon.json ]                    && pass "fixture installed"                          || fail "fixture missing"
[ -f /usr/lib/systemd/system/hwexp.service ]           && pass "service unit installed"                     || fail "service unit missing"

# --- 2. Start the binary directly (no systemd in container) ---
/usr/bin/hwexp \
  --config /etc/hwexp/hwexp.yaml \
  --fixture /etc/hwexp/sample_hwmon.json \
  > /tmp/hwexp.log 2>&1 &
HWEXP_PID=$!
echo "Started hwexp (PID $HWEXP_PID)"

# Wait for HTTP server to be ready
READY=0
for i in $(seq 1 10); do
  sleep 1
  if curl -sf "http://localhost:$PORT/healthz" > /dev/null 2>&1; then
    READY=1
    break
  fi
done

if [ "$READY" -eq 0 ]; then
  fail "service did not start within 10s"
  echo "--- hwexp log ---"
  cat /tmp/hwexp.log
  kill "$HWEXP_PID" 2>/dev/null || true
  exit 1
fi

# --- 3. Endpoint checks ---
HEALTH=$(curl -sf "http://localhost:$PORT/healthz")
echo "$HEALTH" | grep -q '"status":"ok"'                && pass "/healthz returns status ok"   || fail "/healthz bad response: $HEALTH"

VERSION=$(curl -sf "http://localhost:$PORT/version")
echo "$VERSION" | grep -q '"exporter_version"'          && pass "/version returns version info" || fail "/version bad response: $VERSION"

METRICS=$(curl -sf "http://localhost:$PORT/metrics")
echo "$METRICS" | grep -qE 'hwexp_up(\{[^}]*\})? 1'    && pass "/metrics contains hwexp_up 1"  || fail "/metrics bad response"

TEMP_READY=0
for i in $(seq 1 10); do
  METRICS=$(curl -sf "http://localhost:$PORT/metrics")
  if echo "$METRICS" | grep -q 'hw_device_temperature'; then
    TEMP_READY=1
    break
  fi
  sleep 1
done
[ "$TEMP_READY" -eq 1 ] && pass "/metrics contains mapped temp" || fail "/metrics missing temperature metric"

# --- 4. Uninstall and verify cleanup ---
kill "$HWEXP_PID" 2>/dev/null || true
dnf remove -y hwexp
[ ! -x /usr/bin/hwexp ]                                 && pass "binary removed on uninstall"   || fail "binary still present after uninstall"

# --- Summary ---
echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
