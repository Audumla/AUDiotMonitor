# Spec Style Guide

Use this style for all new and updated specs.

## Title Format

- Use: `# Spec <id> — <name>` for numbered specs.
- Use: `# <Component> <Topic> Spec` for non-numbered component docs.

## Required Front Matter Block

- `Status`
- `Date`
- `Project`
- `Covers`
- `Related`

## Required Section Order

1. Goal / problem statement
2. Scope and boundaries
3. Contracts (inputs/outputs)
4. Configuration and examples
5. Error handling and edge cases
6. Test/validation requirements
7. Implementation notes / rollout

## Examples

- Include at least one concrete YAML/JSON example when configuration is involved.
- Include at least one metric or endpoint example when telemetry behavior is involved.

## Agent Readability

- Keep sections component-scoped.
- Avoid mixing dashboard and exporter implementation details in one file.
- Link to component reading paths in `components/` when adding a new spec.
