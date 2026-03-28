# Optimization Review (Build + Modularization)

**Date:** 2026-03-28  
**Scope:** Release workflow, image build path, and `hwexp` module boundaries

---

## 1. Build and Release Pipeline Optimizations

### 1.1 Artifact-First Deploy Guardrails
- Enforce release-only deployment by requiring image tags that match GitHub releases (avoid mutable `latest` in production compose).
- Add CI check that fails deployment PRs if compose files reference non-release tags for production profiles.

### 1.2 Faster Multi-Arch Builds
- Keep builder on `$BUILDPLATFORM` with cross-compilation (already in place).
- Add Docker Buildx registry cache (`cache-from`/`cache-to`) in release workflow to reduce repeated `go mod download`.
- Split dependency copy (`go.mod`, `go.sum`) from source copy (already present) and ensure workflow cache key includes both files only.

### 1.3 Runtime Image Slimming
- Keep runtime minimal but include operational dependencies required by enabled adapters:
  - `smartmontools` (required by `linux_storage`)
  - `dmidecode`, `pciutils` (already present)
- Add a smoke test in CI that validates required binaries in final image:
  - `hwexp --version`
  - `which smartctl`

### 1.4 Release Validation Matrix
- Add post-build integration checks using released container image:
  - `/debug/discovery` includes `gateway_manifest`, `linux_storage`, `linux_system` when enabled.
  - `/metrics` includes one expected metric family per adapter in fixture mode.
- Publish validation summary as release artifact.

---

## 2. Code Modularization Optimizations

### 2.1 Mapper Templating as Dedicated Package
- Extract template expansion logic from `internal/mapper` into a dedicated package (for example `internal/templatex`).
- Benefits:
  - Single implementation for `${logical_device_name}`, metadata placeholders, and regex capture placeholders.
  - Easier unit testing and reuse in future config/template surfaces.

### 2.2 Adapter Capability Contracts
- Introduce an adapter capability interface (for example `Capabilities() []Capability`) to expose:
  - required host tools (`smartctl`)
  - optional metrics families
  - discovery/poll support
- Benefits:
  - better readiness messaging
  - explicit dependency checks before runtime failures

### 2.3 Config Surface Decomposition
- Split `hwexp.yaml` parsing into typed sections by concern:
  - server/runtime
  - adapter registry and settings
  - mapping/join behavior
  - debug endpoints
- Benefits:
  - smaller validation units
  - cleaner schema evolution
  - reduced cross-field coupling

### 2.4 Join Engine Isolation
- Move hardware/software label-join logic behind a dedicated interface in `internal/engine/join`.
- Keep adapter outputs pure and join behavior centralized.
- Benefits:
  - makes correlation strategy testable without full adapter stack
  - supports future join strategies (PCI slot, device UUID, explicit manifest key maps)

---

## 3. Recommended Next Iteration (Priority Order)

1. Pin production compose to release tags (artifact-first safety).
2. Add CI image smoke checks for runtime dependencies (`smartctl`).
3. Extract mapper template expansion into reusable package + tests.
4. Add adapter capability contract and readiness reporting.
5. Refactor join engine into isolated module.
