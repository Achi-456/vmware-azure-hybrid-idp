# ADR 0003: MetalLB With Existing RKE2 Ingress

## Status

Accepted (Phase 3A)

## Context

The cluster already includes a healthy RKE2 bundled NGINX ingress controller:

- DaemonSet: `kube-system/rke2-ingress-nginx-controller`
- IngressClass: `nginx`
- Image: `rancher/nginx-ingress-controller:v1.14.5-hardened2`

Installing a second ingress controller would add duplicate controllers and another operational path without solving a new requirement.

## Decision

Install MetalLB `0.14.5` in L2 mode and expose the existing RKE2 ingress controller through a dedicated `LoadBalancer` service.

Chosen VIP:

- `172.25.188.96`

MetalLB configuration:

- IPAddressPool: `lab-ingress-pool`
- Address: `172.25.188.96/32`
- L2Advertisement: `lab-l2`
- Interface: `ens192`
- FRR disabled because Phase 3A uses only L2 advertisement.

Ingress service:

- Name: `rke2-ingress-nginx-controller-lb`
- Namespace: `kube-system`
- Type: `LoadBalancer`
- External IP: `172.25.188.96`

## Consequences

- Platform services can use host-based routing on one shared VIP.
- `gitea.dclab.local`, `argocd.dclab.local`, `grafana.dclab.local`, and `backstage.dclab.local` point to the same ingress IP.
- Existing RKE2 ingress remains the default ingress path.
- ICMP ping is not used as the final validation method for the VIP; HTTP routing is.
- Strict Harbor-only control of the existing RKE2 ingress image is deferred to later GitOps hardening.
