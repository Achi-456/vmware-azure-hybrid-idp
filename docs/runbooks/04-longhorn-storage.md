# 04 - Longhorn Storage

## Node Prerequisites

Run on all RKE2 nodes:

```bash
dnf install -y iscsi-initiator-utils nfs-utils cryptsetup
systemctl enable --now iscsid
systemctl is-active iscsid
curl -sS -o /dev/null -w '%{http_code}\n' https://harbor-01.dclab.local
```

## Image Mirroring Pattern

Longhorn images were mirrored into Harbor project `longhorn` before install.

Pattern:

```bash
docker pull <upstream-image>
docker tag <upstream-image> harbor-01.dclab.local/longhorn/<image>:<tag>
docker push harbor-01.dclab.local/longhorn/<image>:<tag>
```

## Helm Install

Install chart `longhorn/longhorn` version `1.6.2` into `longhorn-system`.

Important values:

```yaml
defaultSettings:
  defaultDataPath: /var/lib/longhorn
  defaultReplicaCount: 2
persistence:
  defaultClass: true
  defaultClassReplicaCount: 2
longhornUI:
  replicas: 1
```

Use Harbor image repository overrides for all Longhorn and CSI components.

Validate:

```bash
kubectl -n longhorn-system get pods
kubectl get storageclass
```

Expected default storage class:

```text
longhorn (default)
```

## PVC Smoke Test

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: longhorn-test-pvc
  namespace: demo
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: longhorn
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: longhorn-test-pod
  namespace: demo
spec:
  containers:
    - name: test
      image: harbor-01.dclab.local/platform/nginx:tls-test
      volumeMounts:
        - mountPath: /data
          name: test-vol
  volumes:
    - name: test-vol
      persistentVolumeClaim:
        claimName: longhorn-test-pvc
EOF

kubectl -n demo get pvc longhorn-test-pvc
kubectl -n demo exec longhorn-test-pod -- sh -c "echo 'longhorn works' > /data/test.txt"
kubectl -n demo exec longhorn-test-pod -- cat /data/test.txt
```

Cleanup:

```bash
kubectl -n demo delete pod longhorn-test-pod
kubectl -n demo delete pvc longhorn-test-pvc
```

## Longhorn UI

NodePort service:

```bash
kubectl -n longhorn-system get svc longhorn-frontend
```

Tunnel from Windows:

```powershell
ssh -i $env:USERPROFILE\.ssh\hybrid-cloud-idp -L 30080:172.25.188.90:30080 root@172.25.188.94
```

Open:

```text
http://localhost:30080
```
