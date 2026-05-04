#!/usr/bin/env bash
set -euo pipefail

COMPOSE="docker compose --env-file .env"
PROJECT="monitor"

case "${1:-help}" in
  up)
    echo "Starting AUDiotMonitor collector-v2 stack..."
    $COMPOSE up -d
    echo ""
    echo "Services:"
    echo "  Netdata       → http://localhost:19999"
    echo "  Prometheus    → http://localhost:9090"
    echo "  Grafana       → http://localhost:3000  (admin / ${GRAFANA_ADMIN_PASSWORD:-admin})"
    echo "  llamaswap-exporter → http://localhost:9300/metrics"
    ;;
  down)
    $COMPOSE down
    echo "Stack stopped."
    ;;
  restart)
    $COMPOSE restart
    echo "Stack restarted."
    ;;
  logs)
    $COMPOSE logs -f "${2:-}"
    ;;
  status)
    $COMPOSE ps
    ;;
  validate)
    $COMPOSE config > /dev/null && echo "docker-compose.yml is valid."
    ;;
  update)
    $COMPOSE pull
    $COMPOSE up -d
    echo "Stack updated."
    ;;
  *)
    echo "Usage: $0 {up|down|restart|logs [service]|status|validate|update}"
    exit 1
    ;;
esac
