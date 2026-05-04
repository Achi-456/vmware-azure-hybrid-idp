# Phase 3D: Gitea Actions Runner

Gitea Actions runs the on-prem CI side of the lab. A runner pod in RKE2 executes workflows from the internal Gitea `gitadmin/platform` repository, builds images with Docker-in-Docker, pushes them to Harbor, and commits deployment image tags back to Git for Argo CD to sync.

## Current State

- Namespace: `gitea-runner`
- Workload: `StatefulSet/gitea-runner`
- Replicas: `1`
- Runner version: `gitea/act_runner:0.2.11`
- Runner image: `harbor-01.dclab.local/runner/act_runner:0.2.11`
- Docker sidecar: `harbor-01.dclab.local/runner/docker:26.1-dind`
- Job image: `harbor-01.dclab.local/runner/ubuntu:act-latest`
- GitOps app: `argocd/gitea-runner`

The runner is GitOps-managed from:

```text
apps/platform-apps/gitea-runner.yaml
apps/workloads/gitea-runner/
```

## Gitea Actions Enablement

Actions were enabled on the existing Gitea deployment:

```bash
kubectl -n gitea patch deployment gitea --type=json -p='[
  {"op":"add","path":"/spec/template/spec/containers/0/env/-","value":{"name":"GITEA__actions__ENABLED","value":"true"}},
  {"op":"add","path":"/spec/template/spec/containers/0/env/-","value":{"name":"GITEA__actions__DEFAULT_ACTIONS_URL","value":"self"}}
]'
kubectl -n gitea rollout status deployment/gitea
```

Gitea keeps `Recreate` strategy because it mounts a Longhorn `ReadWriteOnce` volume.

Verify:

```bash
kubectl -n gitea exec deploy/gitea -- \
  grep -A3 '^\[actions\]' /data/gitea/conf/app.ini
```

Expected:

```text
[actions]
DEFAULT_ACTIONS_URL = self
ENABLED = true
```

## Secrets

Secrets are stored only on `utility-01`:

```text
/root/gitea-secrets.env
```

Current secret categories:

- `GITEA_RUNNER_TOKEN`
- `HARBOR_CI_USERNAME`
- `HARBOR_CI_PASSWORD`
- Gitea admin credentials for HTTP Git push

The Harbor CI credential is a project-scoped robot account for the `platform` project. Do not use the Harbor admin account in CI.

Gitea repository Actions secrets:

```text
HARBOR_REGISTRY
HARBOR_USERNAME
HARBOR_PASSWORD
PUSH_USERNAME
PUSH_PASSWORD
```

`GITEA_PUSH_USERNAME` and `GITEA_PUSH_PASSWORD` were not used because this Gitea version rejects those secret names.

## Runner Registration Token

The repository token endpoint returned `404` on this Gitea build, so the admin runner token was used:

```bash
source /root/gitea-secrets.env
RUNNER_TOKEN=$(curl -s -u "$GITEA_ADMIN_USERNAME:$GITEA_ADMIN_PASSWORD" \
  http://gitea.dclab.local/api/v1/admin/runners/registration-token \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')
```

The token was stored as `GITEA_RUNNER_TOKEN` in `/root/gitea-secrets.env` and loaded into Kubernetes secret `gitea-runner/runner-secret`.

## Runner Images

Harbor project:

```text
runner
```

Mirrored images:

```text
harbor-01.dclab.local/runner/act_runner:0.2.11
harbor-01.dclab.local/runner/docker:26.1-dind
harbor-01.dclab.local/runner/docker:26.1-cli
harbor-01.dclab.local/runner/ubuntu:act-latest
```

Sample app base images in Harbor project `platform`:

```text
harbor-01.dclab.local/platform/golang:1.22-alpine
harbor-01.dclab.local/platform/alpine:3.19
```

## CI Workflows

Workflow files in the internal platform repo:

```text
.gitea/workflows/ci-smoke.yaml
.gitea/workflows/build-sample-app.yaml
```

`ci-smoke.yaml` validates:

- runner starts a job
- Docker daemon is reachable
- Gitea health endpoint responds
- platform repo can be cloned
- app manifests are visible

`build-sample-app.yaml` validates the full loop:

1. Clone platform repo from Gitea.
2. Login to Harbor with robot credentials.
3. Build `apps/sample-app`.
4. Push `harbor-01.dclab.local/platform/sample-app:<git-sha>` and `latest`.
5. Update `apps/workloads/sample-app/deployment.yaml`.
6. Commit and push the image tag back to Gitea.
7. Argo CD syncs `sample-app` from Git.

## Important Fixes

Harbor robot usernames contain `$`, so the workflow must quote secrets safely:

```bash
HARBOR_USER='${{ secrets.HARBOR_USERNAME }}'
HARBOR_PASS='${{ secrets.HARBOR_PASSWORD }}'
printf '%s' "$HARBOR_PASS" | docker login harbor-01.dclab.local \
  -u "$HARBOR_USER" --password-stdin
```

Docker BuildKit failed private CA validation in the DinD build path, even though regular Docker pulls trusted Harbor. The workflow disables BuildKit:

```yaml
env:
  DOCKER_HOST: tcp://localhost:2375
  DOCKER_TLS_CERTDIR: ""
  DOCKER_BUILDKIT: "0"
```

The runner pod initially failed because the image has a read-only `/run` path and Kubernetes tried to mount a service account token there. The StatefulSet disables automatic service account token mounting:

```yaml
automountServiceAccountToken: false
```

## Validation

Runner:

```bash
kubectl -n gitea-runner get pods
kubectl -n gitea-runner logs gitea-runner-0 -c runner --tail=100
```

Expected:

```text
gitea-runner-0   2/2   Running
runner: rke2-runner-gitea-runner-0 ... declare successfully
```

Argo CD:

```bash
argocd --grpc-web app list
```

Expected:

```text
root-app      Synced Healthy
gitops-smoke  Synced Healthy
gitea-runner  Synced Healthy
sample-app    Synced Healthy
```

Sample app:

```bash
kubectl -n sample-app get pods,svc
kubectl -n sample-app get deploy sample-app \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

Expected image:

```text
harbor-01.dclab.local/platform/sample-app:<git-sha>
```

Gitea action logs:

```bash
kubectl -n gitea exec deploy/gitea -- sh -c \
  'find /data/gitea/actions_log/gitadmin/platform -type f -name "*.log" | sort'
```

Read a specific log:

```bash
kubectl -n gitea exec deploy/gitea -- \
  tail -n 160 /data/gitea/actions_log/gitadmin/platform/05/5.log
```

## Security Notes

- DinD runs privileged in this lab phase.
- Runner starts with one replica to avoid duplicate registration and simplify troubleshooting.
- CI uses HTTP Git access inside the lab network.
- TLS for platform service hostnames is deferred to Vault/cert-manager.
- Rootless builders, Kaniko, or BuildKit with CA trust are future hardening options.
