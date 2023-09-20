# Changelog

This document includes a curated changelog for each release. We also publish a changelog as the description of
a [GitHub release](https://github.com/nginxinc/nginx-gateway-fabric/releases), which, by contrast, is auto-generated
and includes links to all PRs that went into the release.

## Release 0.6.0

*August 31, 2023*

This release adds a Helm chart, dynamic control plane logging, Prometheus metrics, and in-depth guides for various use cases.

FEATURES:

- Helm chart. [PR-840](https://github.com/nginxinc/nginx-gateway-fabric/pull/840)
- Use custom nginx container. [PR-934](https://github.com/nginxinc/nginx-gateway-fabric/pull/934)
- Support dynamic control plane logging. [PR-943](https://github.com/nginxinc/nginx-gateway-fabric/pull/943)
- Support websocket connections. [PR-962](https://github.com/nginxinc/nginx-gateway-fabric/pull/962)
- Support Prometheus metrics. [PR-999](https://github.com/nginxinc/nginx-gateway-fabric/pull/999)

BUG FIXES:

- Ensure NGINX Kubernetes Gateway has least privileges. [PR-1004](https://github.com/nginxinc/nginx-gateway-fabric/pull/1004)

DOCUMENTATION:

- Use case guides: https://github.com/nginxinc/nginx-gateway-fabric/tree/v0.6.0/docs/guides

COMPATIBILITY:

- The Gateway API version: `0.8.0`
- NGINX version: `1.25.2`
- Kubernetes version: `1.23+`

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.6.0`
- Data plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway/nginx:0.6.0`

## Release 0.5.0

*July 17, 2023*

This release completes all v1beta1 Core features of the Gateway API resources. See the [Gateway Compatibility doc](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.5.0/docs/gateway-api-compatibility.md)

FEATURES:

- Support cross-namespace BackendRefs in HTTPRoutes. [PR-806](https://github.com/nginxinc/nginx-gateway-fabric/pull/806)
- Support dynamic certificate rotation with Kubernetes Secrets. [PR-807](https://github.com/nginxinc/nginx-gateway-fabric/pull/807)
- Support SupportedKinds in ListenerStatus. [PR-809](https://github.com/nginxinc/nginx-gateway-fabric/pull/809)

BUG FIXES:

- Set redirect port in location header according to the scheme. [PR-801](https://github.com/nginxinc/nginx-gateway-fabric/pull/801)
- Set proxy host header to the exact value of the request host header. [PR-827](https://github.com/nginxinc/nginx-gateway-fabric/pull/827)
- Ensure Prefix matching requires trailing slash. [PR-817](https://github.com/nginxinc/nginx-gateway-fabric/pull/817)

COMPATIBILITY:

- The Gateway API version: `0.7.1`
- NGINX version: `1.25.x` *
- Kubernetes version: `1.21+`

\*the installation manifests use the `nginx:1.25` image, which always points to the latest version of 1.25.x releases.

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.5.0`

## Release 0.4.0

*July 6, 2023*

This release brings:

- Support for more features of the Gateway API resources. See the [Gateway Compatibility doc](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.4.0/docs/gateway-api-compatibility.md)
- Support for running the conformance test suite. See the [Conformance tests README](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.4.0/conformance/README.md).
- Defined Enhancement Proposal process for NGINX Kubernetes Gateway project. See the [Enhancement Proposal README](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.4.0/docs/proposals/README.md).
- Improved developer documentation for contributing to the project. See the [Development quickstart](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.4.0/docs/developer/quickstart.md).
- Architecture document that explains how NGINX Kubernetes Gateway works at a high level. See the [Architecture doc](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.4.0/docs/architecture.md)
- Miscellaneous enhancements and bug fixes.

FEATURES:

- Allow empty sectionName in HTTPRoute parentRef. [PR-626](https://github.com/nginxinc/nginx-gateway-fabric/pull/626)
- Exact PathMatch support for HTTPRoutes. [PR-603](https://github.com/nginxinc/nginx-gateway-fabric/pull/603)
- Set ResolvedRefs condition to true on HTTPRoutes. [PR-645](https://github.com/nginxinc/nginx-gateway-fabric/pull/645)
- Set gateway Pod IP as GatewayStatus address. [PR-638](https://github.com/nginxinc/nginx-gateway-fabric/pull/638)
- Set Accepted condition type on Gateway status. [PR-633](https://github.com/nginxinc/nginx-gateway-fabric/pull/633)
- Drop unrequired capabilities from containers. [PR-677](https://github.com/nginxinc/nginx-gateway-fabric/pull/677)
- Update route condition where listener is not found. [PR-675](https://github.com/nginxinc/nginx-gateway-fabric/pull/675)
- Set Gateway Programmed condition. [PR-658](https://github.com/nginxinc/nginx-gateway-fabric/pull/658)
- AllowedRoutes support for Listeners. [PR-721](https://github.com/nginxinc/nginx-gateway-fabric/pull/721)
- Support custom listener ports. [PR-745](https://github.com/nginxinc/nginx-gateway-fabric/pull/745)
- Add support for RequestHeaderModifier for HTTPRouteRule objects. [PR-717](https://github.com/nginxinc/nginx-gateway-fabric/pull/717)
- Add wildcard hostname support. [PR-769](https://github.com/nginxinc/nginx-gateway-fabric/pull/769)
- Add Programmed status for listener. [PR-786](https://github.com/nginxinc/nginx-gateway-fabric/pull/786)
- ReferenceGrant from Gateway to Secret. [PR-791](https://github.com/nginxinc/nginx-gateway-fabric/pull/791)

BUG FIXES:

- Set upstream zone size to 512k. [PR-609](https://github.com/nginxinc/nginx-gateway-fabric/pull/609)
- Allow empty HTTPRoute hostnames. [PR-650](https://github.com/nginxinc/nginx-gateway-fabric/pull/650)
- Allow long server names. [PR-651](https://github.com/nginxinc/nginx-gateway-fabric/pull/651)
- Add in required capabilities for writing TLS secrets. [PR-718](https://github.com/nginxinc/nginx-gateway-fabric/pull/718)
- Fix binding to multiple listeners with empty section name. [PR-730](https://github.com/nginxinc/nginx-gateway-fabric/pull/730)
- Add timeout and retry logic for finding NGINX PID file. [PR-676](https://github.com/nginxinc/nginx-gateway-fabric/pull/676)
- Prioritize method matching. [PR-789](https://github.com/nginxinc/nginx-gateway-fabric/pull/789)
- Add NewListenerInvalidRouteKinds condition. [PR-799](https://github.com/nginxinc/nginx-gateway-fabric/pull/799)
- Set ResolvedRefs/False/InvalidKind condition on the HTTPRoute if a BackendRef specifies an unknown kind. [PR-800](https://github.com/nginxinc/nginx-gateway-fabric/pull/800)
- Set GatewayClass status for ignored GatewayClasses. [PR-804](https://github.com/nginxinc/nginx-gateway-fabric/pull/804)

DEPENDENCIES:

- Bump sigs.k8s.io/gateway-api from 0.7.0 to 0.7.1. [PR-711](https://github.com/nginxinc/nginx-gateway-fabric/pull/711)

COMPATIBILITY:

- The Gateway API version: `0.7.1`
- NGINX version: `1.25.x` *
- Kubernetes version: `1.21+`

\*the installation manifests use the `nginx:1.25` image, which always points to the latest version of 1.25.x releases.

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.4.0`

## Release 0.3.0

*April 24, 2023*

This release brings:

- Extensive validation of Gateway API resources for robustness, security and correctness. See the [validation doc](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.3.0/docs/resource-validation.md)
for more details.
- Defined open-source development process for NGINX Kubernetes Gateway project. See the [Issue lifecycle doc](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.3.0/ISSUE_LIFECYCLE.md).
- Miscellaneous enhancements and bug fixes.

FEATURES:

- Report proper Conditions in status of HTTPRoute and Gateway when GatewayClass is invalid or doesn't exist. [PR-576](https://github.com/nginxinc/nginx-gateway-fabric/pull/576)
- Implement NKG-specific field validation for GatewayClasses. [PR-295](https://github.com/nginxinc/nginx-gateway-fabric/pull/495)
- Implement NKG-specific field validation for HTTPRoutes. [PR-455](https://github.com/nginxinc/nginx-gateway-fabric/pull/455)
- Implement NKG-specific field validation for Gateways. [PR-407](https://github.com/nginxinc/nginx-gateway-fabric/pull/407)
- Run webhook validation rules inside NKG control plane. [PR-388](https://github.com/nginxinc/nginx-gateway-fabric/pull/388)
- Make NGINX error log visible in NGINX container logs. [PR-319](https://github.com/nginxinc/nginx-gateway-fabric/pull/319)
- Always generate a root "/" location block in NGINX config to handle unmatched requests with 404 response. [PR-356](https://github.com/nginxinc/nginx-gateway-fabric/pull/356)

BUG FIXES:

- Fix HTTPRoute section name related bugs. [PR-568](https://github.com/nginxinc/nginx-gateway-fabric/pull/568)
- Fix Observed Generation for Gateway Status. [PR-351](https://github.com/nginxinc/nginx-gateway-fabric/pull/351)
- Fix status for parentRef with invalid listener in HTTPRoute. [PR-350](https://github.com/nginxinc/nginx-gateway-fabric/pull/350)
- Fix initContainer failure during pod restart. [PR-337](https://github.com/nginxinc/nginx-gateway-fabric/pull/337).
  Thanks to [Tom Plant](https://github.com/pl4nty)
- Generate default http server in NGINX if http listener exists in Gateway. [PR-320](https://github.com/nginxinc/nginx-gateway-fabric/pull/320)

DEPENDENCIES:

- Bump sigs.k8s.io/gateway-api from 0.6.0 to 0.6.2. [PR-471](https://github.com/nginxinc/nginx-gateway-fabric/pull/471)

COMPATIBILITY:

- The Gateway API version: `0.6.2`
- NGINX version: `1.23.x` *
- Kubernetes version: `1.21+`

\*the installation manifests use the `nginx:1.23` image, which always points to the latest version of 1.23.x releases.

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.3.0`

## Release 0.2.0

*October 25, 2022*

This release extends the support of the features of the Gateway API resources.

FEATURES:

- Support the Pod IPs instead of the virtual IP of a Service in the NGINX upstream. Additionally, NGINX Gateway
  Fabric will pick up any changes to the Pod IPs and update the NGINX upstream
  accordingly. [PR-221](https://github.com/nginxinc/nginx-gateway-fabric/pull/221)
- Support the redirect filter in an HTTPRoute rule. [PR-218](https://github.com/nginxinc/nginx-gateway-fabric/pull/218)
- Support weights in backendRefs in the HTTPRoute (traffic splitting). [PR-261](https://github.com/nginxinc/nginx-gateway-fabric/pull/261)
- Support the ObservedGeneration field in the HTTPRoute status. [PR-254](https://github.com/nginxinc/nginx-gateway-fabric/pull/254)

BUG FIXES:

- Do not require the namespace in the `--gateway-ctlr-name` cli argument. [PR-235](https://github.com/nginxinc/nginx-gateway-fabric/pull/235)
- Ensure NGINX Kubernetes Gateway exits gracefully during shutdown. [PR-250](https://github.com/nginxinc/nginx-gateway-fabric/pull/250)
- Handle query param names in case-sensitive way. [PR-220](https://github.com/nginxinc/nginx-gateway-fabric/pull/220)

DEPENDENCIES:

- Use the latest NGINX 1.23 image. [PR-275](https://github.com/nginxinc/nginx-gateway-fabric/pull/275)
- Bump sigs.k8s.io/gateway-api from 0.5.0 to 0.5.1 [PR-251](https://github.com/nginxinc/nginx-gateway-fabric/pull/251)


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

- A control plane agent (a Kubernetes controller) that updates date plane (NGINX) configuration based on the state of
  the resources in the cluster.
- Support for NGINX as a data plane.
- Kubernetes manifests for a Deployment with a single Pod with the control plane and data plane containers as well as
  Services to enable external connectivity to that Pod.
- Support for a subset of features of GatewayClass, Gateway and HTTPRoute resources (see the [Gateway API Compatibility doc](https://github.com/nginxinc/nginx-gateway-fabric/blob/v0.1.0/README.md)).

We expect that the architecture of NGINX Kubernetes Gateway -- the number of pods and containers and their
interaction -- will change as the project evolves.

NGINX Kubernetes Gateway is ready for experimental usage. We included
the [docs](https://github.com/nginxinc/nginx-gateway-fabric/tree/v0.1.0/docs) as well
as [examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/v0.1.0/examples).

If you'd like to give us feedback or get involved, see
the [README](https://github.com/nginxinc/nginx-gateway-fabric) to learn how.

COMPATIBILITY:

- The Gateway API version: `0.5.0`
- NGINX version: `1.21.3`
- Kubernetes version: `1.19+`

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-kubernetes-gateway:0.1.0`
