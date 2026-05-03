# 00 - Access And Inventory

## VM Inventory

| Host | IP | Role |
| --- | --- | --- |
| `rke2-cp-01.dclab.local` | `172.25.188.90` | RKE2 control plane |
| `rke2-worker-01.dclab.local` | `172.25.188.91` | RKE2 worker |
| `rke2-worker-02.dclab.local` | `172.25.188.92` | RKE2 worker |
| `harbor-01.dclab.local` | `172.25.188.93` | Harbor registry |
| `utility-01.dclab.local` | `172.25.188.94` | Admin workstation |
| `edge-01.dclab.local` | `172.25.188.95` | Reserved edge VM |
| platform ingress VIP | `172.25.188.96` | MetalLB/RKE2 ingress |

## SSH From Windows

```powershell
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp root@172.25.188.94
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp root@172.25.188.90
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp root@172.25.188.93
```

Remove stale host keys after cloning:

```powershell
ssh-keygen -R 172.25.188.90
ssh-keygen -R 172.25.188.91
ssh-keygen -R 172.25.188.92
ssh-keygen -R 172.25.188.93
ssh-keygen -R 172.25.188.94
ssh-keygen -R 172.25.188.95
```

## Linux DNS Checks

```bash
getent ahostsv4 rke2-api.dclab.local
getent ahostsv4 harbor-01.dclab.local
getent ahostsv4 gitea.dclab.local
getent ahostsv4 argocd.dclab.local
getent ahostsv4 grafana.dclab.local
getent ahostsv4 backstage.dclab.local
```

## Hosts File Fallback

Use only when DNS is unreliable:

```bash
cat <<'EOF' >> /etc/hosts
172.25.188.90 rke2-api.dclab.local rke2-api
172.25.188.90 rke2-cp-01.dclab.local rke2-cp-01
172.25.188.91 rke2-worker-01.dclab.local rke2-worker-01
172.25.188.92 rke2-worker-02.dclab.local rke2-worker-02
172.25.188.93 harbor-01.dclab.local harbor-01
172.25.188.94 utility-01.dclab.local utility-01
172.25.188.95 edge-01.dclab.local edge-01
172.25.188.96 gitea.dclab.local argocd.dclab.local grafana.dclab.local backstage.dclab.local
EOF
```

Ensure resolver order prefers files:

```bash
grep '^hosts:' /etc/nsswitch.conf
# Expected: hosts: files dns myhostname
```

## Windows Hosts File Fallback

Run Notepad as Administrator:

```powershell
notepad C:\Windows\System32\drivers\etc\hosts
```

Add:

```text
172.25.188.93 harbor-01.dclab.local harbor-01
172.25.188.96 gitea.dclab.local argocd.dclab.local grafana.dclab.local backstage.dclab.local
172.25.188.90 rke2-api.dclab.local rke2-cp-01.dclab.local
172.25.188.91 rke2-worker-01.dclab.local
172.25.188.92 rke2-worker-02.dclab.local
172.25.188.94 utility-01.dclab.local
172.25.188.95 edge-01.dclab.local
```

Flush DNS:

```powershell
ipconfig /flushdns
```

## Common Health Checks

```bash
hostnamectl --static
ip -4 addr show ens192
timedatectl
free -h
df -h /
systemctl is-active sshd
```

From `utility-01`:

```bash
export KUBECONFIG=/root/.kube/config
kubectl get nodes -o wide
kubectl get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded
```
