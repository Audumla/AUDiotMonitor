# Hardware Telemetry Exporter Platform - Installation Specification

**Document purpose**: This document specifies the installation, update, and lifecycle management of the AUDiot hardware telemetry platform components. It defines:

*   Installation profiles and their components
*   Installation, update, and uninstall flows
*   Compatibility and rollback policies
*   Service lifecycle management
*   Requirements for custom data ingestion and local LLM monitoring
*   Component manifest and release compatibility
*   Installer output, reporting, idempotency, and migration

**Status**: Normative for implementation

**Audience**: Developers implementing the installer/lifecycle manager, packagers, and operational teams managing deployments.

---

# 1. Normative language

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are to be interpreted as requirements.

Where this document conflicts with `Overview.md` or `detailed_specification.md`, this document takes precedence for installation and lifecycle details.

---

# 2. Scope of this document

This document defines the following key areas:

1.  **Installation Profiles**: Standard bundles of components for different deployment scenarios.
2.  **Lifecycle Flows**: Detailed steps for installation, updates, repairs, and uninstalls.
3.  **Custom Data Ingestion**: Requirements and architectural rules for extending the platform to non-hardware data.
4.  **LLM Monitoring**: Specific requirements for integrating local LLM observability.
5.  **Installer Behavior**: Policies for versioning, compatibility, rollback, idempotency, and output.
6.  **Linux-First Requirements**: Specific implementation needs for initial Linux/Raspberry Pi targets.

This document does **not** define Grafana dashboard JSON, Prometheus rule files, or packaging scripts in full detail (these are follow-on documents).

---

# 3. Installation profiles

Installation profiles define standard sets of components for different deployment roles. The lifecycle tool MUST support installation via these named profiles.

## 3.1 `exporter-node` profile

Installs:

*   `hwexp-exporter` (core component)

Default actions:

*   Install and configure the core exporter service.
*   Enable and start the service.
*   Ensure basic operation (`/healthz`, `/readyz`).

Use when:

*   Only hardware telemetry collection is needed on a source machine.
*   The Prometheus/Grafana stack is running elsewhere.

## 3.2 `embedded-pi-node` profile

Installs:

*   `hwexp-exporter`
*   `prometheus`
*   `grafana`
*   `hwexp-kiosk-browser` (optional)

Default actions:

*   Install all components on the Raspberry Pi.
*   Configure Prometheus to scrape the local exporter.
*   Configure Grafana for the local Prometheus and kiosk display.
*   Enable and start all services.

Use when:

*   A self-contained monitoring appliance is desired on a Raspberry Pi.
*   Limited number of hosts and dashboards.

## 3.3 `dashboard-node` profile

Installs:

*   `prometheus`
*   `grafana`
*   `hwexp-dashboard-templates`

Default actions:

*   Install and configure Prometheus and Grafana.
*   Provision standard dashboards.
*   Assume Prometheus will scrape remote `hwexp-exporter` instances.

Use when:

*   A central monitoring server is needed, separate from source machines.

## 3.4 `viewer-node` profile

Installs:

*   `hwexp-kiosk-browser` (optional)

Default actions:

*   Install and configure a browser in kiosk mode.
*   Point the browser to a configured Grafana instance.

Use when:

*   A dedicated display (e.g., Raspberry Pi) is used solely for viewing dashboards.

## 3.5 `dev-node` profile

Installs:

*   `hwexp-exporter`
*   `hwexp-fixture-capture`
*   `hwexp-config`
*   optional local `prometheus`
*   optional local `grafana`
*   `hwexp-sample-configs`
*   `hwexp-maintenance-tools`

Default actions:

*   Install exporter and fixture tools.
*   Install sample configs if absent.
*   Optionally install local dashboard stack for development.
*   Enable rapid validation and local test workflows.

---

# 4. Installation flows

## 4.1 First-time install flow

The lifecycle tool MUST execute first-time install in this order:

1.  Detect host OS and architecture.
2.  Resolve requested role and/or components.
3.  Resolve release version and manifest.
4.  Verify dependency graph.
5.  Create install/config/data directories.
6.  Download required artifacts.
7.  Verify checksums.
8.  Unpack/install artifacts.
9.  Render or copy default configs if requested.
10. Register services.
11. Update install inventory atomically.
12. Enable/start services as requested.
13. Perform post-install validation.
14. Print resulting status and next actions.

## 4.2 Update flow

The lifecycle tool MUST execute updates in this order:

1.  Load install inventory.
2.  Determine installed components.
3.  Resolve target versions for selected components.
4.  Verify compatibility against manifest and schema versions.
5.  Backup replaced binaries/configs unless skipped.
6.  Download and verify new artifacts.
7.  Stop affected services if required.
8.  Replace component files atomically where possible.
9.  Run migrations if defined.
10. Restart affected services.
11. Validate service health.
12. Update install inventory.
13. Report success/failure per component.

## 4.3 Repair flow

The repair flow MUST:

*   Compare install inventory with actual filesystem/service state.
*   Detect missing binaries, broken symlinks, missing units, absent dashboard files, or invalid config templates.
*   Reinstall or regenerate missing managed artifacts.
*   Preserve user-edited config files unless forced.

## 4.4 Uninstall flow

The uninstall flow MUST:

*   Identify selected components.
*   Stop and disable related services.
*   Remove managed files for those components.
*   Optionally preserve config and/or data.
*   Remove inventory entries for removed components.
*   Leave unrelated components untouched.

---

# 5. Update and compatibility policy

## 5.1 Component-targeted updates

The lifecycle tool MUST update only components that are:

*   Explicitly requested, or
*   Already installed when `--all-components` is used.

It MUST NOT install new components during update unless explicitly instructed.

## 5.2 Version pinning

The installer MUST support:

*   Latest install.
*   Exact version install.
*   Exact version update.
*   Pinned component versions in inventory or config.

## 5.3 Major version changes

Major version upgrades MUST require explicit opt-in, for example:

*   `--allow-major-upgrade`

## 5.4 Schema compatibility checks

Before update, the lifecycle tool MUST validate:

*   Config schema compatibility.
*   Dashboard bundle compatibility with Grafana provisioning rules.
*   Exporter compatibility with existing service and config layout.

If incompatible without migration support:

*   Update MUST abort.
*   The user MUST receive a clear explanation.

## 5.5 Rollback support

Rollback SHOULD be supported by retaining:

*   Previous downloaded artifacts, or
*   Previous extracted install directories, or
*   Both.

Rollback MUST restore:

*   Component binaries/files.
*   Service definitions if changed.
*   Inventory version state.

Rollback of mutable data stores is best-effort and MUST be clearly documented where not guaranteed.

---

# 6. Service lifecycle management

## 6.1 Managed services

Typical Linux services include:

*   `hwexp-exporter.service`
*   `prometheus.service`
*   `grafana-server.service`
*   optional kiosk launcher service or autostart unit

## 6.2 Service rules

The lifecycle tool MUST:

*   Create service definitions only for installed service-bearing components.
*   Avoid overwriting user-modified service files unless forced.
*   Support enable, disable, start, stop, restart actions as needed.

## 6.3 Service health validation

After install or update, the lifecycle tool MUST check:

*   Service is enabled if requested.
*   Service is active if start requested.
*   Service-specific readiness endpoint or process health where available.

Examples:

*   Exporter: `/healthz` or process state.
*   Prometheus: web endpoint reachable.
*   Grafana: web endpoint reachable.

---

# 7. Custom data ingestion requirement

The platform MUST support **custom data sources** beyond hardware telemetry.

This requirement is normative and affects both the exporter platform and installation model.

## 7.1 Purpose

Custom data support is required so users can inject additional observability data that is not tied directly to standard hardware sensors.

Examples:

*   Application-defined metrics.
*   Local automation state.
*   Job runner activity.
*   Batch pipeline progress.
*   Home Assistant state mirrors.
*   Custom counters and gauges.
*   Local LLM server activity.

## 7.2 Architectural rule

Custom data ingestion MUST be treated as a first-class extension path, not as a special-case hack.

## 7.3 Supported custom data component

A distinct component MUST exist:

*   `hwexp-custom-ingest`

Responsibilities:

*   Accept or collect custom data from approved sources.
*   Normalize it into stable metric families.
*   Expose it to Prometheus and Grafana alongside other telemetry.

## 7.4 Supported ingestion models

The design MUST support these models over time:

### Pull collector model

A local script/adapter polls an application or endpoint and emits normalized metrics.

### File drop / text ingestion model

An application writes machine-readable data to a known directory, and the ingest layer converts it.

### HTTP ingestion model

A trusted local process posts structured telemetry to a local endpoint, which the ingest layer validates and republishes.

### Exporter bridge model

The ingest layer scrapes an existing exporter or metrics endpoint and remaps it.

For v1, the preferred low-risk models are:

*   Pull collector model.
*   Exporter bridge model.
*   File drop model.

---

# 8. Local LLM monitoring requirement

The platform MUST support **local LLM activity monitoring** as a supported custom-data use case.

## 8.1 Scope

The goal is to monitor local model-serving activity and related workload behavior, not just raw machine sensors.

Examples of desired metrics:

*   Active model name.
*   Active backend name.
*   Request count.
*   In-progress requests.
*   Prompt tokens per second.
*   Completion tokens per second.
*   Context size in use.
*   Queue depth.
*   Request latency.
*   GPU allocation by serving process if available.
*   Model load/unload events.
*   Memory pressure related to LLM serving.

## 8.2 Design rule

Local LLM observability MUST use the same normalization and taxonomy principles as hardware telemetry.

## 8.3 Additional taxonomy classes

The canonical taxonomy MUST be extended to include non-hardware operational classes:

*   `service`
*   `llm`
*   `inference`
*   `queue`
*   `workload`

## 8.4 Example metric families for LLM data

Examples include:

*   `hw_service_requests_total`
*   `hw_service_requests_in_flight`
*   `hw_service_queue_depth`
*   `hw_service_request_duration_seconds`
*   `hw_llm_tokens_per_second`
*   `hw_llm_context_tokens`
*   `hw_llm_model_loaded`
*   `hw_llm_backend_info`

These names may later be refined into a broader non-hardware prefix if the platform evolves beyond hardware-centric scope, but the install and lifecycle model MUST already accommodate these components.

## 8.5 Componentization

A separate optional component SHOULD exist:

*   `hwexp-llm-observer`

Responsibilities:

*   Poll or bridge local LLM servers/processes.
*   Normalize metrics from supported local LLM backends.
*   Expose metrics to Prometheus via exporter integration or a dedicated metrics endpoint.

## 8.6 Supported observation methods

The design MUST allow:

*   Scraping existing local LLM metrics endpoints if available.
*   Parsing local structured logs if configured.
*   Polling local APIs for model/server status.
*   Process-level correlation where safe and supported.

The lifecycle tool MUST treat `hwexp-llm-observer` as optional and independently installable.

---

# 9. Extended profiles with custom/LLM support

## 9.1 `exporter-node` with custom ingest

Optional components:

*   `hwexp-custom-ingest`
*   `hwexp-llm-observer`

Use when:

*   Source host runs local applications or LLM services worth monitoring.

## 9.2 `dashboard-node` with extended dashboards

Optional components:

*   Extended dashboard bundle including application and LLM dashboards.

## 9.3 `dev-node` with LLM observability

Optional components:

*   `hwexp-custom-ingest`
*   `hwexp-llm-observer`
*   Local dashboard bundle with service and LLM views.

---

# 10. Component manifest and release compatibility

The release manifest MUST support component compatibility declarations.

Each component entry SHOULD include:

*   `requires`
*   `conflicts`
*   `min_config_schema`
*   `max_config_schema`
*   `min_supported_platform_version`

### Example

```json
{
  "name": "hwexp-llm-observer",
  "version": "0.1.0",
  "requires": ["hwexp-exporter"],
  "artifacts": [
    {
      "os": "linux",
      "arch": "amd64",
      "format": "tar.gz",
      "file": "hwexp-llm-observer-0.1.0-linux-amd64.tar.gz",
      "sha256": "..."
    }
  ]
}
```

---

# 11. Installer output and reporting

## 11.1 Human-readable output

The lifecycle tool MUST clearly report:

*   Selected role.
*   Selected components.
*   Resolved versions.
*   Installed paths.
*   Service changes.
*   Next steps.

## 11.2 Machine-readable output

The lifecycle tool SHOULD support JSON output for automation.

Recommended flag:

*   `--output json`

## 11.3 Dry-run mode

Dry-run MUST show:

*   Components that would be installed/updated.
*   Versions to be used.
*   Files/services that would change.
*   Whether migrations would run.

Dry-run MUST NOT modify the system.

---

# 12. Idempotency requirements

The installer and updater MUST be idempotent where practical.

Examples:

*   Re-running install for already installed matching version MUST result in no destructive changes.
*   Re-running profile application MUST not duplicate units or inventory entries.
*   Re-running dashboard provisioning MUST refresh managed assets without duplicating unmanaged ones.

---

# 13. Migration requirements

When component layout, config layout, or service definitions change between versions, the lifecycle system MUST support explicit migration steps.

Migration steps MAY include:

*   File moves.
*   Config transforms.
*   Service renames.
*   Dashboard provisioning updates.
*   Inventory schema upgrades.

Migrations MUST be:

*   Versioned.
*   Logged.
*   Restart-safe where possible.
*   Reversible where practical.

---

# 14. Linux-first implementation requirements

For Linux and Raspberry Pi targets, the first implementation MUST include:

*   Shell install bootstrapper or packaged `hwexpctl`.
*   GitHub Releases download support.
*   Checksum verification.
*   Install inventory creation.
*   Systemd integration.
*   Profile-based installs.
*   Update command with component targeting.
*   Exporter-node and embedded-pi-node support.
*   Optional dashboard-node and viewer-node support.
*   Optional custom-ingest and llm-observer component registration hooks.

---

# 15. Immediate follow-on documents

After this document, the next required lifecycle/ops documents are:

1.  **Installer CLI Reference**
    *   Exact command syntax.
    *   Examples.
    *   Exit codes.

2.  **Install Inventory Schema**
    *   Formal JSON schema.

3.  **Release Manifest Schema**
    *   Formal JSON schema.

4.  **Linux Installation Runbook**
    *   Exporter-node.
    *   Dashboard-node.
    *   Viewer-node.
    *   Embedded-pi-node.

5.  **Service and File Layout Reference**
    *   Exact managed files.
    *   Ownership and permissions.

6.  **Custom Ingest and LLM Observer Spec**
    *   Component contracts.
    *   Metric families.
    *   Ingestion models.

---

# 16. Implementation readiness statement

This document is intended to be sufficient to begin designing and implementing:

*   The installation/lifecycle manager.
*   Release artifact structure.
*   Component inventory tracking.
*   Install/update flows.
*   Role/profile logic.
*   Future custom-ingest and LLM-observer component handling.

The highest-value next artifact after this is the **Installer CLI Reference plus the Release Manifest Schema**, because those will lock down the exact operational interface and GitHub release contract used by the install/update system.