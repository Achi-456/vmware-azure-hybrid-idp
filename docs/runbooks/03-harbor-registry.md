# 03 - Harbor Registry

Harbor runs standalone on `harbor-01`, outside Kubernetes.

## Disk Preparation

```bash
mkfs.xfs /dev/sdb
mkdir -p /data
echo '/dev/sdb /data xfs defaults 0 0' >> /etc/fstab
mount -a
mkdir -p /data/harbor
```

## Docker Runtime

```bash
dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
systemctl enable --now docker
```

## Harbor Install

Offline installer location:

```bash
/data/harbor/harbor
```

Key `harbor.yml` settings:

```yaml
hostname: harbor-01.dclab.local
data_volume: /data/harbor/data
http:
  port: 80
https:
  port: 443
  certificate: /data/harbor/certs/harbor.crt
  private_key: /data/harbor/certs/harbor.key
trivy:
  ignore_unfixed: false
  security_check: vuln
```

Install with Trivy:

```bash
cd /data/harbor/harbor
./install.sh --with-trivy
```

Reconfigure safely:

```bash
cd /data/harbor/harbor
./prepare
docker compose down
docker compose up -d
```

Do not use `docker compose down -v` unless intentionally deleting Harbor state.

## TLS CA And Server Cert

Certificates were generated under:

```bash
/root/harbor-certs
```

SAN policy:

```text
DNS.1 = harbor-01.dclab.local
DNS.2 = harbor.dclab.local
IP.1  = 172.25.188.93
```

Install certs into Harbor:

```bash
mkdir -p /data/harbor/certs
cp /root/harbor-certs/harbor.crt /data/harbor/certs/
cp /root/harbor-certs/harbor.key /data/harbor/certs/
```

Trust CA on RHEL nodes:

```bash
cp /tmp/dclab-ca.crt /etc/pki/ca-trust/source/anchors/dclab-ca.crt
update-ca-trust extract
```

Docker trust on `harbor-01`:

```bash
mkdir -p /etc/docker/certs.d/harbor-01.dclab.local
cp /root/harbor-certs/ca.crt /etc/docker/certs.d/harbor-01.dclab.local/ca.crt
systemctl restart docker
```

## Harbor Credentials

Do not put passwords into Git.

Credential files:

```bash
/root/.harbor-admin-password
/root/.harbor-rke2pull-password
```

Docker login:

```bash
ADMIN_PASS=$(cat /root/.harbor-admin-password)
echo "$ADMIN_PASS" | docker login harbor-01.dclab.local -u admin --password-stdin
```

## RKE2 Registry Integration

On each RKE2 node:

```yaml
mirrors:
  harbor-01.dclab.local:
    endpoint:
      - "https://harbor-01.dclab.local"
configs:
  harbor-01.dclab.local:
    auth:
      username: rke2pull
      password: <stored only on node>
```

Restart order:

```bash
systemctl restart rke2-agent   # workers, one at a time
systemctl restart rke2-server  # control plane last
```

Validate:

```bash
curl --cacert /root/harbor-certs/ca.crt https://harbor-01.dclab.local/api/v2.0/systeminfo
docker ps --format '{{.Names}} {{.Status}}'
```
