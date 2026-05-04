# Current Flow Config Export

This folder is a local copy of the files that drive the currently working Gitea Actions -> Harbor -> Argo CD -> RKE2 flow, plus the Phase 4A cert-manager TLS overlay.

No secret values are intentionally included here.

## What Is Live

```text
Push to Gitea platform repo
  -> Gitea Actions runner runs workflow
  -> Docker-in-Docker builds sample-app
  -> image is pushed to Harbor
  -> workflow commits image tag back to Git
  -> Argo CD syncs sample-app
  -> RKE2 runs the new pod
  -> cert-manager issues internal TLS for platform UIs
```

Current image pattern:

```text
harbor-01.dclab.local/platform/sample-app:<git-sha>
```

## Folder Layout

```text
current-flow-configs/
  platform-repo/
    .gitea/workflows/
    apps/root-app.yaml
    apps/platform-apps/
    apps/workloads/
    apps/sample-app/
  runtime-snapshots/
```

## Source Git Files

`platform-repo/` contains the active files copied from the internal Gitea repo:

```text
https://gitea.dclab.local/gitadmin/platform.git
```

Important files:

- `.gitea/workflows/ci-smoke.yaml`
- `.gitea/workflows/build-sample-app.yaml`
- `apps/root-app.yaml`
- `apps/platform-apps/gitea-runner.yaml`
- `apps/platform-apps/gitops-smoke.yaml`
- `apps/platform-apps/sample-app.yaml`
- `apps/platform-apps/cert-manager.yaml`
- `apps/platform-apps/cert-manager-issuers.yaml`
- `apps/platform-apps/platform-tls.yaml`
- `apps/workloads/gitea-runner/config.yaml`
- `apps/workloads/gitea-runner/statefulset.yaml`
- `apps/workloads/cert-manager-issuers/issuers.yaml`
- `apps/workloads/platform-tls/`
- `apps/workloads/sample-app/deployment.yaml`
- `apps/workloads/sample-app/service.yaml`
- `apps/sample-app/Dockerfile`
- `apps/sample-app/main.go`

## Runtime Snapshots

`runtime-snapshots/` contains cluster state exported from `utility-01`:

- `argocd-app-list.txt`
- `argocd-applications.yaml`
- `gitea-runner-statefulset-live.yaml`
- `gitea-runner-configmap-live.yaml`
- `gitea-deployment-actions-enabled-live.yaml`
- `sample-app-deployment-live.yaml`
- `sample-app-service-live.yaml`
- `gitops-smoke-live.yaml`
- `rke2-nodes.txt`
- `cert-manager-pods.txt`
- `cert-manager-clusterissuers.txt`
- `platform-tls-certificates.txt`
- `phase4a-https-checks.txt`
- `gitea-runner-pods.txt`
- `sample-app-pods.txt`

These are snapshots for review and troubleshooting. The Git source of truth remains the internal Gitea `platform` repo.

## Secrets Not Included

Secrets stay in:

```text
utility-01:/root/gitea-secrets.env
harbor-01:/root/.harbor-admin-password
```

Kubernetes secrets were not exported.
