# Modernization Plan (Staged)

**Status:** Proposed (Next Stage; no baseline behavior change by default)  
**Date:** 2026-03-29  
**Project:** AUDiotMonitor  
**Covers:** Netdata-inspired platform hardening and extensibility, combined with existing planned updates  
**Related:** [optimization-review.md](optimization-review.md), [next-stage-metrics-schema-model-update.md](next-stage-metrics-schema-model-update.md), [implementation-plan.md](../exporter/implementation-plan.md), [dashboard-data-spec.md](../dashboard/dashboard-data-spec.md)

> This plan is execution-oriented and staged. Each stage has clear acceptance criteria and can be delivered independently.

---

## 1. Goals

1. Improve reliability (prevent scrape/config regressions).
2. Make schema evolution safe with direct cutover discipline.
3. Make collectors more modular and extensible.
4. Ensure dashboards migrate in lockstep with metric changes.
5. Keep deployments artifact-only and reproducible.

---

## 2. Scope and Boundaries

In scope:
- metric schema/model evolution process
- dashboard query migration process
- collector architecture improvements inspired by Netdata patterns
- CI/release guardrails and validation

Out of scope:
- replacing the stack with Netdata
- changing baseline docs/contracts in a single large cutover

---

## 3. Staged Delivery Plan

## Stage 0: Guardrails First (Immediate)

### Deliverables
- CI check: fail on invalid mapping `metric_type` values.
- CI check: fail if Prometheus scrape of `hwexp` is down in integration run.
- CI check: production compose profiles cannot use mutable runtime tags.
- Release smoke checks:
  - `hwexp --version`
  - required runtime binaries by enabled adapters (for example `smartctl`)

### Acceptance Criteria
- A bad mapping type cannot be merged.
- `up{job="hwexp"}` regression is caught before release.
- Release workflow artifacts include validation summary.

---

## Stage 1: Schema Migration Framework

### Deliverables
- Adopt the three-layer model from `next-stage-metrics-schema-model-update.md`:
  - `raw_*` (diagnostic)
  - `hw_*` (canonical)
  - `audiot_*` (derived/recording)
- Add schema migration matrix template:
  - `old metric/query` -> `new metric/query` -> `affected dashboards` -> `cutover release`
- Add direct-cutover policy:
  - schema and dashboard updates ship in the same release

### Acceptance Criteria
- Every schema change PR includes migration matrix entries.
- No migration entry may rely on compatibility aliases.

---

## Stage 2: Dashboard Migration Program (Mandatory)

### Deliverables
- Dashboard query migration matrix for all active dashboards.
- Update dashboard JSON under `monitoring/dashboard/dashboards/` to canonical/derived queries.
- Add dashboard golden checks for:
  - multi-GPU VRAM used/capacity/percent
  - critical system overview panels

### Acceptance Criteria
- No dashboard ships with pre-cutover queries after release.
- Golden checks pass in CI for critical dashboards.
- Cutover release gates on dashboard verification.

---

## Stage 3: Collector Architecture Upgrades (Netdata-Inspired)

### Deliverables
- Per-adapter job model (multiple job instances per adapter type).
- Adapter capability contract:
  - required tools
  - optional families
  - platform support
- Readiness/reporting includes unmet capability details.

### Acceptance Criteria
- One adapter can run multiple configured jobs cleanly.
- Missing dependencies produce explicit degraded status and logs.

---

## Stage 4: Process Isolation + Resilience

### Deliverables
- Out-of-process collector execution option for high-risk sources.
- Supervisor lifecycle:
  - start/stop/restart policy
  - timeout and backoff
- Exporter internal health metrics:
  - queue depth
  - dropped samples
  - per-adapter poll latency and failures

### Acceptance Criteria
- Collector crash does not take down `hwexp`.
- Health metrics identify failing pipeline segments quickly.

---

## Stage 5: Config and Validation Hardening

### Deliverables
- Config preflight command (schema + mapping + rules + dashboard query checks).
- Label contract validator for canonical `hw_*` families.
- “breaking change detector” for family/required-label changes.

### Acceptance Criteria
- Preflight is required in CI for release branches.
- Breaking changes without migration path are blocked.

---

## Stage 6: Modularization Refactor

### Deliverables
- Extract template engine from mapper into dedicated package.
- Isolate join logic in `internal/engine/join`.
- Split config surface into typed sections (server, adapters, mapping, debug).

### Acceptance Criteria
- Core modules have independent unit tests.
- Join and mapping behavior can be validated without full runtime.

---

## 4. Cross-Stage Dependencies

1. Stage 0 is required before Stage 2+.
2. Stage 1 and Stage 2 must ship together for any schema rename.
3. Stage 3 capability contract should land before Stage 4 isolation.
4. Stage 5 preflight should gate Stage 6 refactors.

---

## 5. Milestone Checklist

1. **Milestone A (Safety):** Stage 0 complete.
2. **Milestone B (Schema + Dashboards):** Stage 1 and Stage 2 complete.
3. **Milestone C (Extensibility):** Stage 3 and Stage 4 complete.
4. **Milestone D (Maintainability):** Stage 5 and Stage 6 complete.

---

## 6. Risks and Mitigations

- Risk: dashboard breakage during schema migration.
  - Mitigation: migration matrix + golden checks + same-release cutover.
- Risk: collector complexity increases too quickly.
  - Mitigation: ship job model and capability contract before process isolation.
- Risk: hidden runtime dependencies.
  - Mitigation: capability checks + release smoke matrix.

---

## 7. Initial Execution Order (Recommended)

1. Implement Stage 0 checks.
2. Build schema migration matrix template and policy (Stage 1).
3. Run dashboard migration matrix and patch high-priority dashboards (Stage 2).
4. Introduce capability contract and readiness surfacing (Stage 3).
5. Continue with resilience, preflight, and modularization stages.
