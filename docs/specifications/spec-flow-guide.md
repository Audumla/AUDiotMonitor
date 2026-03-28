# Spec Flow: High Level to Implementation

This file defines the canonical progression for writing or updating features.

## Phase 1: Intent

- Start in `foundation/overview.md` and `foundation/component-architecture.md`.
- Define the problem, boundaries, and ownership.

## Phase 2: Contract

- Define expected behavior in `exporter/detailed-spec.md`.
- Define interface and lifecycle in `exporter/exporter-interface-spec.md`.
- Define data validity in `exporter/json-schema-spec.md`.

## Phase 3: Component Details

- Adapter behavior: `adapters/`
- Dashboard behavior: `dashboard/`
- Mapping behavior: `exporter/mapping-rule-spec.md`

## Phase 4: Implementation and Rollout

- Capture implementation sequence in `exporter/implementation-plan.md`.
- Capture release/deployment tradeoffs in `operations/optimization-review.md`.

## Example (New Adapter Source Type)

1. Define source intent in `foundation/component-architecture.md`.
2. Add data/endpoint contract requirements in `exporter/detailed-spec.md`.
3. Add adapter-specific rules in `adapters/`.
4. Add mapping rules and examples in `exporter/mapping-rule-spec.md`.
5. Add rollout steps in `exporter/implementation-plan.md`.



