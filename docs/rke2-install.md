# RKE2 Install Notes (Week 1)

## Cluster Topology

- `rke2-cp-01.dclab.local` (`172.25.188.90`) - server, control-plane, etcd
- `rke2-worker-01.dclab.local` (`172.25.188.91`) - agent
- `rke2-worker-02.dclab.local` (`172.25.188.92`) - agent
- `utility-01.dclab.local` (`172.25.188.94`) - kubectl admin host

## Important RHEL 8.8 Requirement

RKE2 `v1.35.x` kubelet requires cgroup v2.

Set this on all RKE2 nodes and reboot before starting services:

```bash
grubby --update-kernel=ALL --args="systemd.unified_cgroup_hierarchy=1"
reboot
```

Validation:

```bash
stat -fc %T /sys/fs/cgroup
```

Expected:

```text
cgroup2fs
```

## Node Preparation

- Installed: `curl vim-enhanced git tar iscsi-initiator-utils nfs-utils iproute-tc chrony`
- Enabled: `iscsid`, `chronyd`
- Disabled swap (`swapoff -a` and fstab comment)
- Set modules: `br_netfilter`, `overlay`
- Set sysctl:
  - `net.bridge.bridge-nf-call-iptables=1`
  - `net.bridge.bridge-nf-call-ip6tables=1`
  - `net.ipv4.ip_forward=1`
- Added NetworkManager unmanaged CNI interfaces via `/etc/NetworkManager/conf.d/rke2-cni.conf`

## Firewall Ports Opened

All RKE2 nodes:

- `10250/tcp`
- `8472/udp`
- `9099/tcp`
- `30000-32767/tcp`

Control-plane node additional:

- `6443/tcp`
- `9345/tcp`
- `2379-2381/tcp`

## RKE2 Configuration

Control-plane `/etc/rancher/rke2/config.yaml`:

```yaml
node-name: rke2-cp-01
tls-san:
  - rke2-api.dclab.local
  - rke2-cp-01.dclab.local
write-kubeconfig-mode: "0644"
```

Workers `/etc/rancher/rke2/config.yaml`:

```yaml
server: https://rke2-api.dclab.local:9345
token: <control-plane node-token>
node-name: rke2-worker-01
```

## Utility Host kubectl

- Copied kubeconfig from control plane to `/root/.kube/config`
- Updated server endpoint to `https://rke2-api.dclab.local:6443`
- Copied matching `kubectl` binary from control plane to `/usr/local/bin/kubectl`

## Validation Snapshot

All nodes:

- `rke2-cp-01` - `Ready`
- `rke2-worker-01` - `Ready`
- `rke2-worker-02` - `Ready`

Demo workload:

- Namespace: `demo`
- Deployment: `nginx`
- Service: `ClusterIP` on port `80`
- Pod reached `Running`

