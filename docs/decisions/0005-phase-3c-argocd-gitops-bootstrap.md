# 0005: Phase 3C Argo CD GitOps Bootstrap

## Status

Accepted

## Context

After Gitea was deployed, the lab needed a GitOps controller so platform changes flow from Git instead of ad hoc `kubectl apply` commands.

The internal platform repo is:

```text
http://gitea.dclab.local/gitadmin/platform.git
```

## Decision

Install Argo CD with:

- Helm chart `argo/argo-cd` version `6.8.1`
- Argo CD image `harbor-01.dclab.local/argocd/argocd:v2.11.3`
- Redis image `harbor-01.dclab.local/argocd/redis:7.0.15-alpine`
- Dex disabled
- `server.insecure=true`
- HTTP ingress at `argocd.dclab.local`

Use an App-of-Apps pattern:

- `root-app` watches `apps/platform-apps`
- child applications live under `apps/platform-apps`
- workload manifests live under `apps/workloads`

The first child application is `gitops-smoke`, a low-risk workload that proves sync behavior without adopting existing stateful platform components.

## Consequences

- Argo CD is now the platform deployment engine.
- Future platform services should be added to the Gitea repo first, then reconciled by Argo CD.
- Longhorn and Gitea are intentionally not managed by Argo CD yet. Their adoption is deferred until exact manifests or Helm values are committed and reviewed.
- Dex/SSO is deferred until the identity phase.
