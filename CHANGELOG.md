# Changelog

This document includes a curated changelog for each release. We also publish a changelog as the description of a [GitHub release](https://github.com/nginxinc/nginx-kubernetes-gateway/releases), which, by contrast, is auto-generated and includes links to all PRs that went into the release.

## Release 0.1.0

*August 22, 2022*

This is an initial release of NGINX Kubernetes Gateway project.

The release includes:
- A control plane agent (a Kubernetes controller) that updates date plane (NGINX) configuration based on the state of the resources in the cluster.
- Support for NGINX as a data plane.
- Kubernetes manifests for a Deployment with a single Pod with the control plane and data plane containers as well as Services to enable external connectivity to that Pod.
- Support for a subset of features of GatewayClass, Gateway and HTTPRoute resources (see the [Gateway API Compatibility doc](https://github.com/nginxinc/nginx-kubernetes-gateway/blob/v0.1.0/README.md)).

We expect that the architecture of NGINX Kubernetes Gateway -- the number of pods and containers and their interaction -- will change as the project evolves.

NGINX Kubernetes Gateway is ready for experimental usage. We included the [docs](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/v0.1.0/docs) as well as [examples](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/v0.1.0/examples).

If you'd like to give us feedback or get involved, see the [README](https://github.com/nginxinc/nginx-kubernetes-gateway) to learn how.

COMPATIBILITY:
- The Gateway API version: `v1beta1`. 
- NGINX version: `1.21.3` 

CONTAINER IMAGES:
- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.1.0`
