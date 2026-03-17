# Hardware Telemetry Exporter Platform

## Detailed Contracts, Configuration, Errors, and Security Specification

**Document purpose**: This document is the implementation-level companion to `Overview.md` and `installation_spec.md`. It removes remaining ambiguity by defining:

* Formal data contracts for internal and external APIs
* Canonical JSON payload schemas for debug endpoints
* Measurement and discovery object models
* Error taxonomy and HTTP/API error responses
* Logging, eventing, and recovery expectations
* Formal configuration schema and validation rules
* Authentication, authorization, and transport security model
* Compatibility and schema versioning requirements

**Status**: Normative for implementation

**Audience**: Developers implementing the exporter, bridge adapters, packaging, tests, and operational deployment

---

# 1. Normative language

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are to be interpreted as requirements.

Where this document conflicts with `Overview.md` or `installation_spec.md`, this document takes precedence for low-level implementation details.

---

# 2. Scope of this document

This document defines four main areas:

1. **Data contracts**

   * Internal object models
   * HTTP response payloads
   * Debug endpoint schemas
   * Exporter self-metrics metadata expectations

2. **Error handling**

   * Error classes and machine-readable codes
   * HTTP status behavior
   * Logging structure and levels
   * Retry, fallback, and recovery rules

3. **Configuration**

   * Formal schema model
   * Validation rules
   * Defaults
   * Merge/override behavior

4. **Security**

   * Endpoint exposure modes
   * Authentication and authorization model
   * TLS strategy
   * Secrets handling

This document does **not** define Grafana dashboard JSON, Prometheus rule files, or packaging scripts in full. Those belong in later deployment and dashboard documents.

---

# 3. Canonical schema identity

The exporter platform MUST track three distinct schema versions:

* `metric_schema_version`
* `api_schema_version`
* `config_schema_version`

Initial values:

* `metric_schema_version = 1.0.0`
* `api_schema_version = 1.0.0`
* `config_schema_version = 1.0.0`

The exporter binary MUST expose these via:

* `/version` endpoint
* startup logs
* optional build metadata command-line output

---

# 4. Core data contracts

## 4.1 Common conventions

All JSON payloads MUST:

* use UTF-8
* use `snake_case` field names
* use RFC 3339 timestamps in UTC with trailing `Z`
* include a top-level `schema_version` field
* avoid null unless semantically meaningful
* omit empty optional arrays/maps where practical

Numeric value rules:

* integers for counts and indexes
* floats for measurements and unit-converted values
* booleans for state flags
* strings for identifiers, names, status enums, and raw textual metadata

IDs MUST be stable within the lifetime rules documented below.

---

## 4.2 DiscoveredDevice contract

A `DiscoveredDevice` represents one physical or logical device known to the exporter.

### 4.2.1 JSON shape

```json
{
  "stable_id": "pci-0000:0b:00.0",
  "platform": "linux",
  "source": "linux_hwmon",
  "device_class": "gpu",
  "device_subclass": "discrete",
  "vendor": "amd",
  "model": "Radeon RX 7900 XTX",
  "driver": "amdgpu",
  "bus": "pci",
  "location": "0000:0b:00.0",
  "display_name": "Main 7900 XTX",
  "logical_device_name": "gpu0",
  "capabilities": ["temperature", "power", "fan_speed", "memory"],
  "raw_identifiers": {
    "pci_address": "0000:0b:00.0",
    "sysfs_path": "/sys/class/drm/card0/device"
  },
  "adapter_metadata": {
    "adapter": "linux_gpu_vendor",
    "priority": 100
  },
  "first_seen": "2026-03-17T12:00:00Z",
  "last_seen": "2026-03-17T12:05:00Z",
  "present": true
}
```

### 4.2.2 Required fields

* `stable_id`
* `platform`
* `source`
* `device_class`
* `vendor`
* `model`
* `capabilities`
* `present`

### 4.2.3 Field rules

#### `stable_id`

MUST be the canonical unique ID used throughout the exporter.

Allowed generation precedence:

1. explicit persistent hardware identifier if stable and safe
2. PCI address
3. deterministic USB identity
4. OS-native canonical path
5. exporter-generated persisted fallback

#### `platform`

Allowed values initially:

* `linux`
* `windows`
* `darwin`
* `unknown`

#### `source`

MUST identify the adapter or bridge that produced the discovery record.
Examples:

* `linux_hwmon`
* `linux_gpu_vendor`
* `linux_node_bridge`
* `windows_exporter_bridge`

#### `device_class`

MUST match the canonical taxonomy.

#### `present`

* `true` if discovered in the current discovery cycle
* `false` if retained as stale state within grace period

---

## 4.3 RawMeasurement contract

A `RawMeasurement` represents a sensor reading before normalization.

### 4.3.1 JSON shape

```json
{
  "measurement_id": "linux_hwmon:pci-0000:0b:00.0:temp1_input",
  "stable_device_id": "pci-0000:0b:00.0",
  "source": "linux_hwmon",
  "raw_name": "temp1_input",
  "raw_value": 54234,
  "raw_unit": "millidegree_celsius",
  "timestamp": "2026-03-17T12:05:03Z",
  "quality": "good",
  "component_hint": "thermal",
  "sensor_hint": "temperature",
  "metadata": {
    "sysfs_path": "/sys/class/hwmon/hwmon5/temp1_input",
    "label": "edge"
  }
}
```

### 4.3.2 Required fields

* `measurement_id`
* `stable_device_id`
* `source`
* `raw_name`
* `raw_value`
* `raw_unit`
* `timestamp`
* `quality`

### 4.3.3 `quality` enum

Allowed values:

* `good`
* `estimated`
* `degraded`
* `stale`
* `invalid`
* `unsupported`

---

## 4.4 NormalizedMeasurement contract

A `NormalizedMeasurement` represents a mapped measurement ready for exposure.

### 4.4.1 JSON shape

```json
{
  "stable_device_id": "pci-0000:0b:00.0",
  "logical_name": "gpu0_core_temp",
  "metric_family": "hw_device_temperature_celsius",
  "metric_type": "gauge",
  "value": 54.234,
  "unit": "celsius",
  "labels": {
    "host": "sanctum",
    "platform": "linux",
    "source": "linux_hwmon",
    "device_class": "gpu",
    "device_id": "pci-0000:0b:00.0",
    "logical_name": "gpu0_core_temp",
    "sensor": "core",
    "component": "thermal",
    "vendor": "amd",
    "model": "Radeon RX 7900 XTX"
  },
  "quality": "good",
  "mapping_rule_id": "amd_gpu_core_temp",
  "timestamp": "2026-03-17T12:05:03Z"
}
```

### 4.4.2 Required fields

* `stable_device_id`
* `logical_name`
* `metric_family`
* `metric_type`
* `value`
* `unit`
* `labels`
* `quality`
* `timestamp`

### 4.4.3 `metric_type` enum

Allowed values:

* `gauge`
* `counter`
* `state_set`
* `info`

`summary` and `histogram` are reserved for exporter self-metrics and MUST NOT be used for normalized hardware values in v1.

---

## 4.5 MappingDecision contract

A `MappingDecision` records how a raw measurement became a normalized measurement or why it was dropped.

### 4.5.1 JSON shape

```json
{
  "measurement_id": "linux_hwmon:pci-0000:0b:00.0:temp1_input",
  "decision": "mapped",
  "mapping_rule_id": "amd_gpu_core_temp",
  "precedence": 2,
  "raw_name": "temp1_input",
  "raw_unit": "millidegree_celsius",
  "converted_value": 54.234,
  "converted_unit": "celsius",
  "metric_family": "hw_device_temperature_celsius",
  "logical_name": "gpu0_core_temp",
  "labels": {
    "device_class": "gpu",
    "sensor": "core"
  },
  "notes": ["unit_scale applied: 0.001"]
}
```

### 4.5.2 `decision` enum

Allowed values:

* `mapped`
* `dropped`
* `ignored`
* `conflicted`
* `error`

---

## 4.6 ErrorObject contract

All machine-readable errors MUST conform to this shape.

```json
{
  "schema_version": "1.0.0",
  "error": {
    "code": "CFG_VALIDATION_FAILED",
    "message": "Configuration validation failed",
    "details": [
      {
        "path": "adapters.linux_hwmon.enabled",
        "issue": "expected boolean"
      }
    ],
    "retryable": false,
    "component": "config_loader",
    "timestamp": "2026-03-17T12:10:00Z",
    "request_id": "req-7f5a7c2d"
  }
}
```

Required fields:

* `code`
* `message`
* `retryable`
* `component`
* `timestamp`

---

# 5. HTTP endpoint contracts

## 5.1 Common endpoint behavior

All JSON endpoints MUST:

* set `Content-Type: application/json; charset=utf-8`
* include `Cache-Control: no-store` for debug endpoints
* include `X-Api-Schema-Version`
* include `X-Exporter-Version`
* include `X-Request-Id`

All endpoints SHOULD support gzip.

---

## 5.2 `/version`

### 5.2.1 Response

```json
{
  "schema_version": "1.0.0",
  "exporter_version": "0.1.0",
  "build_commit": "abc1234",
  "build_time": "2026-03-17T11:59:00Z",
  "api_schema_version": "1.0.0",
  "metric_schema_version": "1.0.0",
  "config_schema_version": "1.0.0",
  "platform": "linux",
  "go_version": "go1.25.0"
}
```

HTTP status:

* `200` always if process is alive

---

## 5.3 `/healthz`

### 5.3.1 Semantics

Reports liveness only.

### 5.3.2 Response

```json
{
  "schema_version": "1.0.0",
  "status": "ok",
  "timestamp": "2026-03-17T12:10:00Z"
}
```

Allowed `status`:

* `ok`
* `error`

HTTP status:

* `200` if process is alive
* `500` only if internal fatal state prevents basic service

---

## 5.4 `/readyz`

### 5.4.1 Semantics

Reports whether exporter is ready for scraping.

### 5.4.2 Response

```json
{
  "schema_version": "1.0.0",
  "status": "ready",
  "timestamp": "2026-03-17T12:10:00Z",
  "checks": {
    "config_loaded": true,
    "initial_refresh_complete": true,
    "listener_active": true,
    "required_adapters_initialized": true
  }
}
```

Allowed `status`:

* `ready`
* `not_ready`

HTTP status:

* `200` when ready
* `503` when not ready

---

## 5.5 `/debug/discovery`

### 5.5.1 Response shape

```json
{
  "schema_version": "1.0.0",
  "generated_at": "2026-03-17T12:10:00Z",
  "host": "sanctum",
  "devices": [
    {
      "stable_id": "pci-0000:0b:00.0",
      "platform": "linux",
      "source": "linux_gpu_vendor",
      "device_class": "gpu",
      "device_subclass": "discrete",
      "vendor": "amd",
      "model": "Radeon RX 7900 XTX",
      "driver": "amdgpu",
      "bus": "pci",
      "location": "0000:0b:00.0",
      "display_name": "Main 7900 XTX",
      "logical_device_name": "gpu0",
      "capabilities": ["temperature", "power", "fan_speed", "memory"],
      "raw_identifiers": {
        "pci_address": "0000:0b:00.0"
      },
      "adapter_metadata": {
        "adapter": "linux_gpu_vendor",
        "priority": 100
      },
      "first_seen": "2026-03-17T12:00:00Z",
      "last_seen": "2026-03-17T12:10:00Z",
      "present": true
    }
  ],
  "summary": {
    "device_count": 1,
    "unknown_device_count": 0
  }
}
```

HTTP status:

* `200` success
* `500` internal error
* `401/403` if auth enabled and unauthorized

---

## 5.6 `/debug/mappings`

### 5.6.1 Response shape

```json
{
  "schema_version": "1.0.0",
  "generated_at": "2026-03-17T12:10:00Z",
  "items": [
    {
      "measurement_id": "linux_hwmon:pci-0000:0b:00.0:temp1_input",
      "decision": "mapped",
      "mapping_rule_id": "amd_gpu_core_temp",
      "precedence": 2,
      "raw_name": "temp1_input",
      "raw_unit": "millidegree_celsius",
      "converted_value": 54.234,
      "converted_unit": "celsius",
      "metric_family": "hw_device_temperature_celsius",
      "logical_name": "gpu0_core_temp",
      "labels": {
        "device_class": "gpu",
        "sensor": "core"
      },
      "notes": ["unit_scale applied: 0.001"]
    }
  ],
  "summary": {
    "mapped": 1,
    "dropped": 0,
    "errors": 0
  }
}
```

Pagination MAY be added later. v1 MAY return all current cycle items.

---

## 5.7 `/debug/raw`

This endpoint is optional and MUST be disabled by default.

If enabled, it returns recent raw measurements for debugging.

### 5.7.1 Response shape

```json
{
  "schema_version": "1.0.0",
  "generated_at": "2026-03-17T12:10:00Z",
  "items": [
    {
      "measurement_id": "linux_hwmon:pci-0000:0b:00.0:temp1_input",
      "stable_device_id": "pci-0000:0b:00.0",
      "source": "linux_hwmon",
      "raw_name": "temp1_input",
      "raw_value": 54234,
      "raw_unit": "millidegree_celsius",
      "timestamp": "2026-03-17T12:10:00Z",
      "quality": "good",
      "metadata": {
        "sysfs_path": "/sys/class/hwmon/hwmon5/temp1_input"
      }
    }
  ]
}
```

---

## 5.8 `/debug/catalog`

### 5.8.1 Response shape

```json
{
  "schema_version": "1.0.0",
  "generated_at": "2026-03-17T12:10:00Z",
  "host": "sanctum",
  "catalog": [
    {
      "logical_name": "gpu0_core_temp",
      "display_name": "GPU 0 Core Temp",
      "stable_device_id": "pci-0000:0b:00.0",
      "device_class": "gpu",
      "device_subclass": "discrete",
      "component": "thermal",
      "sensor": "core",
      "metric_family": "hw_device_temperature_celsius",
      "unit": "celsius",
      "quality": "good",
      "source": "linux_hwmon",
      "group_hints": ["gpus", "temperatures"],
      "dashboard_priority": 100
    }
  ]
}
```

This endpoint is intended for dashboard discovery and tooling.

---

# 6. Internal state model

The exporter MUST internally track the following state stores:

* `device_registry`
* `raw_measurement_store`
* `normalized_measurement_store`
* `mapping_decision_store`
* `adapter_health_store`
* `config_state`

All stores MUST be atomically replaceable per refresh cycle.

At no time may `/metrics` expose a partially updated mix from two incompatible refresh cycles.

---

# 7. Error handling specification

## 7.1 Error classes

The exporter MUST classify all errors into one of these classes:

* `configuration`
* `adapter_initialization`
* `discovery`
* `polling`
* `mapping`
* `serialization`
* `http_request`
* `authorization`
* `filesystem`
* `dependency`
* `internal`

## 7.2 Canonical error codes

The following codes are reserved for v1.

### Configuration

* `CFG_FILE_NOT_FOUND`
* `CFG_PARSE_FAILED`
* `CFG_VALIDATION_FAILED`
* `CFG_UNSUPPORTED_VERSION`
* `CFG_RELOAD_FAILED`

### Adapter lifecycle

* `ADAPTER_INIT_FAILED`
* `ADAPTER_DISCOVERY_FAILED`
* `ADAPTER_POLL_FAILED`
* `ADAPTER_TIMEOUT`
* `ADAPTER_PERMISSION_DENIED`
* `ADAPTER_UNSUPPORTED_PLATFORM`

### Mapping

* `MAP_RULE_NOT_FOUND`
* `MAP_UNIT_CONVERSION_FAILED`
* `MAP_LABEL_VALIDATION_FAILED`
* `MAP_CONFLICT`
* `MAP_DROPPED_BY_FILTER`

### HTTP/API

* `HTTP_ROUTE_NOT_FOUND`
* `HTTP_METHOD_NOT_ALLOWED`
* `HTTP_UNAUTHORIZED`
* `HTTP_FORBIDDEN`
* `HTTP_RATE_LIMITED`
* `HTTP_INTERNAL_ERROR`

### Runtime

* `RUNTIME_NOT_READY`
* `RUNTIME_SNAPSHOT_EMPTY`
* `RUNTIME_DEPENDENCY_UNAVAILABLE`
* `RUNTIME_INTERNAL_PANIC`

## 7.3 Error response mapping

HTTP endpoints returning JSON errors MUST map statuses as follows:

* `400` invalid request or invalid query parameter
* `401` unauthenticated
* `403` authenticated but not authorized
* `404` route not found
* `405` method not allowed
* `409` conflicting operation state
* `429` rate limited if enabled later
* `500` internal failure
* `503` not ready / dependency unavailable

## 7.4 Logging requirements

All logs MUST be structured JSON in non-interactive mode.

Required fields on every log line:

* `timestamp`
* `level`
* `component`
* `event`
* `message`
* `host`
* `request_id` when request-scoped
* `error_code` when applicable

### Example

```json
{
  "timestamp": "2026-03-17T12:10:00Z",
  "level": "error",
  "component": "linux_hwmon",
  "event": "adapter_poll_failed",
  "message": "failed to read temp1_input",
  "host": "sanctum",
  "error_code": "ADAPTER_POLL_FAILED",
  "details": {
    "path": "/sys/class/hwmon/hwmon5/temp1_input",
    "cause": "permission denied"
  }
}
```

## 7.5 Log levels

Allowed levels:

* `debug`
* `info`
* `warn`
* `error`
* `fatal`

Rules:

* `fatal` MUST only be used immediately before process exit
* repeated noisy hardware failures SHOULD be rate-limited in logs
* recurring expected unsupported-device conditions SHOULD log once per discovery cycle at `debug` or `warn`

## 7.6 Recovery behavior

### Config load failure at startup

* process MUST exit non-zero
* MUST emit `CFG_*` code

### Config reload failure at runtime

* process MUST keep last known good config
* MUST emit `CFG_RELOAD_FAILED`
* MUST expose failed reload status in self-metrics

### Adapter poll failure

* MUST preserve last good snapshot for that adapter for grace period
* MUST mark affected measurements `stale` once grace threshold exceeded
* MUST continue serving other adapter data

### Adapter discovery failure

* MUST preserve previous device registry entries until expiry
* MUST mark `present=false` if device absent beyond grace policy

### Panic/internal fatal condition

* MUST log `RUNTIME_INTERNAL_PANIC`
* MUST terminate process unless panic is explicitly recovered in a safe boundary

---

# 8. Formal configuration schema

This section defines the canonical configuration model. A machine-readable JSON Schema MUST be generated from this section and shipped with the project.

## 8.1 Top-level structure

```yaml
server:
identity:
adapters:
mapping:
filters:
security:
debug:
telemetry:
```

## 8.2 Top-level fields

### `server`

Required object.

Fields:

* `listen_address` (string, required)
* `refresh_interval` (duration string, optional, default `5s`)
* `discovery_interval` (duration string, optional, default `60s`)
* `grace_period` (duration string, optional, default `30s`)
* `request_timeout` (duration string, optional, default `5s`)
* `max_concurrent_adapter_polls` (integer, optional, default `4`)

Validation:

* `listen_address` MUST be valid `host:port`
* durations MUST be positive
* `max_concurrent_adapter_polls` MUST be >= 1

### `identity`

Required object.

Fields:

* `host` (string, required)
* `platform` (enum: `linux|windows|darwin|auto`, optional, default `auto`)
* `site` (string, optional)
* `role` (string, optional)
* `environment` (string, optional)
* `static_labels` (map[string]string, optional)

Validation:

* `host` MUST be non-empty
* static label keys MUST match Prometheus label name pattern

### `adapters`

Required object.

Each adapter entry is an object with:

* `enabled` (bool, required)
* `priority` (int, optional, default adapter-specific)
* `timeout` (duration, optional)
* `settings` (map, optional)

Known keys in v1:

* `linux_hwmon`
* `linux_gpu_vendor`
* `linux_vendor_exec`
* `linux_node_bridge`
* `windows_exporter_bridge`
* `darwin_node_bridge`

Unknown adapter keys MAY be rejected in strict mode.

### `mapping`

Required object.

Fields:

* `rules_file` (path string, required)
* `aliases_file` (path string, optional)
* `strict_mode` (bool, optional, default `false`)
* `default_drop_unmapped` (bool, optional, default `false`)

### `filters`

Optional object.

Fields:

* `include_device_classes` (array[string], optional)
* `exclude_device_classes` (array[string], optional)
* `include_sources` (array[string], optional)
* `exclude_raw_name_regex` (array[string], optional)
* `suppress_qualities` (array[string], optional)

Validation:

* include/exclude arrays MUST NOT conflict on same exact value under strict validation

### `security`

Optional object.

Fields:

* `auth_mode` (enum, default `none`)
* `bind_scope` (enum `localhost|lan|custom`, default `lan`)
* `tls_enabled` (bool, default `false`)
* `tls_cert_file` (path, required if tls_enabled)
* `tls_key_file` (path, required if tls_enabled)
* `api_tokens_file` (path, optional)
* `debug_endpoints_enabled` (bool, default `true`)
* `debug_endpoints_auth_required` (bool, default `false`)

### `debug`

Optional object.

Fields:

* `enable_raw_endpoint` (bool, default `false`)
* `log_level` (enum, default `info`)
* `retain_last_mapping_cycles` (int, default `5`)
* `retain_last_raw_cycles` (int, default `1`)

### `telemetry`

Optional object.

Fields:

* `self_metrics_enabled` (bool, default `true`)
* `emit_build_info` (bool, default `true`)

---

## 8.3 Mapping rules schema

The mapping rules file MUST conform to this conceptual structure:

```yaml
schema_version: "1.0.0"
rules:
  - id: amd_gpu_core_temp
    priority: 100
    match:
      platform: linux
      source: linux_hwmon
      device_class: gpu
      vendor: amd
      raw_name_regex: "temp1_input"
    normalize:
      metric_family: hw_device_temperature_celsius
      metric_type: gauge
      device_class: gpu
      component: thermal
      sensor: core
      logical_name_template: "${logical_device_name}_core_temp"
      unit_scale: 0.001
      labels:
        sensor: core
        component: thermal
```

### Rule field requirements

Required per rule:

* `id`
* `match`
* `normalize.metric_family`
* `normalize.metric_type`
* `normalize.logical_name_template`

### Matching semantics

All populated match fields are ANDed.

Supported match keys:

* `platform`
* `source`
* `device_class`
* `device_subclass`
* `vendor`
* `model_regex`
* `stable_id_regex`
* `raw_name_regex`
* `component_hint`
* `sensor_hint`

### Rule precedence

Higher numeric `priority` wins.
If tied:

1. more specific rule wins (more populated match keys)
2. user file wins over built-in rule
3. lexical `id` tie-break only as stable fallback

### Normalize fields

Supported normalize keys:

* `metric_family`
* `metric_type`
* `device_class`
* `device_subclass`
* `component`
* `sensor`
* `logical_name_template`
* `unit_scale`
* `unit_offset`
* `value_transform`
* `labels`
* `drop` (bool)

`value_transform` is reserved for a limited safe expression model and SHOULD NOT be implemented in v1 unless tightly constrained.

---

## 8.4 Aliases file schema

```yaml
schema_version: "1.0.0"
aliases:
  - match:
      stable_id: "pci-0000:0b:00.0"
    set:
      logical_device_name: "gpu0"
      display_name: "Main 7900 XTX"
  - match:
      stable_id: "usb-corsair-hx1500i-01"
    set:
      logical_device_name: "psu0"
      display_name: "Corsair HX1500i"
```

Rules:

* exact `stable_id` match only in v1
* duplicate alias matches MUST fail validation
* `logical_device_name` MUST match `[a-zA-Z_][a-zA-Z0-9_]*`

---

# 9. Security specification

## 9.1 Threat model assumptions

This is a local-network observability component, not a public internet service.

Primary risks:

* unauthorized LAN access to device inventory
* exposure of debug endpoints
* accidental exposure to wider networks
* misuse of raw endpoint data

## 9.2 Default security posture

Defaults MUST be:

* `auth_mode = none`
* `tls_enabled = false`
* `debug_endpoints_enabled = true`
* `debug_endpoints_auth_required = false`
* bind suitable for LAN-only use

This default is acceptable only for trusted LAN environments.

## 9.3 Supported auth modes

Allowed values for `auth_mode` in v1:

* `none`
* `bearer_token`

Planned later:

* `mtls`
* `reverse_proxy`

## 9.4 Bearer token behavior

If `auth_mode = bearer_token`:

* `/metrics` MAY remain open or protected based on config
* all `/debug/*` endpoints MUST require valid bearer token
* `/healthz` MAY remain unauthenticated
* `/readyz` SHOULD be configurable

Token source:

* `api_tokens_file`

Format:

```yaml
schema_version: "1.0.0"
tokens:
  - id: local_admin
    token: "REDACTED_LONG_RANDOM_VALUE"
    scopes: ["metrics:read", "debug:read"]
```

## 9.5 Authorization scopes

Defined scopes:

* `metrics:read`
* `debug:read`
* `health:read`

If auth is enabled:

* `/metrics` requires `metrics:read`
* `/debug/*` requires `debug:read`
* `/healthz` and `/version` MAY require `health:read` depending on config

## 9.6 TLS

If `tls_enabled=true`:

* both cert and key files MUST exist
* startup MUST fail if either file is invalid
* plain HTTP redirect is OPTIONAL and disabled by default

For LAN use behind a reverse proxy, terminating TLS upstream is acceptable.

## 9.7 Secret handling

The exporter MUST NOT:

* log bearer tokens
* expose token file paths in debug payloads unnecessarily
* include sensitive file contents in error responses

## 9.8 Debug endpoint hardening

If `enable_raw_endpoint=true`, the exporter SHOULD log a warning at startup.

If `debug_endpoints_auth_required=false` while auth is enabled, the exporter SHOULD log a warning.

---

# 10. Adapter-specific error and timeout behavior

## 10.1 Adapter initialization

Each adapter MUST initialize independently.

If one adapter fails to initialize:

* log error
* mark adapter unavailable
* continue startup unless adapter is marked required

### Required adapter setting

Adapters MAY later support `required: true`.
If a required adapter fails initialization, readiness MUST remain `not_ready`.

## 10.2 Poll timeout

Each poll operation MUST obey a timeout.

On timeout:

* emit `ADAPTER_TIMEOUT`
* preserve last good values until grace period expires
* set adapter health degraded

## 10.3 Permission-denied behavior

On permission failures:

* log `ADAPTER_PERMISSION_DENIED`
* mark affected adapter degraded
* do not retry continuously in a tight loop
* retry on normal refresh cadence

---

# 11. State retention and staleness rules

## 11.1 Device retention

A previously discovered device MAY remain in registry for a configurable staleness window after disappearance.

Fields:

* `present=false`
* `last_seen` preserved

Default stale device retention:

* `10m`

## 11.2 Measurement retention

Measurements are current until next cycle.
If adapter fails:

* last values retained for `grace_period`
* quality downgraded to `stale` once grace exceeded

Stale measurements MAY continue to appear in `/debug/catalog`, but SHOULD NOT be emitted forever.

## 11.3 Quality downgrade rules

* live successful read -> `good`
* synthetic/derived -> `estimated`
* source partially failing -> `degraded`
* beyond grace -> `stale`
* parse or read failure -> omit value or mark `invalid` in debug paths

---

# 12. Formal validation requirements

## 12.1 Startup validation order

Startup MUST validate in this order:

1. main config parse
2. main config schema validation
3. aliases parse and validation
4. mapping rules parse and validation
5. adapter config validation
6. security config validation
7. listener bind attempt
8. adapter initialization

## 12.2 Validation failure behavior

On any fatal validation error before service start:

* exit non-zero
* print concise human-readable error to stderr
* emit structured log if logging initialized

## 12.3 Strict mode

If `mapping.strict_mode=true`:

* unmapped measurements SHOULD be surfaced as mapping errors
* conflicting rules MUST fail startup if statically detectable
* unknown device classes in user rules MUST fail validation

---

# 13. Minimal JSON schema requirements

A machine-readable JSON Schema MUST be provided for:

* main config
* mapping rules file
* aliases file
* token file

These schemas MUST be versioned and shipped under:

```text
schemas/
  config.schema.json
  mappings.schema.json
  aliases.schema.json
  tokens.schema.json
```

The schemas MUST include:

* field types
* required keys
* enum constraints
* regex patterns for labels and logical names
* `additionalProperties` policy

Recommended policy:

* top-level objects: `additionalProperties=false`
* adapter `settings`: `additionalProperties=true`

---

# 14. Request tracing and correlation

The exporter MUST support per-request correlation IDs.

Behavior:

* incoming `X-Request-Id` SHOULD be honored if safe
* otherwise generate one
* include request ID in logs and error responses for JSON endpoints

---

# 15. Compatibility requirements

## 15.1 Backward compatibility

A change is considered breaking if it modifies:

* any normalized metric family name
* any required label name
* any JSON response required field
* any config required field or semantic meaning

Breaking changes MUST:

* increment relevant schema version
* document migration path
* include test updates and golden updates

## 15.2 Reserved fields

The following field names are reserved for future use and MUST NOT be repurposed:

* `annotations`
* `traits`
* `relations`
* `deprecated`

---

# 16. Test requirements derived from this document

The implementation MUST include tests for:

* JSON payload serialization against sample schemas
* config validation success/failure cases
* mapping precedence conflicts
* error code and HTTP status mapping
* auth required vs unauthenticated vs unauthorized paths
* staleness downgrade behavior
* adapter timeout handling
* strict mode validation failures

At minimum, contract tests MUST verify:

* `/version`
* `/healthz`
* `/readyz`
* `/debug/discovery`
* `/debug/mappings`
* `/debug/catalog`

---

# 17. Immediate follow-on documents

After this document, the next required documents are:

1.  **JSON Schemas Pack**
    *   formal machine-readable schemas for config and endpoint payloads

2.  **Exporter Interface Specification**
    *   Go interfaces, package contracts, lifecycle, concurrency model

3.  **Mapping Rules Reference**
    *   exhaustive rule syntax and examples

4.  **installation_spec.md**
    *   Covers deployment, operations, packaging, and portability details.

5.  **Dashboard Data Contract Guide**
    *   canonical queries, grouping rules, panel data expectations

---

# 18. Implementation readiness statement

This document, together with `Overview.md`, is intended to be sufficient to begin:

* core exporter implementation
* endpoint implementation
* config validation layer
* security/auth integration
* test harness scaffolding
* schema generation work

The next highest-value implementation artifact is the **JSON Schemas Pack**, since it converts these contracts into machine-validatable definitions and removes the remaining ambiguity from config and debug payloads.
