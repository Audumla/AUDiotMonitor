# Hardware Telemetry Exporter Platform

## JSON Schemas Pack Specification

**Document purpose**: This document defines the machine-readable schema pack required by the exporter platform. It specifies which schemas MUST exist, their versioning rules, validation boundaries, file layout, and the normative constraints each schema MUST encode.

**Status**: Normative for implementation

**Audience**: Developers implementing config validation, API payload validation, test harnesses, installer validation, and schema generation/publishing workflows

---

# 1. Scope

This document covers the machine-readable JSON Schemas for:

* main exporter config
* mapping rules file
* aliases file
* bearer token file
* install inventory
* release manifest
* debug endpoint payloads
* common reusable schema fragments

This document does not define the human-readable prose rules in full; it encodes them into machine-validatable schema expectations.

---

# 2. Schema pack goals

The schema pack MUST:

* remove ambiguity from file formats and JSON endpoints
* support validation in CI and at runtime
* be versioned independently from binary version
* be distributable with release artifacts
* support both implementation-time and operations-time validation
* avoid overfitting to one language runtime

---

# 3. Canonical layout

The schema pack MUST be shipped under:

```text
schemas/
  common/
    scalar_types.schema.json
    labels.schema.json
    enums.schema.json
    errors.schema.json
    time.schema.json
  config.schema.json
  mappings.schema.json
  aliases.schema.json
  tokens.schema.json
  install_inventory.schema.json
  release_manifest.schema.json
  api/
    version.schema.json
    healthz.schema.json
    readyz.schema.json
    debug_discovery.schema.json
    debug_mappings.schema.json
    debug_raw.schema.json
    debug_catalog.schema.json
```

---

# 4. Schema dialect and metadata

All schemas MUST:

* use JSON Schema draft 2020-12
* declare `$schema`
* declare stable `$id`
* declare `title`
* declare `type`
* define `additionalProperties` explicitly at each object boundary

Recommended `$id` prefix:

```text
https://example.local/schemas/hwexp/
```

---

# 5. Common reusable schema fragments

## 5.1 `common/scalar_types.schema.json`

MUST define reusable scalar types including:

* `non_empty_string`
* `semantic_version`
* `rfc3339_utc_timestamp`
* `duration_string`
* `hostname_or_identifier`
* `prometheus_label_name`
* `logical_name`
* `listen_address`
* `file_path`

## 5.2 `common/labels.schema.json`

MUST define:

* label map structure
* regex for label keys
* value type rules
* reserved label exclusions where practical

## 5.3 `common/enums.schema.json`

MUST define canonical enums for:

* platform
* device class
* device subclass where enumerated
* quality
* metric type
* auth mode
* bind scope
* log level
* error class

## 5.4 `common/errors.schema.json`

MUST define:

* `ErrorObject`
* detail item shape
* error code string patterns

## 5.5 `common/time.schema.json`

MUST define:

* RFC3339 UTC timestamp fields
* generated_at timestamp field
* first_seen / last_seen timestamp fragment

---

# 6. Main config schema

## 6.1 File

```text
schemas/config.schema.json
```

## 6.2 Top-level object

Required properties:

* `server`
* `identity`
* `adapters`
* `mapping`

Optional properties:

* `filters`
* `security`
* `debug`
* `telemetry`

Top-level `additionalProperties` MUST be `false`.

## 6.3 Validation rules to encode

The schema MUST enforce:

* correct required keys
* correct primitive/object types
* enums for constrained fields
* minimums for integer counts
* regex for logical names and labels
* conditional requirement of `tls_cert_file` and `tls_key_file` when `tls_enabled=true`
* adapter entries shaped as adapter config objects

The schema SHOULD encode only structural constraints, not runtime environment validation like file existence.

## 6.4 Example-required definitions

The schema MUST include definitions for:

* `server`
* `identity`
* `adapter_config`
* `mapping_config`
* `filters_config`
* `security_config`
* `debug_config`
* `telemetry_config`

---

# 7. Mapping rules schema

## 7.1 File

```text
schemas/mappings.schema.json
```

## 7.2 Top-level shape

Required properties:

* `schema_version`
* `rules`

Top-level `additionalProperties=false`.

## 7.3 Rule object requirements

Each rule MUST require:

* `id`
* `match`
* `normalize`

`normalize` MUST require:

* `metric_family`
* `metric_type`
* `logical_name_template`

## 7.4 Match object

Allowed properties:

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

`additionalProperties=false`.

## 7.5 Normalize object

Allowed properties:

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
* `drop`

`additionalProperties=false`.

## 7.6 Schema limitations

The schema MAY validate regex strings as strings but MUST NOT attempt to evaluate them.

The schema SHOULD allow `value_transform` syntactically but implementation MAY reject it at runtime in v1.

---

# 8. Aliases schema

## 8.1 File

```text
schemas/aliases.schema.json
```

## 8.2 Top-level shape

Required:

* `schema_version`
* `aliases`

Each alias item MUST require:

* `match`
* `set`

`match` MUST require:

* `stable_id`

`set` MUST require at least one of:

* `logical_device_name`
* `display_name`

`logical_device_name` MUST match the canonical logical-name regex.

---

# 9. Tokens schema

## 9.1 File

```text
schemas/tokens.schema.json
```

## 9.2 Top-level shape

Required:

* `schema_version`
* `tokens`

Each token item MUST require:

* `id`
* `token`
* `scopes`

Allowed scopes in v1:

* `metrics:read`
* `debug:read`
* `health:read`

---

# 10. Install inventory schema

## 10.1 File

```text
schemas/install_inventory.schema.json
```

## 10.2 Purpose

Defines machine-readable validation of the lifecycle tool inventory.

## 10.3 Required top-level fields

* `schema_version`
* `tool_version`
* `host`
* `platform`
* `installed_components`
* `updated_at`

## 10.4 Installed component item

Each installed component MUST include:

* `name`
* `version`
* `release_tag`
* `installed_at`
* `managed_files`
* `managed_services`

Optional:

* `role_origin`
* `migrations_applied`
* `notes`

---

# 11. Release manifest schema

## 11.1 File

```text
schemas/release_manifest.schema.json
```

## 11.2 Required top-level fields

* `schema_version`
* `release_version`
* `published_at`
* `components`

## 11.3 Component entry

Each component entry MUST require:

* `name`
* `version`
* `artifacts`

Optional:

* `requires`
* `conflicts`
* `min_config_schema`
* `max_config_schema`
* `min_metric_schema`
* `max_metric_schema`
* `migrations`

## 11.4 Artifact entry

Each artifact MUST require:

* `os`
* `arch`
* `format`
* `file`
* `sha256`

Optional:

* `size`
* `url`
* `notes`

---

# 12. API response schemas

## 12.1 `api/version.schema.json`

MUST validate:

* `schema_version`
* `exporter_version`
* `build_commit`
* `build_time`
* `api_schema_version`
* `metric_schema_version`
* `config_schema_version`
* `platform`
* `go_version`

## 12.2 `api/healthz.schema.json`

MUST validate:

* `schema_version`
* `status`
* `timestamp`

Allowed `status`:

* `ok`
* `error`

## 12.3 `api/readyz.schema.json`

MUST validate:

* `schema_version`
* `status`
* `timestamp`
* `checks`

Allowed `status`:

* `ready`
* `not_ready`

## 12.4 `api/debug_discovery.schema.json`

MUST validate:

* top-level response shape
* nested `DiscoveredDevice` items
* summary block

## 12.5 `api/debug_mappings.schema.json`

MUST validate:

* top-level response shape
* nested `MappingDecision` items
* summary block

## 12.6 `api/debug_raw.schema.json`

MUST validate:

* top-level response shape
* nested `RawMeasurement` items

## 12.7 `api/debug_catalog.schema.json`

MUST validate:

* top-level response shape
* catalog item fields
* grouping hints array

---

# 13. Shared object definitions for API schemas

The API schemas SHOULD reuse common object definitions for:

* `DiscoveredDevice`
* `RawMeasurement`
* `NormalizedMeasurement` where needed
* `MappingDecision`
* `ErrorObject`

These MAY be implemented through `$defs` or shared referenced files.

---

# 14. `additionalProperties` policy

## 14.1 Default policy

The schema pack MUST prefer explicit contracts.

Therefore:

* top-level API/config objects: `additionalProperties=false`
* deeply nested structured objects: `additionalProperties=false`
* label maps and metadata bags: `additionalProperties=true` only where intentionally extensible

## 14.2 Extensible fields

These fields MAY allow additional keys:

* `raw_identifiers`
* `adapter_metadata`
* `metadata`
* `static_labels`
* adapter `settings`
* release manifest component `notes`-like extensible structures if added later

---

# 15. Versioning rules

## 15.1 Schema version tracking

Each schema file MUST declare its own version through:

* `$id`
* `title`
* semantic version note in `description` or schema metadata if desired

## 15.2 Breaking changes

A schema change is breaking if it:

* removes a required field
* changes field type
* narrows accepted enum values in a non-compatible way
* changes regex in a way that invalidates previously valid data without migration

Breaking changes MUST increment the relevant schema version.

---

# 16. Validation responsibilities

## 16.1 Runtime validation

The exporter/lifecycle tool MUST validate:

* config files
* mapping files
* aliases files
* token files

at startup and on reload where applicable.

## 16.2 CI validation

CI MUST validate:

* example configs against schemas
* fixture payloads against API schemas where applicable
* release manifests against release manifest schema
* install inventory fixture examples against install inventory schema

---

# 17. Packaging requirements

The schema pack MUST be:

* included in release artifacts
* installed with config/supporting packages where appropriate
* available to tests without requiring internet access

Recommended install locations:

* source tree: `schemas/`
* installed Linux location: `/usr/share/hwexp/schemas/`

---

# 18. Immediate implementation outputs

The first machine-readable schema set to produce MUST include:

1. `config.schema.json`
2. `mappings.schema.json`
3. `aliases.schema.json`
4. `tokens.schema.json`
5. `api/version.schema.json`
6. `api/healthz.schema.json`
7. `api/readyz.schema.json`
8. `api/debug_discovery.schema.json`
9. `api/debug_mappings.schema.json`
10. `api/debug_catalog.schema.json`
11. `install_inventory.schema.json`
12. `release_manifest.schema.json`

---

# 19. Follow-on implementation document

After this document, the next highest-value implementation document is the **Exporter Interface Specification**, because it defines the code-level package boundaries and lifecycle model that will produce the payloads validated by this schema pack.



