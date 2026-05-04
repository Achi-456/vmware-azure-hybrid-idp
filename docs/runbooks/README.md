# Runbook Index

This folder captures the command path used to build the lab so far.

Secrets are intentionally excluded. When a command needs a credential, read it from the secure node-local file documented in `docs/lab-guide.md`.

## Files

- `00-access-and-inventory.md` - VM inventory, SSH, DNS, hosts-file fallback, and common checks.
- `01-rhel-template-and-clones.md` - RHEL template prep, local ISO repos, cloning cleanup, and edge clone cleanup.
- `02-rke2-cluster.md` - RKE2 prep, install, kubeconfig, and cluster validation.
- `03-harbor-registry.md` - Harbor install, TLS, Harbor projects, and registry integration.
- `04-longhorn-storage.md` - Longhorn install, Harbor image path, and PVC validation.
- `05-crossplane-azure-paused.md` - Crossplane install, provider version fix, and Azure credential blocker.
- `06-metallb-ingress.md` - MetalLB L2 VIP and RKE2 ingress LoadBalancer exposure.
- `07-gitea.md` - Gitea deployment on RKE2 with Longhorn, Harbor images, and ingress routing.
- `08-argocd.md` - Argo CD install, Gitea repo connection, and App-of-Apps bootstrap.
- `09-gitea-actions-runner.md` - Gitea Actions runner, Harbor image builds, and the on-prem CI loop.
- `99-troubleshooting.md` - Issues encountered and exact recovery steps.
