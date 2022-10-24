# Changelog

This document includes a curated changelog for each release. We also publish a changelog as the description of a [GitHub release](https://github.com/nginxinc/nginx-kubernetes-gateway/releases), which, by contrast, is auto-generated and includes links to all PRs that went into the release.

## Release 0.2.0

*October 25, 2022*

This release extends the support of the features of the Gateway API resources.

FEATURES:
* Support the Pod IPs instead of the virtual IP of a Service in the NGINX upstream. Additionally, NGINX Kubernetes Gateway will pick up any changes to the Pod IPs and update the NGINX upstream accordingly. [PR-221](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/221)
* Support the redirect filter in an HTTPRoute rule. [PR-218](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/218)
* Support weights in backendRefs in the HTTPRoute (traffic splitting). [PR-261](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/261)
* Support the ObservedGeneration field in the HTTPRoute status. [PR-254](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/254)

BUG FIXES:
* Do not require the namespace in the `--gateway-ctlr-name` cli argument. [PR-235](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/235)
* Ensure NGINX Kubernetes Gateway exits gracefully during shutdown. [PR-250](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/250)
* Handle query param names in case-sensitive way. [PR-220](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/220)

DEPENDENCIES:
* Use the latest NGINX 1.23 image. [PR-275](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/275)
* Bump sigs.k8s.io/gateway-api from 0.5.0 to 0.5.1 [PR-251](https://github.com/nginxinc/nginx-kubernetes-gateway/pull/251)


COMPATIBILITY:
- The Gateway API version: `0.5.1`
- NGINX version: `1.21.x` * 
- Kubernetes version: `1.21+`

\*the installation manifests use the `nginx:1.21` image, which always points to the latest version of 1.21.x releases.

CONTAINER IMAGES:
- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.2.0`

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
- The Gateway API version: `0.5.0`
- NGINX version: `1.21.3` 
- Kubernetes version: `1.19+`

CONTAINER IMAGES:
- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.1.0`
