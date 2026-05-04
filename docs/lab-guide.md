# Hybrid Cloud IDP Lab Guide

## Architecture

This lab models a hybrid internal developer platform:

- VMware vSphere for self-managed Kubernetes nodes
- RKE2 for Kubernetes control plane and workers
- Harbor for enterprise registry workflows
- Azure for managed service integration in later phases

## Host Inventory

| Hostname | IP | Role |
| --- | --- | --- |
| `rke2-api.dclab.local` | `172.25.188.90` | Stable RKE2 API endpoint |
| `rke2-cp-01.dclab.local` | `172.25.188.90` | RKE2 control plane |
| `rke2-worker-01.dclab.local` | `172.25.188.91` | RKE2 worker |
| `rke2-worker-02.dclab.local` | `172.25.188.92` | RKE2 worker |
| `harbor-01.dclab.local` | `172.25.188.93` | Harbor node |
| `utility-01.dclab.local` | `172.25.188.94` | Admin/bastion node |
| `edge-01.dclab.local` | `172.25.188.95` | Reserved edge VM |
| `gitea.dclab.local` | `172.25.188.96` | Platform ingress VIP |
| `argocd.dclab.local` | `172.25.188.96` | Platform ingress VIP |
| `longhorn.dclab.local` | `172.25.188.96` | Platform ingress VIP |
| `grafana.dclab.local` | `172.25.188.96` | Platform ingress VIP |
| `backstage.dclab.local` | `172.25.188.96` | Platform ingress VIP |

Network baseline:

- Subnet prefix: `/16`
- Gateway: `172.25.188.1`
- DNS: `172.25.188.20`
- DNS search: `dclab.local`

## VM Sizing And Storage

| VM | vCPU | RAM | Disk |
| --- | ---: | ---: | --- |
| `rke2-cp-01` | 2 | ~7.5 GB | `sda` 100 GB |
| `rke2-worker-01` | 2 | ~7.5 GB | `sda` 100 GB |
| `rke2-worker-02` | 2 | ~7.5 GB | `sda` 100 GB |
| `harbor-01` | 2 | ~7.5 GB | `sda` 80 GB + `sdb` 150 GB (`/data`) |
| `utility-01` | 2 | ~5.5 GB | `sda` 80 GB |

Common filesystem layout:

- `/` ~48 GB
- `/home` ~24 GB
- `/boot` ~1 GB

Harbor data mount:

- `/data` on `/dev/sdb` (XFS, 150 GB)

## Harbor Endpoint Policy

Use `harbor-01.dclab.local` for all Linux node and cluster-side configuration.

- Harbor hostname in `harbor.yml`: `harbor-01.dclab.local`
- RKE2 mirror key: `harbor-01.dclab.local`
- RKE2 mirror endpoint: `https://harbor-01.dclab.local`

Windows workstation DNS for `dclab.local` is still inconsistent, so workstation-side tests can use `172.25.188.93` as fallback.

## Harbor Runtime And Registry Integration

Deployed Harbor version:

- `v2.15.0` (offline installer)

Harbor host runtime:

- Docker Engine (`docker-ce`) + Docker Compose plugin
- Harbor compose location: `/data/harbor/harbor/docker-compose.yml`
- Harbor config: `/data/harbor/harbor/harbor.yml`
- `harbor_admin_password` was rotated from default.

Credential storage policy:

- Harbor admin password file: `/root/.harbor-admin-password` on `harbor-01` only
- RKE2 pull service credential file: `/root/.harbor-rke2pull-password` on `harbor-01` only
- Do not copy these values into Git-tracked docs or manifests.

RKE2 registry integration:

- File on all RKE2 nodes: `/etc/rancher/rke2/registries.yaml`
- Mirror endpoint: `https://harbor-01.dclab.local`
- Auth user for pulls: `rke2pull` (project-scoped service user in Harbor project `platform`)

Current security posture:

- Harbor registry is HTTPS with private CA trust distributed to all Linux nodes.
- Windows workstation may still require IP fallback for browser checks where local DNS is inconsistent.

## Phase 3A Ingress Edge

MetalLB exposes the existing RKE2 bundled NGINX ingress controller with one stable L2 VIP.

MetalLB:

- Helm chart: `metallb/metallb` version `0.14.5`
- Namespace: `metallb-system`
- Mode: Layer 2
- Interface: `ens192`
- IPAddressPool: `lab-ingress-pool`
- Address: `172.25.188.96/32`
- L2Advertisement: `lab-l2`
- Images:
  - `harbor-01.dclab.local/metallb/controller:v0.14.5`
  - `harbor-01.dclab.local/metallb/speaker:v0.14.5`
- FRR sidecar disabled for this L2-only lab path.

Ingress:

- Reuses RKE2 bundled DaemonSet `kube-system/rke2-ingress-nginx-controller`.
- IngressClass: `nginx`
- LoadBalancer service: `kube-system/rke2-ingress-nginx-controller-lb`
- External IP: `172.25.188.96`
- Ports: `80` and `443`

DNS policy:

- Use explicit records for now: `gitea`, `argocd`, `grafana`, and `backstage` all point to `172.25.188.96`.
- Add wildcard `*.dclab.local -> 172.25.188.96` later if DNS administration supports it.
- `edge-01.dclab.local` remains a separate VM on `172.25.188.95`.

Validation:

```bash
kubectl -n metallb-system get pods
kubectl -n metallb-system get ipaddresspool,l2advertisement
kubectl -n kube-system get svc rke2-ingress-nginx-controller-lb
curl --resolve "echo.dclab.local:80:172.25.188.96" http://echo.dclab.local
```

Notes:

- The VIP may not respond to ICMP ping because Kubernetes `LoadBalancer` services forward TCP/UDP service ports, not arbitrary ICMP.
- The bundled RKE2 ingress controller may not populate `ADDRESS` on each Ingress object unless publish-service behavior is configured. HTTP routing through the VIP is the authoritative validation.

## Phase 3B Internal Git: Gitea

Gitea is deployed on RKE2 as the internal Git server for ArgoCD, Backstage, and future CI workflows.

Endpoint:

- Browser URL: `https://gitea.dclab.local`
- Internal service/root URL: `http://gitea.dclab.local`
- Ingress VIP: `172.25.188.96`
- Namespace: `gitea`

Images:

- `harbor-01.dclab.local/gitea/gitea:1.22.1`
- `harbor-01.dclab.local/gitea/postgres:16.3`

Storage:

- `gitea-data`: `20Gi`, Longhorn
- `gitea-postgres-data`: `5Gi`, Longhorn

Runtime choices:

- Gitea uses PostgreSQL, not SQLite.
- Gitea Deployment strategy is `Recreate` because the pod mounts a single Longhorn `ReadWriteOnce` volume.
- Git over HTTP is enabled through Ingress.
- Git over SSH is deferred until a later TCP routing design.

Credentials:

- Admin username: `gitadmin`
- Admin and database passwords are stored only on `utility-01` in `/root/gitea-secrets.env`.
- Do not commit these values to Git.

Platform monorepo:

- Browser URL: `https://gitea.dclab.local/gitadmin/platform`
- Git URL for Argo CD: `https://gitea.dclab.local/gitadmin/platform.git`
- Purpose: source of truth for ArgoCD apps, Helm charts, Crossplane compositions, and platform docs.

Validation:

```bash
kubectl -n gitea get pods,pvc,svc,ingress
curl -k https://gitea.dclab.local/api/healthz
source /root/gitea-secrets.env
curl -u "$GITEA_ADMIN_USERNAME:$GITEA_ADMIN_PASSWORD" \
  -k https://gitea.dclab.local/api/v1/repos/gitadmin/platform
```

## Phase 3C GitOps Control Plane: Argo CD

Argo CD is deployed on RKE2 as the GitOps engine for platform services.

Endpoint:

- URL: `https://argocd.dclab.local`
- Ingress VIP: `172.25.188.96`
- Namespace: `argocd`

Versions:

- Helm chart: `argo/argo-cd` version `6.8.1`
- Argo CD application image: `v2.11.3`

Images:

- `harbor-01.dclab.local/argocd/argocd:v2.11.3`
- `harbor-01.dclab.local/argocd/redis:7.0.15-alpine`

Runtime choices:

- Dex is disabled for this phase.
- Argo CD server runs with `server.insecure=true`.
- Ingress terminates TLS through the existing RKE2 NGINX controller.
- SSO is deferred to the identity phase.

Credentials:

- Admin username: `admin`
- Admin password is stored only on `utility-01` in `/root/gitea-secrets.env`.
- The initial Argo CD admin secret was deleted after password rotation.

Gitea repo connection:

- Repo URL: `https://gitea.dclab.local/gitadmin/platform.git`
- Status: `Successful`

App-of-Apps:

- Root app: `apps/root-app.yaml`
- Root path watched by Argo CD: `apps/platform-apps`
- First child app: `gitops-smoke`
- Smoke workload path: `apps/workloads/gitops-smoke`

Current Argo CD apps:

```text
root-app      Synced Healthy
gitops-smoke  Synced Healthy
gitea-runner  Synced Healthy
sample-app    Synced Healthy
```

Validation:

```bash
kubectl -n argocd get pods,svc,ingress
curl -k https://argocd.dclab.local
argocd --grpc-web repo list
argocd --grpc-web app list
kubectl -n gitops-smoke get deploy,pod,svc
```

Notes:

- `gitops-smoke` uses `harbor-01.dclab.local/platform/nginx:tls-test`.
- The first GitOps sync test added label `platform.dclab.local/gitops-test=phase-3c` from Git and Argo CD applied it automatically.
- Longhorn and Gitea are not yet retroactively managed by Argo CD; that adoption is deferred until their exact manifests or Helm values are committed.

## Phase 3D On-Prem CI: Gitea Actions Runner

Gitea Actions is enabled and a runner is deployed inside RKE2.

Endpoint:

- Actions UI: `http://gitea.dclab.local/gitadmin/platform/actions`
- Runner namespace: `gitea-runner`
- Runner workload: `StatefulSet/gitea-runner`

Images:

- `harbor-01.dclab.local/runner/act_runner:0.2.11`
- `harbor-01.dclab.local/runner/docker:26.1-dind`
- `harbor-01.dclab.local/runner/docker:26.1-cli`
- `harbor-01.dclab.local/runner/ubuntu:act-latest`

GitOps paths:

```text
apps/platform-apps/gitea-runner.yaml
apps/workloads/gitea-runner/
```

Workflow files:

```text
.gitea/workflows/ci-smoke.yaml
.gitea/workflows/build-sample-app.yaml
```

Current CI loop:

```text
Push to Gitea
  -> Gitea Actions runner starts workflow
  -> Docker-in-Docker builds sample image
  -> image is pushed to Harbor project platform
  -> workflow commits updated image tag to Git
  -> Argo CD syncs sample-app from Git
```

Sample app:

- Namespace: `sample-app`
- Image pattern: `harbor-01.dclab.local/platform/sample-app:<git-sha>`
- Managed by Argo CD app: `sample-app`

Credential policy:

- Runner token and Harbor robot credentials are stored only on `utility-01` in `/root/gitea-secrets.env`.
- Kubernetes runner token is stored in secret `gitea-runner/runner-secret`.
- Gitea repo Actions secrets hold only the values needed by workflows.
- No secret values are committed to Git.

Implementation notes:

- Runner uses a `StatefulSet` so registration data survives restarts.
- Runner starts with one replica.
- DinD is privileged in this lab phase.
- `automountServiceAccountToken: false` is required because the runner image has a read-only `/run`.
- `DOCKER_BUILDKIT=0` is set in the sample app workflow due Harbor private CA handling in the BuildKit path.

Validation:

```bash
kubectl -n gitea-runner get pods
kubectl -n gitea-runner logs gitea-runner-0 -c runner --tail=100
argocd --grpc-web app list
kubectl -n sample-app get pods,svc
kubectl -n sample-app get deploy sample-app \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

## Phase 4A Internal TLS: cert-manager

cert-manager is installed through Argo CD and issues certificates from a lab self-signed root CA.

Endpoint changes:

- Gitea browser URL: `https://gitea.dclab.local`
- Argo CD browser URL: `https://argocd.dclab.local`
- Longhorn browser URL: `https://longhorn.dclab.local`
- Harbor remains: `https://harbor-01.dclab.local`

Versions:

- cert-manager chart: `v1.20.2`
- cert-manager images: `v1.20.2`

Images:

- `harbor-01.dclab.local/cert-manager/cert-manager-controller:v1.20.2`
- `harbor-01.dclab.local/cert-manager/cert-manager-webhook:v1.20.2`
- `harbor-01.dclab.local/cert-manager/cert-manager-cainjector:v1.20.2`
- `harbor-01.dclab.local/cert-manager/cert-manager-startupapicheck:v1.20.2`

Issuer model:

- Bootstrap issuer: `ClusterIssuer/dclab-selfsigned-bootstrap`
- Root CA certificate: `Certificate/cert-manager/dclab-root-ca`
- Platform issuer: `ClusterIssuer/dclab-ca`

GitOps apps:

```text
cert-manager
cert-manager-issuers
platform-tls
```

Argo CD now uses the HTTPS Gitea repo URL:

```text
https://gitea.dclab.local/gitadmin/platform.git
```

Certificate checks:

```bash
kubectl get clusterissuer
kubectl -n cert-manager get certificate dclab-root-ca
kubectl -n gitea get certificate gitea-tls
kubectl -n argocd get certificate argocd-tls
kubectl -n longhorn-system get certificate longhorn-tls
```

Harbor exception:

- Harbor is still standalone on `harbor-01`.
- Harbor TLS is not managed by cert-manager in Phase 4A.
- Vault-backed PKI and possible Harbor CA unification are deferred to Phase 4B.

## Crossplane Phase 2D Status

Crossplane control plane:

- Helm chart: `crossplane-stable/crossplane` version `1.15.2`
- Namespace: `crossplane-system`
- Core pods: `crossplane`, `crossplane-rbac-manager` running

Azure provider packages pinned to valid OCI tags:

- `xpkg.upbound.io/upbound/provider-family-azure:v1.1.0`
- `xpkg.upbound.io/upbound/provider-azure-resources:v1.1.0`
- `xpkg.upbound.io/upbound/provider-azure-dbforpostgresql:v1.1.0`

Provider health check:

```bash
kubectl get providers
```

Current execution blocker:

- Azure login works on `utility-01`, but service principal creation failed with Entra permission error:
  - `Insufficient privileges to complete the operation`
- Until a service principal credential JSON is provided, `ProviderConfig`, `ResourceGroup`, and PostgreSQL smoke provisioning cannot be completed.

Credential handling for Phase 2D:

- Keep Azure credential JSON only on `utility-01` (for example `/root/crossplane-azure-creds.json`).
- Create `azure-secret` in `crossplane-system` from that local file.
- Never commit secrets or credential content to Git-tracked files.

Restart sequence (safe order):

1. Restart worker agents one by one: `rke2-worker-01`, then `rke2-worker-02`.
2. Verify each worker returns `Ready`.
3. Restart `rke2-server` on `rke2-cp-01`.
4. Verify all nodes are `Ready`.

Rollback steps:

1. Restore previous `/etc/rancher/rke2/registries.yaml` on affected node(s).
2. Restart node service (`rke2-agent` on workers, `rke2-server` on control plane).
3. Validate with `kubectl get nodes` and pod image-pull events.

## RHEL 8.8 Base And Repository

Template source:

- `tpl-rhel-8.8-cloudinit`

Local ISO repo on all VMs:

- Mount: `/mnt/rhel-iso`
- Repo file: `/etc/yum.repos.d/rhel8-iso.repo`
- Repos:
  - `rhel8-iso-baseos`
  - `rhel8-iso-appstream`
- `fstab` entry:
  - `/dev/sr0 /mnt/rhel-iso iso9660 ro,nofail 0 0`

## Persistent Networking Standard

All nodes use NetworkManager profile `System ens192` with static settings and cloud-init network override disabled.

Required profile settings:

- `ipv4.method manual`
- `ipv4.addresses` set per host
- `ipv4.gateway=172.25.188.1`
- `ipv4.dns=172.25.188.20`
- `ipv4.dns-search=dclab.local`
- `ipv4.ignore-auto-dns=yes`
- `ipv4.ignore-auto-routes=yes`
- `ipv4.may-fail=no`
- `connection.autoconnect=yes`
- `connection.autoconnect-priority=100`

Cloud-init override:

- `/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg`
- Content: `network: {config: disabled}`

## Access Pattern

Primary SSH key:

```bash
~/.ssh/hybrid-cloud-idp
```

Example connection:

```bash
ssh -i ~/.ssh/hybrid-cloud-idp root@172.25.188.90
```

Do not store credentials in this repository.

## Workflow Summary

1. Build and clean RHEL 8.8 template.
2. Clone five VMs from template.
3. Apply static network and hostname for each VM.
4. Validate DNS, egress, time sync.
5. Install and validate RKE2 (see [RKE2 Install Notes](C:/Users/achinthah/Documents/hybrid-cloud-idp/docs/rke2-install.md)).
6. Continue with Harbor, Crossplane, and Azure integration.
