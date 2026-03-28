# Hardware Telemetry Exporter Platform

## Dashboard Data Contract Guide

**Document purpose**: This document defines the canonical dashboard-facing data contracts, query patterns, grouping rules, and panel expectations for Grafana and other consumers of the exporter data model.

**Status**: Normative for dashboard/query design

**Audience**: Developers building dashboards, provisioning Grafana, writing queries, and designing panel layouts for Pi panels, desktop views, and discovery tools

---

# 1. Scope

This document defines:

* dashboard-facing metric usage rules
* canonical grouping strategies
* dashboard classes
* expected query semantics
* normalized label usage for filtering and templating
* panel-to-metric family mappings

This document does not define full Grafana JSON exports.

---

# 2. Dashboard design principles

Dashboards MUST be built against the normalized data model, not raw vendor metrics.

Dashboards MUST:

* rely on stable metric families
* rely on stable labels
* tolerate unknown devices appearing in discovery views
* separate kiosk/panel layouts from operator/debug layouts
* avoid assumptions tied only to one machine unless explicitly scoped

---

# 3. Dashboard classes

The platform MUST support at least three dashboard classes.

## 3.1 Panel dashboard

Purpose:

* always-on compact display
* optimized for 1920×440 panel use
* low cognitive load

Audience:

* ambient monitoring
* at-a-glance machine health

## 3.2 Operations dashboard

Purpose:

* richer desktop browser view
* host comparison
* historical trends
* operational troubleshooting

Audience:

* desktop/laptop browser users
* multi-host monitoring

## 3.3 Discovery dashboard

Purpose:

* hardware bring-up
* rule verification
* unknown device/sensor inspection
* taxonomy debugging

Audience:

* developer/operator during hardware onboarding

---

# 4. Dashboard-facing identity model

Dashboards MUST assume these labels are the primary stable selectors:

* `host`
* `device_class`
* `device_id`
* `logical_name`
* `sensor`
* `component`
* `vendor`
* `model`
* `source`

Where available, dashboards SHOULD also use:

* `device_subclass`
* `site`
* `role`
* `environment`

Dashboards MUST NOT rely primarily on:

* raw source names
* source-specific ordering assumptions
* ephemeral bus paths not normalized into `device_id`

---

# 5. Canonical grouping rules

## 5.1 Group by host

Use for:

* multi-machine operations view
* desktop dashboards
* compare-one-host-to-another patterns

## 5.2 Group by device class

Use for:

* GPU strips
* PSU cards
* storage sections
* service/LLM sections

## 5.3 Group by logical name

Use for:

* stable single-sensor panels
* top cards and stat panels
* device-specific sparklines

## 5.4 Group by component + sensor

Use for:

* generalized dashboards
* discovery templates
* query reuse across vendors/models

---

# 6. Canonical query expectations

The dashboard system MUST assume the following query patterns are first-class.

## 6.1 Single logical metric current value

Example intent:

* current `gpu0_core_temp`

Expected selector style:

* metric family + `logical_name`
* optionally scoped by `host`

## 6.2 All devices in a class on one host

Example intent:

* all GPU temps on host `sanctum`

Expected selector style:

* metric family + `host` + `device_class`

## 6.3 Compare same logical metric across hosts

Example intent:

* compare `psu_output_power` across multiple hosts

Expected selector style:

* metric family + `logical_name`
* grouped by `host`

## 6.4 Discovery by unknown source/category

Example intent:

* show all unknown devices or unusual sensors

Expected selector style:

* discovery endpoint/catalog-driven, not purely metric query driven

---

# 7. Dashboard template variables

Grafana and similar dashboard systems SHOULD define template variables using these stable dimensions:

* `host`
* `device_class`
* `logical_name`
* `vendor`
* `model`
* `source`

Recommended usage:

* panel dashboard: minimal or no visible template variables
* ops dashboard: host/device class/logical name selectors
* discovery dashboard: host/source/device class selectors

---

# 8. Panel dashboard contract

## 8.1 Purpose and constraints

The panel dashboard is optimized for:

* 1920×440 stretched display
* kiosk mode
* low-overhead rendering
* large, readable values

The panel dashboard SHOULD prioritize:

* stat panels
* short sparklines
* compact layout
* low panel count

## 8.2 Canonical layout regions

Recommended sections:

* GPU region
* PSU region
* Host/system region
* Optional service/LLM region

## 8.3 GPU region contract

The dashboard SHOULD support rendering one card per GPU using normalized queries.

Expected canonical values per GPU:

* core temp
* hotspot temp if available
* power
* utilization
* VRAM used
* fan RPM if available

Selection strategy:

* `device_class=gpu`
* group by `device_id` or stable logical device name

## 8.4 PSU region contract

Expected values:

* output power
* input power if available
* internal temp
* fan speed
* efficiency if derived or directly exposed

Selection strategy:

* `device_class=psu`

## 8.5 Host/system region contract

Expected values:

* CPU package temp
* RAM used
* storage usage or relevant storage temps
* network throughput summary
* uptime

## 8.6 Service / LLM region contract

If backend extensions are installed, the panel dashboard MAY include:

* requests in flight
* queue depth
* tokens/sec
* active model

Selection strategy:

* `device_class=service` or `device_class=llm`

---

# 9. Operations dashboard contract

## 9.1 Purpose

The operations dashboard provides:

* richer historical views
* cross-host comparisons
* more detailed time-series panels
* topology/coverage awareness

## 9.2 Required sections

At minimum:

* host overview
* GPU overview
* power/thermal overview
* storage/network overview
* optional service/LLM overview
* exporter health overview

## 9.3 Exporter health section

The operations dashboard MUST include exporter self-metrics such as:

* adapter failure counts
* last successful refresh
* refresh duration
* config reload status
* mapping failures

## 9.4 Cross-host patterns

The operations dashboard MUST support comparing:

* same metric across hosts
* same device class across hosts
* exporter health across hosts

---

# 10. Discovery dashboard contract

## 10.1 Purpose

This dashboard focuses on onboarding and debugging.

## 10.2 Required data sources

The discovery dashboard SHOULD use:

* `/debug/discovery`
* `/debug/mappings`
* `/debug/catalog`
* selected normalized metrics

## 10.3 Required sections

* discovered devices list
* unknown devices count/list
* mapping decisions summary
* recently dropped/unmapped items
* normalized catalog view

## 10.4 Contract with debug endpoints

The discovery dashboard or companion tooling MUST assume the JSON debug endpoints are the authoritative source for:

* discovery state
* mapping decisions
* catalog metadata

---

# 11. Canonical metric-to-panel mapping

## 11.1 Temperature metrics

Metric family:

* `hw_device_temperature_celsius`

Typical panels:

* stat
* sparkline stat
* time series
* threshold color panels

Primary selectors:

* `host`
* `device_class`
* `logical_name`
* `sensor`

## 11.2 Power metrics

Metric family:

* `hw_device_power_watts`

Typical panels:

* stat
* time series
* stacked comparisons by host/device where sensible

## 11.3 Fan metrics

Metric family:

* `hw_device_fan_speed_rpm`

Typical panels:

* stat
* compact table
* time series where useful

## 11.4 Memory metrics

Metric families:

* `hw_device_memory_used_bytes`
* `hw_device_memory_total_bytes`
* `hw_device_memory_bytes` where generic family is used

Typical panels:

* gauge-like percentage derived panel
* stat with total/used context

## 11.5 Service/LLM metrics

Potential metric families:

* `hw_service_requests_total`
* `hw_service_requests_in_flight`
* `hw_service_queue_depth`
* `hw_service_request_duration_seconds`
* `hw_llm_tokens_per_second`
* `hw_llm_context_tokens`
* `hw_llm_model_loaded`

Typical panels:

* stat
* rate panels
* queue sparkline
* active model info panel

---

# 12. Label usage rules for dashboards

## 12.1 Required dashboard assumptions

Dashboards MUST assume `logical_name` is the most stable selector for a single exposed conceptual sensor.

Dashboards MUST assume `device_class` is the most stable selector for broad sections.

## 12.2 Avoid over-filtering

Dashboards SHOULD avoid filtering on `source` unless building debugging or discovery views.

Reason:

* source adapters may evolve while normalized intent remains the same.

## 12.3 Use of `vendor` and `model`

Use vendor/model for:

* optional grouping
* legends
* discovery/debug
* side labels in rich desktop dashboards

Avoid making vendor/model required to basic panel functionality.

---

# 13. Query portability requirement

Dashboard queries MUST be portable between:

* embedded Pi server mode
* central server mode

This means dashboards MUST NOT depend on:

* host-local-only assumptions from the dashboard node
* raw exporter endpoint names
* Pi-specific display-only state

---

# 14. Performance guidance

## 14.1 Panel dashboard

The panel dashboard SHOULD be optimized for lower-end viewer devices such as Pi 3 by:

* limiting panel count
* preferring stat + sparkline over dense graphs
* avoiding excessive transformations
* minimizing animated/heavy plugins

## 14.2 Operations dashboard

The operations dashboard MAY be richer and more query-heavy, as it is expected to be viewed on more capable machines.

## 14.3 Discovery dashboard

The discovery dashboard MAY trade visual polish for richer debugging visibility.

---

# 15. Dashboard provisioning expectations

Dashboard bundles SHOULD be organized into:

* `panel`
* `ops`
* `discovery`

Provisioned dashboards MUST assume the normalized metrics and labels are the source of truth.

Provisioned datasource defaults SHOULD target Prometheus.

---

# 16. Example panel-to-query intent mapping

## 16.1 GPU core temp card

Intent:

* current temperature for each GPU core on selected host

Selection basis:

* `hw_device_temperature_celsius`
* `host=<selected>`
* `device_class=gpu`
* `sensor=core`

## 16.2 PSU output power card

Intent:

* current PSU output power

Selection basis:

* `hw_device_power_watts`
* `device_class=psu`
* `sensor=output`

## 16.3 LLM queue card

Intent:

* current local LLM queue depth

Selection basis:

* `hw_service_queue_depth`
* `device_class=llm` or `device_class=service`

## 16.4 Exporter health card

Intent:

* is exporter healthy and refreshing normally

Selection basis:

* exporter self-metrics prefixed `hwexp_`

---

# 17. Backward compatibility rules for dashboards

A dashboard contract change is breaking if it depends on:

* renamed normalized metric families
* removed required labels
* changed logical-name generation semantics

Dashboard bundles SHOULD version alongside schema compatibility declarations.

---

# 18. Testing obligations from this guide

Dashboard-facing tests SHOULD verify:

* expected query selectors resolve against fixture-backed data
* panel dashboards can populate all expected sections from canonical labels
* ops dashboards tolerate multi-host data
* discovery views show unknown devices and mapping issues where applicable
* optional LLM/custom-data sections appear only when relevant metrics exist 

---

# 19. Immediate follow-on implementation artifact

After this guide, the next useful implementation artifact is a **Dashboard Provisioning Pack** containing actual Grafana folder structure, datasource provisioning, and starter dashboards for panel, ops, and discovery layouts.



