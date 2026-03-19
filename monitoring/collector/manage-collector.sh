#!/usr/bin/env bash
# AUDiot Collector Manager
#
# Reproducible entry point for installing, updating, validating, and operating
# the host-owned collector stack.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${INSTALL_DIR:-/opt/docker/collector}"

info() { echo "[collector-manage] $*"; }
die() { echo "[collector-manage] ERROR: $*" >&2; exit 1; }

run_in_install_dir() {
    (cd "$INSTALL_DIR" && "$@")
}

validate_yaml_tree() {
    python3 - <<'PY' "$INSTALL_DIR"
import sys, yaml
from pathlib import Path
root = Path(sys.argv[1])
for path in sorted(root.rglob("*.yml")) + sorted(root.rglob("*.yaml")):
    yaml.safe_load(path.read_text())
    print(f"OK {path}")
PY
}

validate_collector_layout() {
    [ -f "$INSTALL_DIR/docker-compose.yml" ] || die "Missing $INSTALL_DIR/docker-compose.yml"
    [ -f "$INSTALL_DIR/config/prometheus/prometheus.yml" ] || die "Missing Prometheus config"
    [ -f "$INSTALL_DIR/config/hwexp/hwexp.yaml" ] || die "Missing hwexp config"
    validate_yaml_tree
    if command -v docker >/dev/null 2>&1; then
        run_in_install_dir docker compose config >/dev/null
    fi
    info "Validation passed"
}

case "${1:-}" in
    install)
        INSTALL_DIR="$INSTALL_DIR" "$SCRIPT_DIR/install-layout.sh"
        ;;
    update)
        INSTALL_DIR="$INSTALL_DIR" "$SCRIPT_DIR/install-layout.sh"
        validate_collector_layout
        ;;
    validate)
        validate_collector_layout
        ;;
    generate-rules)
        run_in_install_dir env INSTALL_DIR="$INSTALL_DIR" python3 ./generate-prometheus-custom-rules.py "${@:2}"
        ;;
    up)
        run_in_install_dir docker compose up -d
        ;;
    restart-prometheus)
        run_in_install_dir docker compose restart prometheus
        ;;
    restart-hwexp)
        run_in_install_dir docker compose restart hwexp
        ;;
    status)
        run_in_install_dir docker compose ps
        ;;
    logs)
        service="${2:-prometheus}"
        run_in_install_dir docker compose logs --tail=100 "$service"
        ;;
    *)
        cat <<'EOF'
Usage:
  ./manage-collector.sh install
  ./manage-collector.sh update
  ./manage-collector.sh validate
  ./manage-collector.sh generate-rules [--force]
  ./manage-collector.sh up
  ./manage-collector.sh restart-prometheus
  ./manage-collector.sh restart-hwexp
  ./manage-collector.sh status
  ./manage-collector.sh logs [service]
EOF
        exit 1
        ;;
esac
