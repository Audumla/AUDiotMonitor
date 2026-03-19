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

---
