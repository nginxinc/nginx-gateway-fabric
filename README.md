[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway.svg?type=shield)](https://app.fossa.com/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway?ref=badge_shield)

# NGINX Kubernetes Gateway

NGINX Kubernetes Gateway is an open-source project that provides an implementation of the [Gateway API](https://gateway-api.sigs.k8s.io/) using [NGINX](https://nginx.org/) as the data plane. The goal of this project is to implement the core Gateway APIs -- `Gateway`, `GatewayClass`, `HTTPRoute`, `TCPRoute`, `TLSRoute`, and `UDPRoute` -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. NGINX Kubernetes Gateway is currently under development and supports a subset of the Gateway API.

> Warning: This project is actively in development (pre-alpha feature state) and should not be deployed in a production environment.
> All APIs, SDKs, designs, and packages are subject to change.

## Run NGINX Kubernetes Gateway

1. [Build](docs/building-the-image.md) the NGINX Kubernetes Gateway container image.
2. [Install](docs/installation.md) NGINX Kubernetes Gateway.
3. Deploy various [examples](examples). 
