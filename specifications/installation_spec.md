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

# 17. File and Directory Layout

Standard installation paths for Linux / Docker:

| Path | Purpose |
| --- | --- |
| `/etc/hwexp/hwexp.yaml` | Main configuration file. |
| `/etc/hwexp/conf.d/` | **Modular Config**: Additional `.yaml` files merged at startup. |
| `/etc/hwexp/custom.d/` | **Custom Plugins**: Executable scripts for custom data. |
| `/etc/hwexp/mappings.yaml` | Manual metric mapping rules. |
| `/var/lib/hwexp/mappings.auto.yaml` | Auto-generated mapping rules. |
| `/usr/bin/hwexp` | Exporter binary. |
