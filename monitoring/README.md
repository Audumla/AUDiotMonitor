# AUDiot Monitoring Stack

The monitoring stack consists of two primary components designed to run as Docker Compose services.

## Components

### 1. Collector Stack (`/monitoring/collector`)
The collector runs on every machine you want to monitor. It handles sensor data collection via `hwexp` and `node-exporter`, and stores it in a local Prometheus instance.
- **Key Features:** Automatic sensor discovery, custom Prometheus recording rules, system-specific GPU aliases.
- **Management:** `manage-collector.sh` (install, update, status, verify-metrics).

### 2. Dashboard Stack (`/monitoring/dashboard`)
The dashboard stack runs once (per display or per network) and provides the Grafana visualization layer.
- **Key Features:** Pre-baked dashboard profiles, self-sufficient Docker image, integrated kiosk mode controller.
- **Management:** `audiot-dashboard` (install, update, status, restart, logs, set).

## Deployment

### Remote Deployment (Recommended)
Use the `deploy-remote.sh` script from the repository root to push and update the stack on remote hosts via SSH and rsync.

### Local Installation
Use the `install-layout.sh` scripts in each component directory to create a host-owned bind-mounted layout.

For full instructions, see [**INSTALL.md**](./INSTALL.md).
