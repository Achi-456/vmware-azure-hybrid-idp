# Phase 4A: cert-manager Internal TLS

Phase 4A installs cert-manager through Argo CD and uses a lab CA to issue TLS certificates for Kubernetes-hosted platform UIs.

## Scope

TLS is enabled for:

- `https://gitea.dclab.local`
- `https://argocd.dclab.local`
- `https://longhorn.dclab.local`

Harbor is not changed in this phase. It remains a standalone Docker Compose service on `harbor-01` with its existing HTTPS certificate and private CA.

## Versions

| Component | Version |
| --- | --- |
| cert-manager Helm chart | `v1.20.2` |
| cert-manager controller | `v1.20.2` |
| cert-manager webhook | `v1.20.2` |
| cert-manager cainjector | `v1.20.2` |
| cert-manager startupapicheck | `v1.20.2` |

## Harbor Images

Mirrored into Harbor project `cert-manager`:

```text
harbor-01.dclab.local/cert-manager/cert-manager-controller:v1.20.2
harbor-01.dclab.local/cert-manager/cert-manager-webhook:v1.20.2
harbor-01.dclab.local/cert-manager/cert-manager-cainjector:v1.20.2
harbor-01.dclab.local/cert-manager/cert-manager-startupapicheck:v1.20.2
```

## GitOps Paths

Internal Gitea platform repo paths:

```text
apps/platform-apps/cert-manager.yaml
apps/platform-apps/cert-manager-issuers.yaml
apps/platform-apps/platform-tls.yaml
apps/workloads/cert-manager-issuers/issuers.yaml
apps/workloads/platform-tls/
```

Argo CD applications:

```bash
argocd --grpc-web app get cert-manager
argocd --grpc-web app get cert-manager-issuers
argocd --grpc-web app get platform-tls
```

## Issuer Model

Phase 4A uses a self-signed lab root CA:

```text
ClusterIssuer/dclab-selfsigned-bootstrap
  -> Certificate/cert-manager/dclab-root-ca
  -> Secret/cert-manager/dclab-root-ca
  -> ClusterIssuer/dclab-ca
```

The root CA private key stays inside Kubernetes secret `cert-manager/dclab-root-ca`. Export only the public certificate for workstation trust.

## Validation

```bash
kubectl -n cert-manager get pods
kubectl get clusterissuer
kubectl -n cert-manager get certificate dclab-root-ca
kubectl -n gitea get certificate gitea-tls
kubectl -n argocd get certificate argocd-tls
kubectl -n longhorn-system get certificate longhorn-tls
```

Expected certificate state:

```text
dclab-root-ca   True
gitea-tls       True
argocd-tls      True
longhorn-tls    True
```

HTTPS checks:

```bash
curl -k https://gitea.dclab.local
curl -k https://argocd.dclab.local
curl -k https://longhorn.dclab.local
curl -sk -o /dev/null -w "%{http_code}\n" https://harbor-01.dclab.local
```

Expected: all return HTTP `200`.

## Argo CD Repository Note

After enabling HTTPS for Gitea, Argo CD was moved to the HTTPS platform repo URL:

```text
https://gitea.dclab.local/gitadmin/platform.git
```

The old HTTP repo entry was removed from Argo CD because Gitea can redirect HTTP Git requests after HTTPS exposure.

## Workstation CA Trust

Export the public root CA from utility-01:

```bash
kubectl -n cert-manager get secret dclab-root-ca \
  -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/dclab-root-ca.crt
```

Copy only `dclab-root-ca.crt` to Windows and import it into:

```text
Trusted Root Certification Authorities
```

Do not export or copy the CA private key.

## Rollback

If TLS routing breaks a platform UI:

1. Revert the latest platform repo commit that changed `apps/workloads/platform-tls/`.
2. Let Argo CD sync `platform-tls`.
3. If needed, temporarily access Longhorn through the existing NodePort tunnel:

```powershell
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp -L 30080:172.25.188.90:30080 root@172.25.188.94
```

4. Harbor is unaffected by this rollback because it is outside Kubernetes.
