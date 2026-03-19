# Changelog

## Unreleased

### Implemented missing API endpoints and self-metrics to align with specifications. (New Feature, Code Refactoring)
- Added /readyz and /debug/catalog endpoints.
- Implemented tracking and exposure of self-metrics (refresh duration, success, last success time, discovered devices, mapping failures).
- Updated StateStore to track if at least one poll cycle has completed (readiness).
- Updated HTTP server to expose these metrics in /metrics.

### Added comprehensive testing suite including mocks and cross-stack integration tests. (Test Update)
- Created unit tests for hwmon adapter using simulated sysfs.
- Created unit tests for Engine using mock adapters.
- Created unit tests for HTTP API using httptest and StateStore mocks.
- Developed cross-stack integration test in Python to verify end-to-end data flow from hwexp to Prometheus to Grafana.
- Verified and highlighted firewall requirements in INSTALL.md.

### Added GPU utilization tracking and expanded dashboard with OS/GPU telemetry. (New Feature, UI Improvement)
- Implemented linux_gpu adapter to collect utilization metrics from AMD (sysfs) and NVIDIA (nvidia-smi) GPUs.
- Enhanced automapper to handle percentage-based utilization metrics.
- Updated System Overview dashboard to include GPU Performance section (utilization and memory usage).
- Verified that OS metrics (CPU/RAM) are provided by node-exporter and already integrated into dashboards.
- Enabled linux_gpu_vendor adapter in default and collector configurations.

### Added hwinfo-level hardware specifications and inventory tracking. (New Feature, UI Improvement)
- Implemented linux_static adapter to collect Motherboard model, BIOS version, CPU core/thread counts, and total system RAM.
- Updated linux_gpu adapter to report VRAM capacity for AMD and NVIDIA GPUs.
- Introduced hw_device_info metric to expose hardware metadata as Prometheus labels.
- Added Hardware Inventory table and spec stats (Cores, RAM) to the System Overview dashboard.
- Enhanced automapper to support 'count' and 'bytes' units for inventory metrics.

### Added Llamaswap support to monitor local LLM models. (New Feature, UI Improvement)
- Implemented llamaswap adapter to discover and poll models from a local Llamaswap/OpenAI-compatible API.
- Updated dashboard to include AI & LLM Services section with model inventory.
- Configured default Llamaswap endpoint to http://localhost:50099.

### Added support for external/modular configuration and custom collection scripts. (Configuration Cleanup, New Feature)
- Implemented conf.d support to allow merging multiple YAML configuration files from /etc/hwexp/conf.d/.
- Implemented vendor_exec adapter to run custom scripts/binaries from /etc/hwexp/custom.d/ for highly customized telemetry.
- Updated Dockerfile to create and expose these directories as volumes.
- Added documentation for these customization features in INSTALL.md.

### Updated all project documentation and specifications to match current implementation. (Documentation Update)
- Refreshed README.md and INSTALL.md with new features (GPU, Inventory, AI, Plugins).
- Updated specifications/overview.md to align with final taxonomy and naming standards.
- Updated exporter_interface_spec.md to reflect implemented API endpoints and self-metrics.
- Updated installation_spec.md with new directory layouts for external customization (conf.d, custom.d).

### Created specialized dashboard defaults for common screen sizes, including 1920x440. (UI Improvement)
- Developed AUDiot Panel [1920x440] optimized for wide sensor panels.
- Developed AUDiot Panel [Portrait] optimized for vertical side-screens or mobile.
- Developed AUDiot Dashboard [1080p] for standard desktop monitoring.
- Updated documentation in INSTALL.md to guide users on selecting the right dashboard for their display.

### Implemented dashboard profiles with first-run scaffolding and persistent user edits. (UI Improvement, Configuration Cleanup)
- Organized dashboards into logical profile folders (Standard, Wide-Screens, Mobile, Debug).
- Updated docker-compose.yml with a grafana-init bootstrapper to populate mapped volumes with default profiles if empty.
- Configured Grafana to automatically create UI folders based on the filesystem structure.
- Ensured user edits to dashboards are persisted across container restarts.

### Synchronized all documentation and specifications with new dashboard profile and scaffolding architecture. (Documentation Update)
- Updated README.md with dashboard profile and scaffolding features.
- Updated specifications/overview.md and specifications/dashboard_implementation_spec.md to reflect logical profile organization and grafana-init behavior.
- Ensured INSTALL.md provides clear instructions for persisting user edits via local folder mapping.

---
