#!/usr/bin/env bash
# Unit tests for monitoring/dashboard/setup/lib.sh
#
# Runs in a debian:bookworm-slim container — no system daemons required.
# Commands that require running services (udevadm, usermod) are overridden
# with no-op mock functions before each relevant test group.
#
# Local usage:
#   docker run --rm \
#     -v "$(pwd)/monitoring/dashboard/setup:/setup:ro" \
#     debian:bookworm-slim bash /setup/test-lib.sh
#
# Exit code: 0 = all passed, 1 = one or more failures.

set -uo pipefail

PASS=0; FAIL=0; SKIP=0

pass() { echo "[PASS] $1"; (( PASS++ )) || true; }
fail() { echo "[FAIL] $1"; (( FAIL++ )) || true; }
skip() { echo "[SKIP] $1 ($2)"; (( SKIP++ )) || true; }

echo "=== lib.sh tests ==="
echo "Running as: $(id)"
echo ""

# ── Load library ──────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

# ── Mocks: system commands unavailable in a minimal container ─────────────────
# Defined after sourcing so they shadow the real commands at call time.
udevadm() { :; }   # no udev daemon in container
usermod()  { :; }  # no real user management in the test env

# ── Fixtures ──────────────────────────────────────────────────────────────────
TUSER="kiosktest"
THOME="/home/$TUSER"
mkdir -p "$THOME"

# Ensure required /etc directories exist (bookworm-slim has them, but be safe)
mkdir -p /etc/systemd/system
mkdir -p /etc/udev/rules.d

# ── 1. setup_getty_autologin ─────────────────────────────────────────────────
echo "--- 1. setup_getty_autologin ---"
setup_getty_autologin "$TUSER"

CONF=/etc/systemd/system/getty@tty1.service.d/autologin.conf
[ -f "$CONF" ] \
    && pass "autologin.conf created" \
    || fail "autologin.conf not created"
grep -q "autologin $TUSER" "$CONF" 2>/dev/null \
    && pass "conf references correct user ($TUSER)" \
    || fail "conf has wrong user; expected '$TUSER'"
grep -q 'ExecStart=$' "$CONF" 2>/dev/null \
    && pass "conf clears default ExecStart" \
    || fail "conf missing bare ExecStart= reset line"

# idempotency: second call must not duplicate entries
setup_getty_autologin "$TUSER"
COUNT=$(grep -c "autologin $TUSER" "$CONF")
[ "$COUNT" -eq 1 ] \
    && pass "idempotent (no duplicate autologin line after second call)" \
    || fail "duplicate entries after second call (count=$COUNT)"

# ── 2. setup_bash_profile ────────────────────────────────────────────────────
echo ""
echo "--- 2. setup_bash_profile ---"
PROFILE="$THOME/.bash_profile"
setup_bash_profile "$TUSER"

[ -f "$PROFILE" ] \
    && pass ".bash_profile created" \
    || fail ".bash_profile not created"
grep -q 'audiot-kiosk' "$PROFILE" \
    && pass "contains audiot-kiosk idempotency marker" \
    || fail "missing audiot-kiosk marker"
grep -q 'startx' "$PROFILE" \
    && pass "contains startx invocation" \
    || fail "missing startx invocation"
grep -q '/dev/tty1' "$PROFILE" \
    && pass "guarded to tty1 only" \
    || fail "missing tty1 guard"
grep -q 'DISPLAY' "$PROFILE" \
    && pass "checks DISPLAY before launching X" \
    || fail "missing DISPLAY check"

# idempotency: second call must not add a duplicate block
setup_bash_profile "$TUSER"
COUNT=$(grep -c 'audiot-kiosk' "$PROFILE")
[ "$COUNT" -eq 1 ] \
    && pass "idempotent (marker appears exactly once after second call)" \
    || fail "duplicate block after second call (marker count=$COUNT)"

# ── 3. setup_xinitrc ─────────────────────────────────────────────────────────
echo ""
echo "--- 3. setup_xinitrc ---"
setup_xinitrc "$TUSER"
XINITRC="$THOME/.xinitrc"

[ -f "$XINITRC" ] \
    && pass ".xinitrc created" \
    || fail ".xinitrc not created"
[ -x "$XINITRC" ] \
    && pass ".xinitrc is executable" \
    || fail ".xinitrc not executable"
grep -q 'openbox' "$XINITRC" \
    && pass "contains openbox launch" \
    || fail "missing openbox"
grep -q 'xset s off' "$XINITRC" \
    && pass "disables screensaver (xset s off)" \
    || fail "missing xset s off"
grep -q 'xset -dpms' "$XINITRC" \
    && pass "disables DPMS" \
    || fail "missing xset -dpms"
grep -q 'xset s noblank' "$XINITRC" \
    && pass "disables screen blanking" \
    || fail "missing xset s noblank"

# overwrite is intentional (not idempotency issue): verify file is still valid
setup_xinitrc "$TUSER"
[ -x "$XINITRC" ] && grep -q 'openbox' "$XINITRC" \
    && pass "safe to call multiple times (file still valid)" \
    || fail "file corrupted after second call"

# ── 4. install_udev_rule_ilitek ──────────────────────────────────────────────
echo ""
echo "--- 4. install_udev_rule_ilitek ---"
install_udev_rule_ilitek
RULE=/etc/udev/rules.d/99-ilitek-touch.rules

[ -f "$RULE" ] \
    && pass "rule file created" \
    || fail "rule file not created"
grep -q '222a' "$RULE" \
    && pass "contains ILITEK vendor ID (222a)" \
    || fail "missing ILITEK vendor ID"
grep -q 'power/control' "$RULE" \
    && pass "rule sets power/control attribute" \
    || fail "missing power/control attribute"
grep -q 'SUBSYSTEM=="usb"' "$RULE" \
    && pass "rule targets USB subsystem" \
    || fail "rule missing USB subsystem match"

# idempotency: udev rules file is overwritten (one rule = correct behaviour)
install_udev_rule_ilitek
COUNT=$(grep -c '222a' "$RULE")
[ "$COUNT" -eq 1 ] \
    && pass "idempotent (single rule entry after second call)" \
    || fail "duplicate rule entries (count=$COUNT)"

# ── 5. add_docker_group ──────────────────────────────────────────────────────
echo ""
echo "--- 5. add_docker_group ---"
add_docker_group "$TUSER" \
    && pass "returns 0 (usermod mocked)" \
    || fail "returned non-zero"

# ── 6. extract_wl_gammarelay ─────────────────────────────────────────────────
echo ""
echo "--- 6. extract_wl_gammarelay ---"
skip "extract_wl_gammarelay" "requires Docker daemon — covered by install integration test"

# ── 7. lib.sh syntax check ───────────────────────────────────────────────────
echo ""
echo "--- 7. syntax check ---"
bash -n "$SCRIPT_DIR/lib.sh" \
    && pass "lib.sh passes bash -n syntax check" \
    || fail "lib.sh has syntax errors"
bash -n "$SCRIPT_DIR/install.sh" \
    && pass "install.sh passes bash -n syntax check" \
    || fail "install.sh has syntax errors"
bash -n "$SCRIPT_DIR/dietpi-setup.sh" \
    && pass "dietpi-setup.sh passes bash -n syntax check" \
    || fail "dietpi-setup.sh has syntax errors"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
