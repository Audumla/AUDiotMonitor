# Hardware Telemetry Exporter Platform - Installation Specification

**Document purpose**: This document specifies the installation, update, and lifecycle management of the AUDiot hardware telemetry platform components.

---

# 7. Custom data ingestion requirement

## 7.5 Implementation: External Scripts (`vendor_exec`)
Custom hardware or software data can be ingested by placing executable scripts in the `/etc/hwexp/custom.d/` directory.
- **Cycle**: The exporter runs all scripts in this directory every poll cycle.
- **Contract**: Scripts MUST output JSON matching the `RawMeasurement` schema.
- **Discovery**: Scripts SHOULD handle the `--discover` flag to return device metadata.

---

# 8. Local LLM monitoring requirement

## 8.7 Implementation: Llamaswap Adapter
LLM monitoring is now a built-in capability of the core exporter via the `llamaswap` adapter.
- **Configuration**: Enabled via `adapters.llamaswap.enabled: true`.
- **Default Endpoint**: `http://localhost:50099`.
- **Data Source**: Polls the OpenAI-compatible `/v1/models` API.

---

# 18. Operational Lifecycle

The monitoring stack follows a **Research -> Strategy -> Execution** lifecycle for updates, mediated by a set of management scripts.

## 18.1 Host-Owned Layout
The primary deployment model is the **host-owned layout**.
- **Installation**: Handled by `install-layout.sh`.
- **Customization**: All user-specific config (hwexp.yaml, prometheus.yml, custom dashboards) is stored on the host and bind-mounted into containers.
- **Persistence**: Updates (via `manage-*.sh update`) preserve all user-edited files and directories.

## 18.2 Remote Deployment (`deploy-remote.sh`)
Deployments to remote hosts (e.g., Raspberry Pi dashboard host or multiple collector nodes) are orchestrated via `deploy-remote.sh`.
- **Mechanism**: SSH and rsync are used to push the monitoring subdirectory to the target.
- **Execution**: Triggers a remote `manage-*.sh update` to apply changes.

## 18.3 Management Interface
- **Collector**: `manage-collector.sh` (install, update, validate, verify-metrics, status, logs).
- **Dashboard**: `manage-dashboard.sh` (install, update, validate, list-dashboards, set-dashboard, restart-kiosk).
- **Kiosk**: `kiosk.sh` (auto-resolution detection, browser lifecycle management).

# 19. File and Directory Layout

Standard installation paths for Linux / Docker:

| Path | Purpose |
| --- | --- |
| `/etc/hwexp/hwexp.yaml` | Main configuration file. |
| `/etc/hwexp/conf.d/` | **Modular Config**: Additional `.yaml` files merged at startup. |
| `/etc/hwexp/custom.d/` | **Custom Plugins**: Executable scripts for custom data. |
| `/etc/hwexp/mappings.yaml` | Manual metric mapping rules. |
| `/var/lib/hwexp/mappings.auto.yaml` | Auto-generated mapping rules. |
| `/usr/bin/hwexp` | Exporter binary. |
| `/opt/docker/collector/` | **Collector Layout**: Prometheus rules, hwexp config, and compose file. |
| `/opt/docker/dashboard/` | **Dashboard Layout**: Grafana provisioning, profiles, and custom dashboards. |
