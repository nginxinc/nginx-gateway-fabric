# Changelog

This document includes a curated changelog for each release. We also publish a changelog as the description of
a [GitHub release](https://github.com/nginxinc/nginx-gateway-fabric/releases), which, by contrast, is auto-generated
and includes links to all PRs that went into the release.

## Release 1.3.0

* Update minimum k8s version to 1.25 by @lucacome in https://github.com/nginxinc/nginx-gateway-fabric/pull/1885
* Support NginxProxy CRD and global tracing settings by @sjberman in https://github.com/nginxinc/nginx-gateway-fabric/pull/1870
* Remove zone size for invalid backend ref by @bjee19 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1931
* Add directive for SSL Server block by @salonichf5 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1934
* Add support for ResponseHeaderModifier for HTTPRouteRule objects by @salonichf5 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1880
* Fix rbac ServiceAccount imagePullSecrets template by @bjee19 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1953
* Add request header filter support for gRPC by @ciarams87 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1909
* Add field to NginxProxy to allow disabling HTTP2 by @ciarams87 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1925
* Support response header filter for GRPCRoute by @ciarams87 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1983
* Upgrade Gateway API to v1.1.0 by @ciarams87 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1975
* Collect backendTLSPolicy and GRPCRoute Count by @salonichf5 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1954
* Implement ClientSettingsPolicy by @kate-osborn in https://github.com/nginxinc/nginx-gateway-fabric/pull/1940
* Allow NGF to run on Openshift  by @bjee19 in https://github.com/nginxinc/nginx-gateway-fabric/pull/1976
* Bump NGINX Plus to R32 by @lucacome in https://github.com/nginxinc/nginx-gateway-fabric/pull/2057
* Support tracing via the ObservabilityPolicy by @sjberman in https://github.com/nginxinc/nginx-gateway-fabric/pull/2004

%%DATE%%

FEATURES:

-

BUG FIXES:

-

DOCUMENTATION:

-

HELM CHART:

- The version of the Helm chart is now 1.3.0
-

UPGRADE:

-

KNOWN ISSUES:

-

COMPATIBILITY:

- The Gateway API version: ``
- NGINX version: ``
- NGINX Plus version: ``
- Kubernetes version: ``

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-gateway-fabric:1.3.0`
- Data plane: `ghcr.io/nginxinc/nginx-gateway-fabric/nginx:1.3.0`
- Data plane with NGINX Plus: `private-registry.nginx.com/nginx-gateway-fabric/nginx-plus:1.3.0`
## Release 1.2.0

*March 21, 2024*

FEATURES:

- [NGINX Plus](https://docs.nginx.com/nginx-gateway-fabric/overview/nginx-plus) can now be used as the data plane. [PR-1394](https://github.com/nginxinc/nginx-gateway-fabric/pull/1394)
  - Supports dynamic upstream reloads. [PR-1469](https://github.com/nginxinc/nginx-gateway-fabric/pull/1469)
  - Contains advanced Prometheus metrics. [PR-1394](https://github.com/nginxinc/nginx-gateway-fabric/pull/1394)
  - Includes the NGINX Plus monitoring dashboard. [PR-1488](https://github.com/nginxinc/nginx-gateway-fabric/pull/1488)
- Support for [BackendTLSPolicy](https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/). [PR-1487](https://github.com/nginxinc/nginx-gateway-fabric/pull/1487)
- Support for URLRewrite HTTPRoute Filter. [PR-1396](https://github.com/nginxinc/nginx-gateway-fabric/pull/1396)
- NGINX Gateway Fabric will collect and report product telemetry to an F5 telemetry service every 24h. Read https://docs.nginx.com/nginx-gateway-fabric/overview/product-telemetry/ for more info, including what gets collected and how to opt out. [PR-1699](https://github.com/nginxinc/nginx-gateway-fabric/pull/1699)

ENHANCEMENTS:

- Stop processing resources that haven't changed. [PR-1422](https://github.com/nginxinc/nginx-gateway-fabric/pull/1422) Thanks to [Kai-Hsun Chen](https://github.com/kevin85421).
- Maintain Gateway Status order. [PR-1324](https://github.com/nginxinc/nginx-gateway-fabric/pull/1324) Thanks to [Kai-Hsun Chen](https://github.com/kevin85421).

BUG FIXES:

- Prevent paths in HTTPRoute matches from conflicting with internal locations in NGINX. [PR-1445](https://github.com/nginxinc/nginx-gateway-fabric/pull/1445)

DOCUMENTATION:

- Sample Grafana dashboard added. [PR-1620](https://github.com/nginxinc/nginx-gateway-fabric/pull/1620)
- Add a document about how to get support. [PR-1388](https://github.com/nginxinc/nginx-gateway-fabric/pull/1388)
- [Documentation](https://docs.nginx.com/nginx-gateway-fabric/installation/ngf-images) on how to build or install the NGINX Plus image.

HELM CHART:

- The version of the Helm chart is now 1.2.0
- nodeSelector is now configurable. [PR-1531](https://github.com/nginxinc/nginx-gateway-fabric/pull/1531) Thanks to [Leandro Martins](https://github.com/leandrocostam)

KNOWN ISSUES:

- Shutdown of non-leader Pods starts leader jobs. [1738](https://github.com/nginxinc/nginx-gateway-fabric/issues/1738)
- Too many matching conditions can cause reload errors. [1107](https://github.com/nginxinc/nginx-gateway-fabric/issues/1107)
- NGF Pod fails to become ready due to nginx reload failure. [1695](https://github.com/nginxinc/nginx-gateway-fabric/issues/1695)

COMPATIBILITY:

- The Gateway API version: `1.0.0`
- NGINX version: `1.25.4`
- NGINX Plus version: `R31`
- Kubernetes version: `1.23+`

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-gateway-fabric:1.2.0`
- Data plane: `ghcr.io/nginxinc/nginx-gateway-fabric/nginx:1.2.0`
- Data plane with NGINX Plus: `private-registry.nginx.com/nginx-gateway-fabric/nginx-plus:1.2.0`

## Release 1.1.0

*December 14, 2023*

This release updates NGINX Gateway Fabric to support version 1.0.0 of the Gateway API in addition to bug fixes and documentation updates. Our docs are now available at https://docs.nginx.com/nginx-gateway-fabric.

FEATURES:

- Update to v1.0.0 of the Gateway API. [PR-1250](https://github.com/nginxinc/nginx-gateway-fabric/pull/1250)
- Set the SupportedVersion Condition on GatewayClass. [PR-1301](https://github.com/nginxinc/nginx-gateway-fabric/pull/1301)

BUG FIXES:

- Merge HTTPRoute conditions from all Gateway controllers. [PR-1220](https://github.com/nginxinc/nginx-gateway-fabric/pull/1220)
- Validate header names and report validation errors in the HTTPRoute status. [PR-1239](https://github.com/nginxinc/nginx-gateway-fabric/pull/1239)
- Remove usage info from log output. [PR-1242](https://github.com/nginxinc/nginx-gateway-fabric/pull/1242)
- Set the Gateway Listener status AttachedRoutes field to the number of Routes associated with a Listener regardless of Gateway or Route status. [PR-1275](https://github.com/nginxinc/nginx-gateway-fabric/pull/1275)
- Set file mode explicitly for regular NGINX configuration files. [PR-1323](https://github.com/nginxinc/nginx-gateway-fabric/pull/1323). Thanks to [Kai-Hsun Chen](https://github.com/kevin85421).

DOCUMENTATION:

- Documentation is now available on docs.nginx.com. [Link](https://docs.nginx.com/nginx-gateway-fabric/)
- Update the resource validation documents to cover CEL validation. [Link](https://docs.nginx.com/nginx-gateway-fabric/overview/resource-validation/)
- Non-functional testing results. [Link](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/tests)

HELM CHART:

- The version of the Helm chart is now 1.1.0.
- Add tolerations to the helm chart. [PR-1192](https://github.com/nginxinc/nginx-gateway-fabric/pull/1192). Thanks to [Jerome Brown](https://github.com/oWretch).
- Add extra volume mounts to the helm chart. [PR-1193](https://github.com/nginxinc/nginx-gateway-fabric/pull/1193). Thanks to [Jerome Brown](https://github.com/oWretch).
- Fix broken helm chart icon links. [PR-1290](https://github.com/nginxinc/nginx-gateway-fabric/pull/1290). Thanks to [arukiidou](https://github.com/arukiidou).

UPGRADE:

- This version of NGINX Gateway Fabric is not compatible with v0.8.0 of the Gateway API. You must upgrade the Gateway API CRDs to v1.0.0 before upgrading NGINX Gateway Fabric. For instructions, see the upgrade documentation for [helm](https://docs.nginx.com/nginx-gateway-fabric/installation/installing-ngf/helm/#upgrade-nginx-gateway-fabric) or [manifests](https://docs.nginx.com/nginx-gateway-fabric/installation/installing-ngf/manifests/#upgrade-nginx-gateway-fabric).

COMPATIBILITY:

- The Gateway API version: `1.0.0`
- NGINX version: `1.25.3`
- Kubernetes version: `1.23+`

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-gateway-fabric:1.1.0`
- Data plane: `ghcr.io/nginxinc/nginx-gateway-fabric/nginx:1.1.0`

## Release 1.0.0

*October 24, 2023*

This is the official v1.0.0 release of NGINX Gateway Fabric.

BREAKING CHANGES:

- Rename the product from NGINX Kubernetes Gateway to NGINX Gateway Fabric. [PR-1070](https://github.com/nginxinc/nginx-gateway-fabric/pull/1070)

FEATURES:

- Add readiness probe. [PR-1047](https://github.com/nginxinc/nginx-gateway-fabric/pull/1047)
- Support horizontal scaling. [PR-1048](https://github.com/nginxinc/nginx-gateway-fabric/pull/1048)
- Add NGINX reload metrics. [PR-1049](https://github.com/nginxinc/nginx-gateway-fabric/pull/1049)
- Retry status updater on failures. [PR-1062](https://github.com/nginxinc/nginx-gateway-fabric/pull/1062)
- Add event processing histogram metric. [PR-1134](https://github.com/nginxinc/nginx-gateway-fabric/pull/1134)
- Set Service address in Gateway Status. [PR-1141](https://github.com/nginxinc/nginx-gateway-fabric/pull/1141)

BUG FIXES:

- Optimize default NGINX config. [PR-1040](https://github.com/nginxinc/nginx-gateway-fabric/pull/1040)
- Ensure NGINX reload occurs. [PR-1033](https://github.com/nginxinc/nginx-gateway-fabric/pull/1033)
- Fix failure to recover if conf files are unexpectedly removed. [PR-1132](https://github.com/nginxinc/nginx-gateway-fabric/pull/1132)
- Only update a resource's status if it has changed. [PR-1151](https://github.com/nginxinc/nginx-gateway-fabric/pull/1151)

DOCUMENTATION:

- Non-functional testing guides and results. [Link](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/tests)

COMPATIBILITY:

- The Gateway API version: `0.8.1`
- NGINX version: `1.25.2`
- Kubernetes version: `1.23+`

CONTAINER IMAGES:

- Control plane: `ghcr.io/nginxinc/nginx-gateway-fabric:1.0.0`
- Data plane: `ghcr.io/nginxinc/nginx-gateway-fabric/nginx:1.0.0`

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
