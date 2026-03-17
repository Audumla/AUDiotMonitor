# Hardware Telemetry Exporter Platform

## Exporter Interface Specification

**Document purpose**: This document defines the code-level interfaces, package boundaries, lifecycle hooks, concurrency model, and implementation contracts for the exporter runtime.

**Status**: Normative for implementation

**Audience**: Developers implementing the exporter daemon, adapters, bridges, mapping engine, HTTP server, fixture mode, and tests

---

# 1. Scope

This document defines:

* package/module responsibilities
* adapter interfaces
* runtime lifecycle stages
* refresh/discovery concurrency model
* state snapshot model
* fixture mode behavior
* extension points for future platforms and bridges

This document does not define packaging, installer behavior, or dashboard layouts.

---

# 2. Language and runtime assumptions

The primary implementation language is **Go**.

The exporter MUST be buildable as a single long-running daemon binary.

The exporter MUST support:

* normal runtime mode
* fixture-backed simulation mode
* config validation mode
* optional dry-run/discovery inspection mode

---

# 3. Package boundaries

Recommended internal package layout:

```text
cmd/
  hwexp/
  hwexp-fixture-capture/
internal/
  app/
  config/
  model/
  taxonomy/
  adapters/
  bridges/
  discovery/
  polling/
  mapper/
  snapshot/
  httpapi/
  auth/
  selfmetrics/
  logging/
  validation/
```

## 3.1 `app`

Responsible for:

* process startup
* lifecycle orchestration
* config load/reload
* signal handling
* main run loop ownership

## 3.2 `config`

Responsible for:

* config file loading
* schema validation integration
* typed config conversion
* reload semantics

## 3.3 `model`

Responsible for canonical runtime types:

* `DiscoveredDevice`
* `RawMeasurement`
* `NormalizedMeasurement`
* `MappingDecision`
* `AdapterHealth`
* `Snapshot`

## 3.4 `taxonomy`

Responsible for:

* canonical device classes
* components
* sensor naming constants
* validation helpers

## 3.5 `adapters`

Responsible for platform-native discovery/polling.

## 3.6 `bridges`

Responsible for scraping/translating other exporters or metrics endpoints.

## 3.7 `discovery`

Responsible for discovery-cycle orchestration and registry update logic.

## 3.8 `polling`

Responsible for polling orchestration, timeouts, and per-adapter execution.

## 3.9 `mapper`

Responsible for:

* mapping rule evaluation
* normalization
* label generation
* logical name generation
* precedence resolution

## 3.10 `snapshot`

Responsible for:

* building complete runtime snapshots
* atomic swap of current state
* exposing read-only current snapshot handle

## 3.11 `httpapi`

Responsible for:

* HTTP routing
* `/metrics`
* JSON debug endpoints
* auth middleware integration
* request correlation IDs

## 3.12 `auth`

Responsible for:

* bearer token auth
* scope validation
* future reverse proxy / mTLS hooks

## 3.13 `selfmetrics`

Responsible for exporter self-observability metrics.

## 3.14 `logging`

Responsible for:

* structured JSON logs
* request context enrichment
* rate-limited repeated events

## 3.15 `validation`

Responsible for:

* runtime semantic checks not captured by JSON Schema
* strict-mode checks
* conflict detection

---

# 4. Canonical runtime interfaces

## 4.1 DiscoveryAdapter

Purpose: enumerate devices known to a source.

Conceptual interface:

```go
type DiscoveryAdapter interface {
    Name() string
    Platform() string
    Priority() int
    Init(ctx context.Context, cfg AdapterConfig) error
    Discover(ctx context.Context) ([]model.DiscoveredDevice, error)
    Close(ctx context.Context) error
}
```

### Requirements

* `Name()` MUST be stable and match config/metrics source naming
* `Priority()` MUST support source precedence decisions
* `Discover()` MUST be safe to call multiple times
* `Close()` MUST be idempotent

## 4.2 PollAdapter

Purpose: poll raw measurements for a discovered device or source domain.

Conceptual interface:

```go
type PollAdapter interface {
    Name() string
    Platform() string
    Priority() int
    Init(ctx context.Context, cfg AdapterConfig) error
    Poll(ctx context.Context, devices []model.DiscoveredDevice) ([]model.RawMeasurement, error)
    Close(ctx context.Context) error
}
```

### Requirements

* `Poll()` MUST obey context cancellation and timeout
* returned measurements MUST identify stable device associations where possible
* `Poll()` MUST NOT mutate shared global state

## 4.3 BridgeAdapter

Purpose: ingest and translate existing exporter or metrics endpoint data.

Conceptual interface:

```go
type BridgeAdapter interface {
    Name() string
    Platform() string
    Priority() int
    Init(ctx context.Context, cfg AdapterConfig) error
    ScrapeAndBridge(ctx context.Context) ([]model.RawMeasurement, error)
    Close(ctx context.Context) error
}
```

## 4.4 MappingEngine

Purpose: convert raw measurements into normalized output.

Conceptual interface:

```go
type MappingEngine interface {
    Map(ctx context.Context, devices []model.DiscoveredDevice, raw []model.RawMeasurement) (mapped []model.NormalizedMeasurement, decisions []model.MappingDecision, err error)
}
```

## 4.5 SnapshotStore

Purpose: hold current immutable runtime state.

Conceptual interface:

```go
type SnapshotStore interface {
    Current() model.Snapshot
    Replace(next model.Snapshot)
}
```

`Current()` MUST return an immutable snapshot reference from the caller’s perspective.

---

# 5. Canonical runtime types

## 5.1 Snapshot

A `Snapshot` MUST contain:

* `generated_at`
* `devices`
* `raw_measurements`
* `normalized_measurements`
* `mapping_decisions`
* `adapter_health`
* `config_state`
* `schema_versions`

The snapshot MUST be treated as immutable after publication.

## 5.2 AdapterHealth

Required fields:

* `adapter_name`
* `state`
* `last_success`
* `last_error_code`
* `last_error_message`
* `consecutive_failures`
* `last_duration`

Allowed `state`:

* `ready`
* `degraded`
* `failed`
* `disabled`
* `initializing`

---

# 6. Runtime lifecycle model

## 6.1 Startup phases

The exporter MUST start in this order:

1. parse CLI args
2. load main config
3. validate config semantically
4. load mappings/aliases/tokens if configured
5. initialize logger
6. initialize self-metrics
7. initialize adapters
8. start HTTP listener
9. run initial discovery
10. run initial polling
11. build first snapshot
12. mark readiness true
13. enter steady-state loops

## 6.2 Steady-state loops

The exporter MUST maintain at least:

* discovery loop
* refresh loop
* optional config reload handler
* HTTP serving loop

## 6.3 Shutdown phases

On shutdown signal:

1. mark readiness false
2. stop accepting new internal work
3. cancel active discovery/poll jobs
4. close adapters
5. stop HTTP server gracefully
6. flush logs if needed
7. exit with appropriate status

---

# 7. Discovery model

## 7.1 Discovery orchestration

Discovery MUST:

* run on configured interval
* gather records from all enabled discovery adapters
* deduplicate devices by stable ID
* preserve previous known device state where configured

## 7.2 Discovery conflict handling

If multiple adapters produce the same `stable_id`:

* the registry entry MUST preserve merged knowledge where safe
* source precedence MUST determine canonical source fields where conflict exists
* the conflict MUST be visible in debug and logs if material

## 7.3 Unknown-device handling

Unknown devices MUST remain discoverable and visible in debug outputs unless filtered by explicit policy.

---

# 8. Polling model

## 8.1 Poll orchestration

Polling MUST:

* execute on configured refresh interval
* use the latest device registry snapshot
* invoke enabled poll and bridge adapters
* collect raw measurements
* pass them to the mapping engine
* publish a complete replacement snapshot atomically

## 8.2 Timeout model

Each adapter execution MUST be bounded by timeout.

Context cancellation MUST be propagated to adapters.

## 8.3 Failure isolation

A failing adapter MUST NOT block publication of other successful adapter data.

## 8.4 Last-good retention

If an adapter fails, the previous successful measurements for that adapter MAY be retained subject to grace rules from the contracts spec.

---

# 9. Concurrency model

## 9.1 General rules

The exporter MUST use concurrency only where it does not compromise snapshot consistency.

## 9.2 Adapter concurrency

Discovery and polling MAY run adapters concurrently, subject to:

* configured max concurrency
* deterministic merge rules
* immutable input snapshots for each cycle

## 9.3 Snapshot publication

Snapshot publication MUST be atomic.

The `/metrics` and JSON endpoints MUST always observe a single coherent published snapshot, never a partially constructed one.

## 9.4 Reload concurrency

Config reload MUST NOT corrupt active snapshot state.

A safe pattern is:

* validate new config fully
* initialize replacement runtime dependencies
* switch active runtime references only after successful preparation

---

# 10. Mapping engine behavior

## 10.1 Inputs

The mapping engine MUST receive:

* current discovered devices
* current raw measurements
* loaded mapping rules
* loaded aliases
* static identity labels

## 10.2 Outputs

The mapping engine MUST return:

* normalized measurements
* mapping decisions
* fatal error only for cycle-level unrecoverable mapping issues

Non-fatal per-measurement mapping failures SHOULD become mapping decisions rather than full-cycle fatal errors.

## 10.3 Rule evaluation

The engine MUST implement precedence exactly as defined in the contracts spec.

## 10.4 Label generation

The engine MUST:

* merge static labels
* apply canonical required labels
* apply rule-specified labels
* validate final label set

## 10.5 Logical names

Logical names MUST be generated deterministically.

Alias-derived logical device names MUST take precedence over fallback device name generation.

---

# 11. Metrics exposition model

## 11.1 `/metrics`

The HTTP metrics handler MUST expose:

* normalized hardware/application/service metrics
* exporter self-metrics
* build info metric if enabled

## 11.2 Metric generation

Metric families MUST be generated from the current immutable snapshot.

The exposition path MUST NOT trigger live sensor reads.

## 11.3 Self-metrics separation

Exporter self-metrics SHOULD use distinct families prefixed with `hwexp_`.

Normalized source metrics SHOULD use the normalized metric families defined by the taxonomy and mapping rules.

---

# 12. HTTP API implementation rules

## 12.1 Routing

Required routes:

* `/metrics`
* `/version`
* `/healthz`
* `/readyz`
* `/debug/discovery`
* `/debug/mappings`
* `/debug/catalog`
* `/debug/raw` if enabled

## 12.2 Middleware chain

Recommended middleware order:

1. request ID
2. panic recovery
3. auth/scope validation
4. gzip where applicable
5. handler execution
6. access logging

## 12.3 Error handling

JSON routes MUST return machine-readable errors shaped per the contracts spec.

---

# 13. Auth integration model

The HTTP layer MUST integrate with auth in a route-aware way.

The auth module MUST support:

* no-auth mode
* bearer token mode
* scope checks per route class

Future auth strategies MUST be addable without changing handler business logic.

---

# 14. Fixture mode

## 14.1 Purpose

Fixture mode exists to:

* support CI without live hardware
* support regression tests
* support local development on unsupported hardware

## 14.2 Behavior

In fixture mode:

* discovery adapters MAY be replaced by fixture-backed adapters
* poll/bridge adapters MAY read fixture files instead of live sources
* HTTP/API behavior MUST remain as close to normal runtime as possible

## 14.3 Determinism

Fixture mode outputs MUST be deterministic for the same fixture set and config.

---

# 15. Extension model

## 15.1 New adapter addition requirements

A new adapter MUST:

* implement the appropriate interface
* provide stable `Name()` and `Platform()` values
* integrate with config registration
* document required permissions and dependencies
* provide fixtures/tests

## 15.2 New bridge addition requirements

A new bridge MUST:

* define external source assumptions
* provide translation logic into raw measurements
* document source precedence expectations

## 15.3 Non-hardware extensions

The exporter MUST support non-hardware operational telemetry through the same model.

This includes optional components like:

* custom ingest
* local LLM observer

These sources SHOULD enter the runtime through bridge or poll-style adapters and still emit normalized measurements and mapping decisions.

---

# 16. Testing obligations from this spec

The implementation MUST include tests for:

* adapter init/close idempotency
* discovery deduplication
* polling timeout/cancellation
* atomic snapshot replacement
* fixture-mode determinism
* mapping engine precedence
* HTTP route responses against current snapshot only
* config reload safety
* failure isolation across adapters

---

# 17. Immediate next implementation focus

After this document, the next highest-value implementation document is the **Mapping Rules Reference**, because it defines the most variable and error-prone part of the normalization behavior in exhaustive detail.
