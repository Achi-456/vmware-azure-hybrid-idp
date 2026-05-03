# 02 - RKE2 Cluster

Run RKE2 operations from `utility-01` unless noted.

## RHEL 8.8 Cgroup v2 Requirement

Run on all RKE2 nodes, then reboot:

```bash
grubby --update-kernel=ALL --args="systemd.unified_cgroup_hierarchy=1"
reboot
```

Validate:

```bash
stat -fc %T /sys/fs/cgroup
# Expected: cgroup2fs
```

## Node Preparation

Run on `rke2-cp-01`, `rke2-worker-01`, and `rke2-worker-02`:

```bash
dnf install -y curl vim-enhanced git tar iscsi-initiator-utils nfs-utils iproute-tc chrony
systemctl enable --now iscsid chronyd
swapoff -a
sed -i.bak '/ swap / s/^/#/' /etc/fstab
```

Kernel modules:

```bash
cat > /etc/modules-load.d/rke2.conf <<'EOF'
br_netfilter
overlay
EOF

modprobe br_netfilter
modprobe overlay
```

Sysctl:

```bash
cat > /etc/sysctl.d/90-rke2.conf <<'EOF'
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
net.ipv4.ip_forward=1
EOF

sysctl --system
```

NetworkManager CNI ignore:

```bash
cat > /etc/NetworkManager/conf.d/rke2-cni.conf <<'EOF'
[keyfile]
unmanaged-devices=interface-name:cali*;interface-name:flannel*;interface-name:vxlan.calico
EOF

systemctl reload NetworkManager
```

## Firewalld Ports

All RKE2 nodes:

```bash
firewall-cmd --permanent --add-port=10250/tcp
firewall-cmd --permanent --add-port=8472/udp
firewall-cmd --permanent --add-port=9099/tcp
firewall-cmd --permanent --add-port=30000-32767/tcp
firewall-cmd --reload
```

Control plane only:

```bash
firewall-cmd --permanent --add-port=6443/tcp
firewall-cmd --permanent --add-port=9345/tcp
firewall-cmd --permanent --add-port=2379-2381/tcp
firewall-cmd --reload
```

Longhorn/RKE2 recovery firewall additions used later:

```bash
firewall-cmd --permanent --add-source=10.42.0.0/16 --zone=trusted
firewall-cmd --permanent --add-source=10.43.0.0/16 --zone=trusted
firewall-cmd --permanent --add-port=9500-9503/tcp
firewall-cmd --reload
```

## Control Plane Install

On `rke2-cp-01`:

```bash
mkdir -p /etc/rancher/rke2

cat > /etc/rancher/rke2/config.yaml <<'EOF'
node-name: rke2-cp-01
tls-san:
  - rke2-api.dclab.local
  - rke2-cp-01.dclab.local
write-kubeconfig-mode: "0644"
EOF

curl -sfL https://get.rke2.io | sh -
systemctl enable --now rke2-server
```

Get token:

```bash
cat /var/lib/rancher/rke2/server/node-token
```

## Worker Install

On each worker, replace token and node name:

```bash
mkdir -p /etc/rancher/rke2

cat > /etc/rancher/rke2/config.yaml <<'EOF'
server: https://rke2-api.dclab.local:9345
token: <CONTROL_PLANE_NODE_TOKEN>
node-name: rke2-worker-01
EOF

curl -sfL https://get.rke2.io | INSTALL_RKE2_TYPE=agent sh -
systemctl enable --now rke2-agent
```

## Utility Kubeconfig

On `utility-01`:

```bash
mkdir -p /root/.kube
scp root@rke2-cp-01.dclab.local:/etc/rancher/rke2/rke2.yaml /root/.kube/config
sed -i 's/127.0.0.1/rke2-api.dclab.local/' /root/.kube/config
scp root@rke2-cp-01.dclab.local:/var/lib/rancher/rke2/bin/kubectl /usr/local/bin/kubectl
chmod +x /usr/local/bin/kubectl
```

Validate:

```bash
export KUBECONFIG=/root/.kube/config
kubectl get nodes -o wide
kubectl get pods -A
```
