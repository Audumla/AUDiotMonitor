# AUDiotMonitor Specifications

This is the canonical specification set for AUDiotMonitor.

## Read Order (High Level to Implementation)

1. [Foundation Overview](foundation/overview.md)
2. [System Architecture (Spec 100)](foundation/spec-100-component-architecture.md)
3. [Installation and Lifecycle](foundation/installation-spec.md)
4. [Exporter Detailed Contract](exporter/detailed-specification.md)
5. [Adapter-Specific Specs](adapters/README.md)
6. [Dashboard Specs](dashboard/README.md)
7. [Implementation Plan](exporter/implementation-plan-stages.md)
8. [Flow Guide (Intent to Implementation)](flow-high-to-low.md)

## Component-Focused Entry Points (For Agents)

- [Collector + Exporter Path](components/collector.md)
- [Adapters Path](components/adapters.md)
- [Dashboard Path](components/dashboard.md)
- [Schemas + Mapping Path](components/schemas-and-mapping.md)
- [Release + Ops Path](components/release-and-operations.md)

## Folder Layout

- `foundation/` high-level design and architecture intent
- `exporter/` core contracts, interfaces, schema, and mapping rules
- `adapters/` adapter-specific design specs
- `dashboard/` Grafana/dashboard contracts and implementation notes
- `operations/` optimization and operational review notes
- `components/` quick-reading guides so agents can consume only required sections

## Authoring Standard

- [Spec Style Guide](DOC_STYLE.md)
