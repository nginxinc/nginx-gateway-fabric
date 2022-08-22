[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway.svg?type=shield)](https://app.fossa.com/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway?ref=badge_shield)

# NGINX Kubernetes Gateway

NGINX Kubernetes Gateway is an open-source project that provides an implementation of the [Gateway API](https://gateway-api.sigs.k8s.io/) using [NGINX](https://nginx.org/) as the data plane. The goal of this project is to implement the core Gateway APIs -- `Gateway`, `GatewayClass`, `HTTPRoute`, `TCPRoute`, `TLSRoute`, and `UDPRoute` -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. NGINX Kubernetes Gateway is currently under development and supports a subset of the Gateway API.

For a list of supported Gateway API resources and features, see the [Gateway API Compatibility](docs/gateway-api-compatibility.md.md) doc.

> Warning: This project is actively in development (beta feature state) and should not be deployed in a production environment.
> All APIs, SDKs, designs, and packages are subject to change.

## Getting Started

1. [Quick Start on a kind cluster](docs/running-on-kind.md).
2. [Install](docs/installation.md) NGINX Kubernetes Gateway.
3. [Build](docs/building-the-image.md) an NGINX Kubernetes Gateway container image from source or use a pre-built image available on [GitHub Container Registry](https://github.com/nginxinc/nginx-kubernetes-gateway/pkgs/container/nginx-kubernetes-gateway)
4. Deploy various [examples](examples). 

## NGINX Kubernetes Gateway Releases

We publish NGINX Kubernetes Gateway releases on GitHub. See our [releases page](https://github.com/nginxinc/nginx-kubernetes-gateway/releases).

The latest release is [0.1.0](https://github.com/nginxinc/kubernetes-ingress/releases/tag/v0.1.0).

The edge version is useful for experimenting with new features that are not yet published in a release. To use, choose the *edge* version built from the [latest commit](https://github.com/nginxinc/nginx-kubernetes-gateway/commits/main) from the main branch.

To use NGINX Kubernetes Gateway, you need to have access to:
* An NGINX Kubernetes Gateway image.
* Installation manifests.
* Documentation and examples.

It is important that the versions of those things above match.

The table below summarizes the options regarding the images, manifests, documentation and examples and gives your links to the correct versions:

| Version | Description | Image | Installation Manifests | Documentation and Examples |
|-|-|-|-|-|
| Latest release | For experimental use | Use the 0.1.0 image from [GitHub](https://github.com/nginxinc/nginx-kubernetes-gateway/pkgs/container/nginx-kubernetes-gateway) | [Manifests](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/v0.1.0/deploy). | [Documentation](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/v0.1.0/docs). [Examples](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/v0.1.0/examples). |
| Edge| For experimental use and latest features | Use the edge image from [GitHub](https://github.com/nginxinc/nginx-kubernetes-gateway/pkgs/container/nginx-kubernetes-gateway) | [Manifests](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/main/deploy). | [Documentation](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/main/docs). [Examples](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/main/examples). |
## Technical Specifications

The following table lists the software versions NGINX Kubernetes Gateway supports.

| NGINX Kubernetes Gateway | Gateway API | Kubernetes | NGINX OSS |
|-|-|-|-|
| 0.1.0 | 0.5.0 | 1.19+ | 1.21.3|

## Contacts

We’d like to hear your feedback! If you have any suggestions or experience issues with our Gateway Controller, please create an issue or send a pull request on GitHub. You can contact us directly via kubernetes@nginx.com or on the [NGINX Community Slack](https://nginxcommunity.slack.com/channels/nginx-kubernetes-gateway) in the `#nginx-kubernetes-gateway` channel.

## Contributing

Please read our [Contributing guide](CONTRIBUTING.md) if you'd like to contribute to the project.

## Support

NGINX Kubernetes Gateway is not covered by any support contract.