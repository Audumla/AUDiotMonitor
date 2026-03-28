# Hardware Telemetry Exporter Platform

## Mapping Rules Reference

**Document purpose**: This document defines the exhaustive syntax, semantics, precedence rules, and examples for mapping rules used to normalize raw measurements into stable metric families, labels, and logical names.

**Status**: Normative for implementation

**Audience**: Developers implementing the mapping engine, adapter authors, operators writing custom rules, and test authors building golden fixtures

---

# 1. Scope

This document covers:

* rule file structure
* match semantics
* normalize semantics
* rule precedence
* alias interaction
* conflict handling
* examples by device class
* validation expectations

This document does not define the machine-readable schema itself; that belongs to the JSON Schemas Pack.

---

# 2. Conceptual purpose of mapping

Mapping exists to translate unstable raw source measurements into:

* stable metric families
* stable labels
* stable logical names
* stable device classification

The mapping layer is the primary contract boundary between:

* vendor-/platform-specific raw data
* dashboard-/query-stable observability data

---

# 3. File structure

Canonical file shape:

```yaml
schema_version: "1.0.0"
rules:
  - id: example_rule
    priority: 100
    match:
      source: linux_hwmon
      device_class: gpu
      raw_name_regex: "temp1_input"
    normalize:
      metric_family: hw_device_temperature_celsius
      metric_type: gauge
      component: thermal
      sensor: core
      logical_name_template: "${logical_device_name}_core_temp"
      unit_scale: 0.001
      labels:
        component: thermal
        sensor: core
```

---

# 4. Rule anatomy

Each rule consists of:

* `id`
* `priority`
* `match`
* `normalize`

## 4.1 `id`

Requirements:

* MUST be unique within a rule file
* SHOULD be human-readable
* SHOULD remain stable over time for debugging and golden tests

Recommended pattern:

* `<vendor_or_scope>_<device>_<sensor>`

Examples:

* `amd_gpu_core_temp`
* `corsair_psu_output_power`
* `generic_hwmon_fan_speed`

## 4.2 `priority`

Requirements:

* integer
* higher value wins
* default if omitted MAY be implementation-defined, but explicit priority is preferred

Recommended ranges:

* 1000+: site/user overrides
* 500–999: device-specific built-ins
* 200–499: vendor/model class rules
* 100–199: class-level generic rules
* 2–99: fallback/default rules
* 1: auto-generated rules (written by the automapper on first run)

The mapping engine sorts rules by priority descending at load time and after
any dynamic rule addition. The first matching rule wins, so a manual rule with
priority ≥ 2 will always take precedence over an auto-generated rule (priority 1)
for the same sensor, even if the auto-generated rule was loaded first.

---

# 5. Match section

The `match` block defines when a rule applies.

All populated match fields are ANDed.

## 5.1 Supported keys

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

## 5.2 Exact-match fields

These fields are exact string matches in v1:

* `platform`
* `source`
* `device_class`
* `device_subclass`
* `vendor`
* `component_hint`
* `sensor_hint`

## 5.3 Regex fields

These fields are regex patterns in v1:

* `model_regex`
* `stable_id_regex`
* `raw_name_regex`

Regex rules:

* regex flavor MUST match implementation language/runtime policy
* invalid regex MUST fail validation
* regex matching SHOULD be deterministic and documented

## 5.4 Matching examples

### Example 1: AMD GPU hwmon temperature

```yaml
match:
  platform: linux
  source: linux_hwmon
  device_class: gpu
  vendor: amd
  raw_name_regex: "temp1_input"
```

### Example 2: Corsair PSU fan speed

```yaml
match:
  platform: linux
  source: linux_hwmon
  device_class: psu
  vendor: corsair
  model_regex: "HX1500i"
  raw_name_regex: "fan1_input"
```

### Example 3: Generic fallback fan speed

```yaml
match:
  source: linux_hwmon
  raw_name_regex: "fan[0-9]+_input"
```

---

# 6. Normalize section

The `normalize` block describes the normalized output.

## 6.1 Required normalize fields

* `metric_family`
* `metric_type`
* `logical_name_template`

## 6.2 Supported normalize keys

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

## 6.3 Field semantics

### `metric_family`

The final normalized metric family name.

Must follow canonical naming rules.

### `metric_type`

Allowed values:

* `gauge`
* `counter`
* `state_set`
* `info`

### `device_class`

Overrides or sets canonical device class if needed.

### `device_subclass`

Optional refined classification.

### `component`

Canonical component grouping.
Examples:

* `thermal`
* `fan`
* `memory`
* `output`
* `rail`

### `sensor`

Canonical sensor identifier within component.
Examples:

* `core`
* `hotspot`
* `rpm`
* `voltage_12v`

### `logical_name_template`

A deterministic template for logical-name generation.

### `unit_scale`

Multiplier applied to raw numeric value.
Example:

* `0.001` for millidegree Celsius to Celsius

### `unit_offset`

Offset applied after scaling.
Example:

* temperature or unusual vendor transforms if needed

### `value_transform`

Reserved for constrained expression support.
SHOULD be avoided in v1 unless tightly limited.

### `labels`

Additional fixed labels to merge into the normalized label set.

### `drop`

If `true`, the rule drops matching measurements.
Useful for noisy or irrelevant source values.

---

# 7. Template interpolation rules

## 7.1 Supported template variables

The implementation MUST support interpolation at least for:

* `${logical_device_name}`
* `${stable_device_id}`
* `${vendor}`
* `${model}`
* `${source}`

Regex capture substitution MAY also be supported.
If supported, capture syntax MUST be clearly documented and consistent.

Recommended capture form:

* `${1}`
* `${2}`

## 7.2 Logical name examples

### Example

```yaml
logical_name_template: "${logical_device_name}_core_temp"
```

Possible result:

* `gpu0_core_temp`

### Example with capture group

```yaml
raw_name_regex: "fan([0-9]+)_input"
logical_name_template: "${logical_device_name}_fan_${1}_rpm"
```

Possible result:

* `psu0_fan_1_rpm`

---

# 8. Precedence rules

When multiple rules match the same raw measurement, precedence MUST be applied in this order:

1. highest `priority`
2. most specific rule wins
3. user-defined rule wins over built-in rule
4. lexical `id` stable tie-break

## 8.1 Specificity definition

Specificity is determined by the number of populated match keys.

A rule matching on:

* `source + vendor + model_regex + raw_name_regex`

is more specific than one matching only:

* `source + raw_name_regex`

## 8.2 Conflict handling

When two matching rules remain tied after precedence evaluation:

* mapping decision MUST record `conflicted`
* implementation MAY choose deterministic lexical tie-break for forward progress
* strict mode SHOULD treat this as validation or runtime mapping error where statically detectable

---

# 9. Alias interaction

Aliases affect mapping by influencing:

* `logical_device_name`
* `display_name`

## 9.1 Alias precedence

If an alias exists for a discovered device:

* alias-derived `logical_device_name` MUST be used before fallback logical device naming

## 9.2 Alias non-effect areas

Aliases MUST NOT directly alter:

* `stable_id`
* `vendor`
* `model`
* raw source identity

---

# 10. Drop rules

Drop rules are legitimate mapping rules that suppress unwanted source data.

Example:

```yaml
- id: drop_hwmon_alarm_flags
  priority: 900
  match:
    source: linux_hwmon
    raw_name_regex: ".*_alarm$"
  normalize:
    metric_family: hw_device_status
    metric_type: gauge
    logical_name_template: "drop_unused"
    drop: true
```

Requirements:

* dropped items MUST still appear in mapping decisions with `decision=dropped`
* dropped items MUST NOT be emitted to `/metrics`

---

# 11. Generic fallback rules

Generic fallback rules SHOULD exist for broad usefulness, but MUST remain lower priority than device/vendor-specific rules.

Examples:

* generic temperature conversion from millidegree Celsius
* generic fan speed conversion
* generic voltage conversion for common hwmon sources

Fallback rules MUST avoid overclassifying unknown devices incorrectly.

---

# 12. Recommended rule families by device class

## 12.1 GPU

Common normalized outputs:

* core temp
* hotspot temp
* power
* utilization
* VRAM used
* VRAM total
* fan RPM

## 12.2 PSU

Common normalized outputs:

* output power
* input power
* internal temp
* fan RPM
* rail voltages
* rail currents

## 12.3 CPU

Common normalized outputs:

* package temp
* utilization
* power estimate if available

## 12.4 Storage

Common normalized outputs:

* temp
* wear
* used bytes
* throughput

## 12.5 LLM/service/custom data

Common normalized outputs:

* requests total
* requests in flight
* queue depth
* request duration
* tokens per second
* context tokens
* model loaded state

These may use service/llm-oriented metric families while still following the same mapping engine behavior.

---

# 13. Examples

## 13.1 AMD GPU core temp

```yaml
- id: amd_gpu_core_temp
  priority: 700
  match:
    platform: linux
    source: linux_hwmon
    device_class: gpu
    vendor: amd
    raw_name_regex: "temp1_input"
  normalize:
    metric_family: hw_device_temperature_celsius
    metric_type: gauge
    component: thermal
    sensor: core
    logical_name_template: "${logical_device_name}_core_temp"
    unit_scale: 0.001
    labels:
      component: thermal
      sensor: core
```

## 13.2 AMD GPU hotspot temp

```yaml
- id: amd_gpu_hotspot_temp
  priority: 700
  match:
    platform: linux
    source: linux_hwmon
    device_class: gpu
    vendor: amd
    raw_name_regex: "temp2_input"
  normalize:
    metric_family: hw_device_temperature_celsius
    metric_type: gauge
    component: thermal
    sensor: hotspot
    logical_name_template: "${logical_device_name}_hotspot_temp"
    unit_scale: 0.001
    labels:
      component: thermal
      sensor: hotspot
```

## 13.3 Corsair PSU output power

```yaml
- id: corsair_psu_output_power
  priority: 800
  match:
    platform: linux
    source: linux_hwmon
    device_class: psu
    vendor: corsair
    model_regex: "HX1500i"
    raw_name_regex: "power1_input"
  normalize:
    metric_family: hw_device_power_watts
    metric_type: gauge
    component: output
    sensor: output
    logical_name_template: "${logical_device_name}_output_power"
    unit_scale: 0.000001
    labels:
      component: output
      sensor: output
```

## 13.4 Generic fan speed

```yaml
- id: generic_hwmon_fan_speed
  priority: 120
  match:
    source: linux_hwmon
    raw_name_regex: "fan([0-9]+)_input"
  normalize:
    metric_family: hw_device_fan_speed_rpm
    metric_type: gauge
    component: fan
    sensor: rpm
    logical_name_template: "${logical_device_name}_fan_${1}_rpm"
    labels:
      component: fan
      sensor: rpm
```

## 13.5 LLM requests in flight

```yaml
- id: llm_requests_in_flight
  priority: 900
  match:
    source: llm_observer
    device_class: llm
    raw_name_regex: "requests_in_flight"
  normalize:
    metric_family: hw_service_requests_in_flight
    metric_type: gauge
    component: inference
    sensor: requests_in_flight
    logical_name_template: "${logical_device_name}_requests_in_flight"
    labels:
      component: inference
      sensor: requests_in_flight
```

---

# 14. Validation expectations

The mapping system MUST validate:

* unique rule IDs
* valid regex strings
* valid metric family names
* valid metric types
* valid logical name template syntax
* label key validity
* duplicate/conflicting alias usage where applicable

In strict mode, the implementation SHOULD also detect:

* obvious unreachable duplicate rules
* conflicting equal-precedence rules on the same exact match pattern

---

# 15. Operational guidance

## 15.1 Built-ins vs user overrides

User rules SHOULD be kept in separate files from built-ins.

User rules SHOULD override built-ins through higher priority or explicit file-layer precedence.

## 15.2 Safe evolution

When changing a rule that affects a stable metric family or logical name:

* treat it as a potentially breaking observability change
* update tests/golden outputs
* document migration impact

---

# 16. Testing obligations from this spec

Tests MUST cover:

* exact-match behavior
* regex-match behavior
* priority ordering
* specificity ordering
* alias impact on logical names
* drop-rule behavior
* capture-group interpolation if supported
* deterministic conflict resolution
* LLM/custom-data rule handling where implemented

---

# 17. Immediate next document

After this document, the next highest-value implementation document is the **Dashboard Data Contract Guide**, because it translates stable normalized metrics into predictable dashboard query and grouping patterns.



