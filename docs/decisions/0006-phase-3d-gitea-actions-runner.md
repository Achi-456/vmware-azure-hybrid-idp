# 0006: Phase 3D Gitea Actions Runner

## Status

Accepted

## Context

The lab needs an on-prem CI loop before adding more platform services. Gitea is already the internal Git server and Argo CD is already syncing applications from the `gitadmin/platform` repository.

The desired flow is:

```text
Git push to Gitea
  -> Gitea Actions runner starts a job
  -> job builds an image with Docker-in-Docker
  -> image is pushed to Harbor
  -> workflow commits the new image tag to Git
  -> Argo CD syncs the workload
```

## Decision

Deploy one Gitea Actions runner as a GitOps-managed `StatefulSet` in namespace `gitea-runner`.

Use:

- `gitea/act_runner:0.2.11`
- privileged Docker-in-Docker sidecar `docker:26.1-dind`
- Longhorn-backed runner registration data
- Harbor-hosted runner and build images
- Harbor project robot credentials for CI pushes

Keep secrets outside Git:

- runner token in Kubernetes secret `gitea-runner/runner-secret`
- Gitea, Harbor, and runner secrets in `/root/gitea-secrets.env` on `utility-01`
- repo Actions secrets in Gitea

Use the internal Gitea platform repo for runner manifests:

```text
apps/platform-apps/gitea-runner.yaml
apps/workloads/gitea-runner/
```

## Corrections Made During Implementation

The repository runner token endpoint returned `404` on the installed Gitea build, so the admin runner token endpoint was used.

Gitea rejected repo secret names prefixed with `GITEA_`, so the workflow uses `PUSH_USERNAME` and `PUSH_PASSWORD`.

Harbor robot usernames contain `$`, which shell expands unless handled carefully. CI scripts now assign these values through single-quoted literals and use `docker login --password-stdin`.

Docker BuildKit failed to validate the Harbor private CA while resolving base image metadata. The sample app workflow disables BuildKit for this phase with `DOCKER_BUILDKIT=0`.

The runner image uses a read-only `/run`; Kubernetes service account token auto-mounting caused container start failure. The StatefulSet sets `automountServiceAccountToken: false`.

## Consequences

Benefits:

- The lab now has a complete on-prem CI/CD loop.
- CI no longer needs GitHub Actions or external SaaS runners.
- Argo CD remains the deployment engine; CI only updates Git.
- Harbor remains the image source and scan point.

Tradeoffs:

- DinD requires privileged containers.
- BuildKit is disabled for the sample app workflow until CA handling is hardened.
- Runner scale is intentionally one replica for easier troubleshooting.

## Follow-Up

- Replace privileged DinD with a rootless builder or Kaniko.
- Add TLS for `gitea.dclab.local` and `argocd.dclab.local`.
- Add branch protection and required CI checks in Gitea.
- Scale runners after stable token and registration behavior is proven.
