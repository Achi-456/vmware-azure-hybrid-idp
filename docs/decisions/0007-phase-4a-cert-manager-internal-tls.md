# ADR 0007: Phase 4A cert-manager Internal TLS

## Status

Accepted

## Context

The platform had working HTTP ingress for Kubernetes-hosted services and a separate standalone HTTPS Harbor registry. The next requirement was to automate internal TLS for platform UIs without waiting for Vault PKI.

The services in scope were Gitea, Argo CD, and Longhorn because they are Kubernetes ingress workloads. Harbor was excluded because it runs outside Kubernetes on `harbor-01` through Docker Compose.

## Decision

Install cert-manager `v1.20.2` through Argo CD and use a self-signed lab root CA for Phase 4A.

Issuer model:

- `ClusterIssuer/dclab-selfsigned-bootstrap`
- `Certificate/cert-manager/dclab-root-ca`
- `ClusterIssuer/dclab-ca`

TLS hostnames:

- `gitea.dclab.local`
- `argocd.dclab.local`
- `longhorn.dclab.local`

Harbor remains on its existing standalone TLS certificate and CA until Phase 4B.

## Rationale

This gives the Kubernetes-hosted platform UIs automated certificate lifecycle now, while avoiding a risky mid-phase replacement of Harbor's working registry certificate. It also creates a clean bridge to Phase 4B, where Vault can become the long-term PKI backend.

Argo CD was moved to the HTTPS Gitea repo URL after Gitea HTTPS was enabled:

```text
https://gitea.dclab.local/gitadmin/platform.git
```

The older HTTP Argo CD repo entry was removed to avoid redirects and failed repository health checks.

## Consequences

Positive:

- Gitea, Argo CD, and Longhorn are available over HTTPS.
- Certificates are issued and renewed by cert-manager.
- Harbor remains stable during the transition.

Tradeoffs:

- The Phase 4A CA is a lab CA, not the final enterprise PKI.
- Windows browsers need the public root CA imported to avoid warnings.
- Harbor uses a separate CA until Phase 4B.

## Follow-Up

Phase 4B should evaluate Vault PKI as the shared issuer and decide whether Harbor should be migrated to a Vault-issued certificate.
