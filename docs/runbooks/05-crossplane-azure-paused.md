# 05 - Crossplane And Azure Provider

Azure provisioning is paused because the signed-in Azure user cannot create Entra app registrations/service principals.

## Crossplane Install

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

helm upgrade --install crossplane crossplane-stable/crossplane \
  --namespace crossplane-system \
  --create-namespace \
  --version 1.15.2 \
  --wait \
  --timeout 10m
```

Validate:

```bash
kubectl -n crossplane-system get pods
kubectl get crds | grep crossplane
```

## Azure Providers

Invalid versions tried:

```text
v1.12.1
v1.11.4
```

Those failed with `MANIFEST_UNKNOWN`.

Working provider versions:

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: upbound-provider-family-azure
spec:
  package: xpkg.upbound.io/upbound/provider-family-azure:v1.1.0
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: upbound-provider-azure-resources
spec:
  package: xpkg.upbound.io/upbound/provider-azure-resources:v1.1.0
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: upbound-provider-azure-dbforpostgresql
spec:
  package: xpkg.upbound.io/upbound/provider-azure-dbforpostgresql:v1.1.0
EOF
```

Validate:

```bash
kubectl get providers
kubectl -n crossplane-system get pods
```

## Azure CLI

Install on `utility-01`:

```bash
rpm --import https://packages.microsoft.com/keys/microsoft.asc

cat > /etc/yum.repos.d/azure-cli.repo <<'EOF'
[azure-cli]
name=Azure CLI
baseurl=https://packages.microsoft.com/yumrepos/azure-cli
enabled=1
gpgcheck=1
gpgkey=https://packages.microsoft.com/keys/microsoft.asc
EOF

dnf install -y azure-cli
az version
```

Login:

```bash
az login --use-device-code
az account show -o table
```

## Blocked SP Creation

Attempted:

```bash
SUB_ID=$(az account show --query id -o tsv)

az ad sp create-for-rbac \
  --name crossplane-idp-sp \
  --role Contributor \
  --scopes /subscriptions/${SUB_ID} \
  --sdk-auth > ~/crossplane-azure-creds.json
```

Observed error:

```text
Insufficient privileges to complete the operation.
```

Required admin action:

- Assign user `Application Developer`, or
- Enable application registration for users temporarily, or
- Create service principal `crossplane-idp-sp` and provide credential JSON.

## Commands To Resume Later

Place credential JSON on `utility-01`:

```bash
chmod 600 /root/crossplane-azure-creds.json
```

Create secret:

```bash
kubectl -n crossplane-system delete secret azure-secret --ignore-not-found
kubectl -n crossplane-system create secret generic azure-secret \
  --from-file=creds=/root/crossplane-azure-creds.json
```

ProviderConfig:

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: azure.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: azure-secret
      key: creds
EOF
```
