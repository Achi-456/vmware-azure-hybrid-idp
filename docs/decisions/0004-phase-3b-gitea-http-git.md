# 0004: Phase 3B Gitea HTTP Git

## Status

Accepted

## Context

The lab needs an internal Git server before ArgoCD and Backstage are installed. Gitea provides a lightweight Git service that can run on the existing RKE2 cluster and use Longhorn for persistence.

The cluster already has a stable ingress VIP from Phase 3A:

```text
gitea.dclab.local -> 172.25.188.96
```

## Decision

Deploy Gitea in namespace `gitea` using:

- `harbor-01.dclab.local/gitea/gitea:1.22.1`
- `harbor-01.dclab.local/gitea/postgres:16.3`
- Longhorn PVCs for Gitea data and PostgreSQL data
- HTTP Ingress through the existing RKE2 NGINX controller

Use `postgres:16.3` instead of `bitnami/postgresql:16.3.0` because the Bitnami tag did not resolve from Docker Hub during the mirror step.

Use `Recreate` deployment strategy for Gitea. The Gitea pod uses one `ReadWriteOnce` Longhorn volume, and the default `RollingUpdate` strategy caused a multi-attach conflict during restart.

Defer Git SSH exposure. Phase 3B uses HTTP Git only.

## Consequences

- ArgoCD can use `http://gitea.dclab.local/gitadmin/platform.git` in Phase 3C.
- The platform monorepo is now internal to the lab instead of depending on GitHub.
- Gitea restarts are clean because only one pod mounts the RWO volume at a time.
- SSH clone support requires a later TCP routing design.
