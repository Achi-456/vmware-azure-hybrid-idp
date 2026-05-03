# 06 - MetalLB And RKE2 Ingress

Phase 3A exposes the existing RKE2 bundled NGINX ingress controller with one MetalLB VIP.

## Preflight

```bash
ping -c 2 -W 1 172.25.188.96 || true
ip neigh show 172.25.188.96 || true
getent ahostsv4 gitea.dclab.local | head -n 1
```

Expected before MetalLB assignment:

```text
100% packet loss
172.25.188.96 dev ens192 INCOMPLETE
172.25.188.96 STREAM gitea.dclab.local
```

Check existing RKE2 ingress:

```bash
kubectl -n kube-system get ds rke2-ingress-nginx-controller
kubectl get ingressclass nginx
```

## Mirror Images To Harbor

On `harbor-01`:

```bash
ADMIN_PASS=$(cat /root/.harbor-admin-password)
echo "$ADMIN_PASS" | docker login harbor-01.dclab.local -u admin --password-stdin

for PROJECT in metallb platform; do
  curl -s -o /tmp/harbor-project-$PROJECT.txt -w "%{http_code}" \
    --cacert /root/harbor-certs/ca.crt \
    -u admin:"$ADMIN_PASS" \
    -X POST https://harbor-01.dclab.local/api/v2.0/projects \
    -H "Content-Type: application/json" \
    -d "{\"project_name\":\"${PROJECT}\",\"public\":true}"
done

docker pull quay.io/metallb/controller:v0.14.5
docker tag quay.io/metallb/controller:v0.14.5 harbor-01.dclab.local/metallb/controller:v0.14.5
docker push harbor-01.dclab.local/metallb/controller:v0.14.5

docker pull quay.io/metallb/speaker:v0.14.5
docker tag quay.io/metallb/speaker:v0.14.5 harbor-01.dclab.local/metallb/speaker:v0.14.5
docker push harbor-01.dclab.local/metallb/speaker:v0.14.5

docker pull hashicorp/http-echo:latest
docker tag hashicorp/http-echo:latest harbor-01.dclab.local/platform/http-echo:latest
docker push harbor-01.dclab.local/platform/http-echo:latest
```

## Install MetalLB

On `utility-01`:

```bash
export KUBECONFIG=/root/.kube/config

helm repo add metallb https://metallb.github.io/metallb
helm repo update

helm upgrade --install metallb metallb/metallb \
  --namespace metallb-system \
  --create-namespace \
  --version 0.14.5 \
  --set controller.image.repository=harbor-01.dclab.local/metallb/controller \
  --set controller.image.tag=v0.14.5 \
  --set speaker.image.repository=harbor-01.dclab.local/metallb/speaker \
  --set speaker.image.tag=v0.14.5 \
  --set speaker.frr.enabled=false \
  --wait \
  --timeout 5m
```

Validate images:

```bash
kubectl -n metallb-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{" => "}{.spec.containers[*].image}{"\n"}{end}'
```

## Configure VIP Pool

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: lab-ingress-pool
  namespace: metallb-system
spec:
  addresses:
    - 172.25.188.96/32
  autoAssign: true
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: lab-l2
  namespace: metallb-system
spec:
  ipAddressPools:
    - lab-ingress-pool
  interfaces:
    - ens192
EOF
```

## Expose RKE2 Ingress

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: rke2-ingress-nginx-controller-lb
  namespace: kube-system
spec:
  type: LoadBalancer
  loadBalancerIP: 172.25.188.96
  selector:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: rke2-ingress-nginx
    app.kubernetes.io/name: rke2-ingress-nginx
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 80
    - name: https
      port: 443
      protocol: TCP
      targetPort: 443
EOF
```

If service has both `loadBalancerIP` and `metallb.universe.tf/loadBalancerIPs`, remove the annotation:

```bash
kubectl -n kube-system annotate svc rke2-ingress-nginx-controller-lb metallb.universe.tf/loadBalancerIPs- || true
```

Validate:

```bash
kubectl -n kube-system get svc rke2-ingress-nginx-controller-lb
kubectl -n kube-system describe svc rke2-ingress-nginx-controller-lb
```

Expected:

```text
EXTERNAL-IP: 172.25.188.96
```

## Echo Ingress Test

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-test
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: echo-test
  template:
    metadata:
      labels:
        app: echo-test
    spec:
      containers:
        - name: echo
          image: harbor-01.dclab.local/platform/http-echo:latest
          args:
            - "-text=MetalLB + NGINX routing works on dclab.local"
          ports:
            - containerPort: 5678
---
apiVersion: v1
kind: Service
metadata:
  name: echo-test
  namespace: default
spec:
  selector:
    app: echo-test
  ports:
    - port: 80
      targetPort: 5678
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: echo-test
  namespace: default
spec:
  ingressClassName: nginx
  rules:
    - host: echo.dclab.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: echo-test
                port:
                  number: 80
EOF

kubectl rollout status deployment/echo-test --timeout=180s
curl --resolve "echo.dclab.local:80:172.25.188.96" -sS http://echo.dclab.local
```

Expected:

```text
MetalLB + NGINX routing works on dclab.local
```

Cleanup test app only:

```bash
kubectl delete ingress echo-test --ignore-not-found
kubectl delete service echo-test --ignore-not-found
kubectl delete deployment echo-test --ignore-not-found
```

## Important Note

The VIP may not respond to ICMP ping. Use HTTP routing as the real validation.
