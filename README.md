[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway.svg?type=shield)](https://app.fossa.com/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway?ref=badge_shield)

# NGINX Kubernetes Gateway

NGINX Kubernetes Gateway is an open-source project that provides an implementation of the [Gateway API](https://gateway-api.sigs.k8s.io/) using [NGINX](https://nginx.org/) as the data plane. The goal of this project is to implement the core Gateway APIs -- `Gateway`, `GatewayClass`, `HTTPRoute`, `TCPRoute`, `TLSRoute`, and `UDPRoute` -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. NGINX Kubernetes Gateway is currently under development and supports a subset of the Gateway API.

> Warning: This project is actively in development (beta feature state) and should not be deployed in a production environment.
> All APIs, SDKs, designs, and packages are subject to change.

## Getting Started

1. [Quick Start on a kind cluster](docs/running-on-kind.md).
2. [Build](docs/building-the-image.md) the NGINX Kubernetes Gateway container image.
3. [Install](docs/installation.md) NGINX Kubernetes Gateway.
4. Deploy various [examples](examples). 

## Technical Specifications

The following table lists the software versions NGINX Kubernetes Gateway supports.

| NGINX Kubernetes Gateway | Gateway API | Kubernetes | NGINX OSS |
|-|-|-|-|
| 0.1.0 | 0.5.0 | 1.16+ | 1.21.3|

## Contacts

We’d like to hear your feedback! If you have any suggestions or experience issues with our Gateway Controller, please create an issue or send a pull request on GitHub. You can contact us directly via kubernetes@nginx.com or on the [NGINX Community Slack](https://nginxcommunity.slack.com/channels/nginx-kubernetes-gateway) in the `#nginx-kubernetes-gateway` channel.

## Contributing

Please read our [Contributing guide](CONTRIBUTING.md) if you'd like to contribute to the project.

## Support

NGINX Kubernetes Gateway is not covered by any support contract.