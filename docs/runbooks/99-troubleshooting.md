# 99 - Troubleshooting Notes

## SSH Host Key Changed After Cloning

Symptom:

```text
WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED
Host key verification failed.
```

Fix from Windows:

```powershell
ssh-keygen -R <IP_ADDRESS>
ssh -o StrictHostKeyChecking=accept-new -i $env:USERPROFILE\.ssh\hybrid-cloud-idp root@<IP_ADDRESS>
```

## SSH Permission Denied

Symptom:

```text
Permission denied (publickey,gssapi-keyex,gssapi-with-mic).
```

Fix: use the explicit private key:

```powershell
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp root@<IP_ADDRESS>
```

## DNS Works On Linux But Not Windows

Linux validation:

```bash
getent ahostsv4 harbor-01.dclab.local
```

Windows workaround:

```powershell
notepad C:\Windows\System32\drivers\etc\hosts
ipconfig /flushdns
```

## Static IP Lost After Reboot

Cause: NetworkManager profile or cloud-init network config not persistent.

Fix:

```bash
nmcli con mod "System ens192" ipv4.method manual connection.autoconnect yes connection.autoconnect-priority 100
echo 'network: {config: disabled}' > /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
```

## Longhorn CSI CrashLoops Or Connectivity Errors

Firewall recovery applied on RKE2 nodes:

```bash
firewall-cmd --permanent --add-source=10.42.0.0/16 --zone=trusted
firewall-cmd --permanent --add-source=10.43.0.0/16 --zone=trusted
firewall-cmd --permanent --add-port=9500-9503/tcp
firewall-cmd --reload
```

Recycle unhealthy Longhorn pods only if needed:

```bash
kubectl -n longhorn-system delete pod <pod-name>
```

## Crossplane Provider Package 404

Symptom:

```text
MANIFEST_UNKNOWN: manifest unknown
```

Cause: invalid provider package tags.

Working tags:

```text
provider-family-azure:v1.1.0
provider-azure-resources:v1.1.0
provider-azure-dbforpostgresql:v1.1.0
```

## Azure Service Principal Creation Fails

Symptom:

```text
Insufficient privileges to complete the operation.
```

Cause: signed-in user lacks Entra app registration rights.

Required action:

- Assign `Application Developer`, or
- Enable user app registrations temporarily, or
- Have admin create `crossplane-idp-sp` and provide credential JSON.

## MetalLB Service Stuck Pending

Symptom:

```text
service can not have both metallb.universe.tf/loadBalancerIPs and svc.Spec.LoadBalancerIP
```

Fix:

```bash
kubectl -n kube-system annotate svc rke2-ingress-nginx-controller-lb metallb.universe.tf/loadBalancerIPs- || true
```

## MetalLB Pulls FRR From Quay In L2 Mode

Cause: MetalLB chart enables FRR sidecar by default.

Fix:

```bash
helm upgrade metallb metallb/metallb \
  --namespace metallb-system \
  --version 0.14.5 \
  --set controller.image.repository=harbor-01.dclab.local/metallb/controller \
  --set controller.image.tag=v0.14.5 \
  --set speaker.image.repository=harbor-01.dclab.local/metallb/speaker \
  --set speaker.image.tag=v0.14.5 \
  --set speaker.frr.enabled=false \
  --wait \
  --timeout 5m
```

## Longhorn UI Not Reachable From Windows

Direct NodePort may time out from Windows.

Use tunnel:

```powershell
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp -L 30080:172.25.188.90:30080 root@172.25.188.94
```

Open:

```text
http://localhost:30080
```

## Harbor Password Location

Harbor admin password is not in Git.

Retrieve when needed:

```powershell
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp root@172.25.188.93 "cat /root/.harbor-admin-password"
```

## Gitea Runner Pod Fails With Read-Only `/run`

Symptom:

```text
RunContainerError
read-only file system
```

Cause: Kubernetes tried to mount the default service account token under `/var/run/secrets`, but the runner image has a read-only `/run`.

Fix in the runner StatefulSet:

```yaml
spec:
  template:
    spec:
      automountServiceAccountToken: false
```

After Argo CD applies the change, recycle the failed pod:

```bash
kubectl -n gitea-runner delete pod gitea-runner-0
```

## Harbor Robot Username Breaks Docker Login

Symptom:

```text
docker login harbor-01.dclab.local
unauthorized
```

Cause: Harbor robot usernames contain `$`, and shell expansion can alter the username when a workflow uses double quotes.

Fix in Gitea workflow:

```bash
HARBOR_USER='${{ secrets.HARBOR_USERNAME }}'
HARBOR_PASS='${{ secrets.HARBOR_PASSWORD }}'
printf '%s' "$HARBOR_PASS" | docker login harbor-01.dclab.local \
  -u "$HARBOR_USER" --password-stdin
```

## Docker BuildKit Fails Harbor Private CA Validation

Symptom:

```text
failed to fetch oauth token
x509: certificate signed by unknown authority
```

Context: regular `docker pull` from the DinD daemon works, but `docker build` using BuildKit fails while resolving Harbor base image metadata.

Fix for the lab workflow:

```yaml
env:
  DOCKER_HOST: tcp://localhost:2375
  DOCKER_TLS_CERTDIR: ""
  DOCKER_BUILDKIT: "0"
```

Future hardening should restore BuildKit with explicit CA trust or move to a rootless builder.
