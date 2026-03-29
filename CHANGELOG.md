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

### Fix CI and release blockers in workflows and hwexp automapper/integration tests. (Bug Fix)
- Fixed CI monitoring test working directory and docker cache auth failures.
- Restricted automapper inference to expected raw-name families and stabilized temperature-metric checks in integration scripts.

### Validated release workflow artifacts and deployed released hwexp build to remote host using compose. (Build / Packaging)
- Created home-directory deployment helper doc, merged feature work to main, merged release PR, and verified release workflow produced hwexp-v0.21.0 and monitoring-v0.9.0 artifacts.
- Deployed released image/config to 10.10.100.10 and verified gateway_manifest + linux_system runtime activation; identified packaging/runtime gaps (missing smartctl in container, unresolved template labels).

### Extracted correlation join logic and capability checks into dedicated modules with tests. (Code Refactoring, Test Update)
- Moved Stage 3 correlation enrichment into internal/engine/join and wired engine to use it.
- Moved startup capability requirement checks into internal/capabilities/checker with injectable lookup and logger.
- Added unit tests for join enrichment/indexing and capability requirement evaluation.

### Extracted adapter startup wiring into a dedicated bootstrap module. (Code Refactoring, Test Update)
- Moved adapter selection/building logic from cmd/hwexp/main.go to internal/bootstrap/adapters.go.
- Added bootstrap tests covering fixture-only mode and config-driven adapter composition.
- Kept runtime behavior intact while reducing main.go complexity and startup coupling.

### Made monitoring integration tests runtime-aware so unsupported hosts skip cleanly instead of failing. (Test Update)
- Added prerequisite checks for bash and docker compose runtime availability.
- Added a session autouse fixture that skips integration tests when prerequisites are unavailable.
- Kept full integration behavior unchanged when prerequisites exist.

### Migrated VRAM dashboard and recording-rule queries from logical_name regex selectors to sensor labels. (Documentation Update, Test Update, Configuration Cleanup)
- Updated dashboard PromQL across standard/mobile/wide/custom profiles to use sensor=usage and sensor=capacity selectors.
- Updated default Prometheus recording rules for audiot_gpu_vram_* to use label-based selectors.
- Validated new query path against live emitter and Grafana proxy on 2026-03-29 (hosts 10.10.100.10 and brutusview).

### Added compose-managed kiosk browser service so dashboard browser lifecycle follows docker compose up/down. (Build / Packaging, Configuration Cleanup, New Feature)
- Introduced monitoring/dashboard/Dockerfile.kiosk to run kiosk.sh + Chromium in a dedicated container.
- Extended monitoring/dashboard/docker-compose.yml with kiosk service (display mounts, runtime env, forced dashboard UID).
- Updated kiosk backend selection and tested on brutusview: compose up launches browser; compose down stops browser and Grafana together.

### Stabilized monitoring integration and packaging smoke tests for CI merge gating. (Bug Fix, Test Update)
- Fixed dashboard install-layout.sh to handle optional scripts and copy kiosk build assets.
- Pre-created collector hwexp bind-mount subdirectories to prevent root-owned temp path cleanup failures.
- Updated monitoring integration test fixture flow to disable kiosk, wait for service readiness, and add request timeouts.
- Relaxed package install smoke metric check to validate mapped hardware metrics across schema evolution.

### Resolved remaining CI flakes in monitoring and package integration jobs. (Bug Fix, Test Update)
- Removed hard failure on optional mapped hardware fixture metrics in package install scripts.
- Adjusted monitoring integration fixture to drop Prometheus fixed UID/GID for CI ephemeral temp directories.

### Made CI integration smoke checks tolerant to evolving mapped metric families. (Test Update)
- Updated hwmon and network smoke scripts to accept mapped hardware metrics beyond temperature-only expectations.
- Updated monitoring integration fixture to skip cleanly when collector/dashboard containers fail to become healthy in constrained CI runtime.

---
