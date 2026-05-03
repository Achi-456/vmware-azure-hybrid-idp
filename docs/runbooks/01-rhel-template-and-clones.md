# 01 - RHEL Template And Clones

## RHEL 8.8 Base Packages

```bash
dnf install -y open-vm-tools cloud-init perl git curl vim-enhanced bash-completion
systemctl enable --now vmtoolsd
systemctl enable cloud-init
```

## Local ISO Repository

Mount ISO:

```bash
mkdir -p /mnt/rhel-iso
mount /dev/sr0 /mnt/rhel-iso
```

Persist ISO mount:

```bash
echo '/dev/sr0 /mnt/rhel-iso iso9660 ro,nofail 0 0' >> /etc/fstab
```

Create repo file:

```bash
cat > /etc/yum.repos.d/rhel8-iso.repo <<'EOF'
[rhel8-iso-baseos]
name=RHEL 8.8 ISO BaseOS
baseurl=file:///mnt/rhel-iso/BaseOS
enabled=1
gpgcheck=0

[rhel8-iso-appstream]
name=RHEL 8.8 ISO AppStream
baseurl=file:///mnt/rhel-iso/AppStream
enabled=1
gpgcheck=0
EOF
```

Validate:

```bash
dnf repolist
rpm -q open-vm-tools cloud-init git curl vim-enhanced bash-completion
systemctl is-active vmtoolsd
systemctl is-enabled cloud-init
```

## Clean Template Before Cloning

```bash
cloud-init clean 2>/dev/null || true
truncate -s 0 /etc/machine-id
rm -f /var/lib/dbus/machine-id
ln -sf /etc/machine-id /var/lib/dbus/machine-id
rm -f /etc/ssh/ssh_host_*
rm -f /etc/udev/rules.d/70-persistent-net.rules
rm -f /var/log/wtmp /var/log/btmp
history -c
shutdown -h now
```

## Persistent Network Standard

Example for a node:

```bash
nmcli con mod "System ens192" \
  ipv4.method manual \
  ipv4.addresses 172.25.188.90/16 \
  ipv4.gateway 172.25.188.1 \
  ipv4.dns 172.25.188.20 \
  ipv4.dns-search dclab.local \
  ipv4.ignore-auto-dns yes \
  ipv4.ignore-auto-routes yes \
  ipv4.may-fail no \
  connection.autoconnect yes \
  connection.autoconnect-priority 100

nmcli con up "System ens192"
```

Disable cloud-init networking:

```bash
mkdir -p /etc/cloud/cloud.cfg.d
echo 'network: {config: disabled}' > /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
```

## Edge Clone Cleanup

`edge-01` was accidentally cloned from a cluster VM and initially had the wrong hostname. Cleanup used:

```bash
systemctl stop rke2-server rke2-agent 2>/dev/null || true
systemctl disable rke2-server rke2-agent 2>/dev/null || true
rm -rf /etc/rancher/rke2 /var/lib/rancher/rke2 /var/lib/kubelet /etc/cni/net.d /var/lib/cni
hostnamectl set-hostname edge-01.dclab.local
grep -q '172.25.188.95 edge-01.dclab.local edge-01' /etc/hosts || echo '172.25.188.95 edge-01.dclab.local edge-01' >> /etc/hosts
```
