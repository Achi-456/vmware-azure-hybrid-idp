# Phase 3B: Gitea

Gitea runs inside RKE2 and provides the internal Git server for ArgoCD, Backstage, and future CI workflows.

## Current Endpoint

- URL: `http://gitea.dclab.local`
- Ingress VIP: `172.25.188.96`
- Namespace: `gitea`
- Admin username: `gitadmin`
- Admin password file: `/root/gitea-secrets.env` on `utility-01`

Read the password only on `utility-01`:

```bash
source /root/gitea-secrets.env
echo "$GITEA_ADMIN_PASSWORD"
```

Do not copy this value into Git-tracked files.

## Images

Images are mirrored into Harbor before deployment:

```bash
docker login https://harbor-01.dclab.local -u admin
docker pull gitea/gitea:1.22.1
docker tag gitea/gitea:1.22.1 harbor-01.dclab.local/gitea/gitea:1.22.1
docker push harbor-01.dclab.local/gitea/gitea:1.22.1

docker pull postgres:16.3
docker tag postgres:16.3 harbor-01.dclab.local/gitea/postgres:16.3
docker push harbor-01.dclab.local/gitea/postgres:16.3
```

The original candidate `bitnami/postgresql:16.3.0` was not used because Docker Hub did not return a valid manifest for that tag.

## Kubernetes Resources

Core objects:

```bash
kubectl create namespace gitea
kubectl -n gitea get pods,pvc,svc,ingress
```

Persistent volumes:

- `gitea-data`: `20Gi`, StorageClass `longhorn`
- `gitea-postgres-data`: `5Gi`, StorageClass `longhorn`

Images in use:

- `harbor-01.dclab.local/gitea/gitea:1.22.1`
- `harbor-01.dclab.local/gitea/postgres:16.3`

Gitea Deployment uses `Recreate` strategy. This is required because the single Gitea pod mounts a Longhorn `ReadWriteOnce` volume. The default Kubernetes `RollingUpdate` strategy caused a multi-attach conflict during restart.

Check the strategy:

```bash
kubectl -n gitea get deployment gitea -o jsonpath='{.spec.strategy.type}{"\n"}'
```

Expected:

```text
Recreate
```

## Validation

Health:

```bash
curl http://gitea.dclab.local/api/healthz
```

Expected:

```json
{
  "status": "pass"
}
```

Admin API:

```bash
source /root/gitea-secrets.env
curl -u "$GITEA_ADMIN_USERNAME:$GITEA_ADMIN_PASSWORD" \
  http://gitea.dclab.local/api/v1/user
```

Platform repo:

```bash
curl -u "$GITEA_ADMIN_USERNAME:$GITEA_ADMIN_PASSWORD" \
  http://gitea.dclab.local/api/v1/repos/gitadmin/platform
```

Expected HTTP status: `200`.

Image source validation:

```bash
kubectl -n gitea get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{range .spec.initContainers[*]}  init: {.image}{"\n"}{end}{range .spec.containers[*]}  container: {.image}{"\n"}{end}{end}'
```

Expected image prefix: `harbor-01.dclab.local/gitea/`.

Persistence check:

```bash
kubectl -n gitea rollout restart deployment/gitea
kubectl -n gitea rollout status deployment/gitea

source /root/gitea-secrets.env
curl -u "$GITEA_ADMIN_USERNAME:$GITEA_ADMIN_PASSWORD" \
  http://gitea.dclab.local/api/v1/repos/gitadmin/platform
```

The repo must still return HTTP `200` after restart.

## Platform Monorepo

Repository:

```text
http://gitea.dclab.local/gitadmin/platform
```

Initial structure:

```text
apps/
charts/
crossplane/
  claims/
  compositions/
docs/
```

This repository becomes the source watched by ArgoCD in Phase 3C.

## Notes

- Git over HTTP is enabled through Ingress.
- Git SSH is intentionally deferred; NGINX HTTP Ingress does not route TCP port `22` by hostname without extra stream configuration.
- Use temporary HTTP credentials or tokens for bootstrap operations. Do not persist passwords in Git remotes.
