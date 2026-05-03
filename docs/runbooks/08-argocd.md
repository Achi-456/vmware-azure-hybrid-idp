# Phase 3C: Argo CD

Argo CD is the GitOps control plane for the lab. It watches the internal Gitea `platform` repository and applies declared platform state to the RKE2 cluster.

## Current Endpoint

- URL: `http://argocd.dclab.local`
- Ingress VIP: `172.25.188.96`
- Namespace: `argocd`
- Admin username: `admin`
- Admin password file: `/root/gitea-secrets.env` on `utility-01`

Read the password only on `utility-01`:

```bash
source /root/gitea-secrets.env
echo "$ARGOCD_ADMIN_PASSWORD"
```

Do not copy this value into Git-tracked files.

## Versions And Images

Helm chart:

- `argo/argo-cd` chart version `6.8.1`

Application image:

- Argo CD `v2.11.3`

Mirrored Harbor images:

```text
harbor-01.dclab.local/argocd/argocd:v2.11.3
harbor-01.dclab.local/argocd/redis:7.0.15-alpine
```

Dex is disabled in this phase. The original `ghcr.io/dex-idp/dex:v2.39.1` source image returned an access error during preflight, and SSO is not required until the identity phase.

## Install Values

Argo CD runs behind the existing RKE2 NGINX ingress over internal HTTP:

```yaml
global:
  domain: argocd.dclab.local
  image:
    repository: harbor-01.dclab.local/argocd/argocd
    tag: v2.11.3

configs:
  params:
    server.insecure: true

dex:
  enabled: false

redis:
  image:
    repository: harbor-01.dclab.local/argocd/redis
    tag: 7.0.15-alpine

server:
  ingress:
    enabled: true
    ingressClassName: nginx
    hostname: argocd.dclab.local
    tls: false
    annotations:
      nginx.ingress.kubernetes.io/backend-protocol: HTTP
```

## Gitea Repository Connection

Repository:

```text
http://gitea.dclab.local/gitadmin/platform.git
```

Add or refresh the repo connection:

```bash
source /root/gitea-secrets.env
argocd --grpc-web repo add http://gitea.dclab.local/gitadmin/platform.git \
  --username "$GITEA_ADMIN_USERNAME" \
  --password "$GITEA_ADMIN_PASSWORD" \
  --upsert
```

Validate:

```bash
argocd --grpc-web repo list
```

Expected status: `Successful`.

## App-of-Apps Layout

The Gitea platform repo now contains:

```text
apps/
  root-app.yaml
  platform-apps/
    gitops-smoke.yaml
  workloads/
    gitops-smoke/
      deployment.yaml
      service.yaml
```

`root-app` watches:

```text
apps/platform-apps
```

`gitops-smoke` watches:

```text
apps/workloads/gitops-smoke
```

Only the root application was applied manually:

```bash
kubectl apply -f apps/root-app.yaml
```

Future platform apps should be added through Git under `apps/platform-apps/`.

## Validation

Argo CD health:

```bash
kubectl -n argocd get pods,svc,ingress
curl http://argocd.dclab.local
argocd --grpc-web app list
```

Expected:

```text
root-app      Synced Healthy
gitops-smoke  Synced Healthy
```

Smoke workload:

```bash
kubectl -n gitops-smoke get deploy,pod,svc
```

Expected image:

```text
harbor-01.dclab.local/platform/nginx:tls-test
```

GitOps test:

1. Commit a small change under `apps/workloads/gitops-smoke/`.
2. Push to Gitea.
3. Refresh or wait for Argo CD.
4. Confirm Argo CD applies the change without another `kubectl apply`.

The first GitOps test added label:

```text
platform.dclab.local/gitops-test=phase-3c
```

## Password Rotation

The initial admin password was rotated after validation.

The bootstrap secret was deleted:

```bash
kubectl -n argocd delete secret argocd-initial-admin-secret
```

Current password is stored only in:

```text
/root/gitea-secrets.env
```

on `utility-01`.

## Notes

- Argo CD uses HTTP internally in this phase.
- TLS for platform apps is deferred to the Vault/cert-manager phase.
- Longhorn and Gitea are not yet retroactively managed by Argo CD. That adoption requires committing their exact manifests or Helm values first to avoid drift and prune risk.
