# ADR 0002: Crossplane Azure Provider Pinning And Azure Scope

## Status

Accepted (Phase 2D)

## Context

During Crossplane Phase 2D, initial Azure provider package tags (`v1.12.x`, `v1.11.x`) failed with `MANIFEST_UNKNOWN` from `xpkg.upbound.io`.

To keep the lab reproducible and avoid repeated package resolution failures, provider tags must be pinned to known-valid versions.

## Decision

Pin Crossplane Azure providers to:

- `xpkg.upbound.io/upbound/provider-family-azure:v1.1.0`
- `xpkg.upbound.io/upbound/provider-azure-resources:v1.1.0`
- `xpkg.upbound.io/upbound/provider-azure-dbforpostgresql:v1.1.0`

Crossplane core is pinned to:

- Helm chart `crossplane-stable/crossplane` version `1.15.2`

Azure scope policy for this lab:

- Service principal RBAC scope: subscription-level `Contributor` for initial lab bring-up.
- After Phase 3 compositions are stable, reduce scope to dedicated resource groups where possible.

## Consequences

- Provider install is stable and reproducible.
- Future upgrades must be explicit (test new OCI tags before changing pins).
- Current execution remains blocked until Entra permissions allow service principal creation or an existing SP credential JSON is provided.
