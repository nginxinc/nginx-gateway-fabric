[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/nginxinc/nginx-gateway-fabric/badge)](https://api.securityscorecards.dev/projects/github.com/nginxinc/nginx-gateway-fabric)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-gateway-fabric.svg?type=shield)](https://app.fossa.com/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-gateway-fabric?ref=badge_shield)

# NGINX Gateway Fabric

NGINX Gateway Fabric is an open-source project that provides an implementation of
the [Gateway API](https://gateway-api.sigs.k8s.io/) using [NGINX](https://nginx.org/) as the data plane. The goal of
this project is to implement the core Gateway APIs -- `Gateway`, `GatewayClass`, `HTTPRoute`, `TCPRoute`, `TLSRoute`,
and `UDPRoute` -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running
on Kubernetes. NGINX Gateway Fabric supports a subset of the Gateway API.

For a list of supported Gateway API resources and features, see
the [Gateway API Compatibility](docs/gateway-api-compatibility.md) doc.

Learn about our [design principles](/docs/developer/design-principles.md) and [architecture](/docs/architecture.md).

## Getting Started

1. [Quick Start on a kind cluster](docs/running-on-kind.md).
2. [Install](docs/installation.md) NGINX Gateway Fabric.
3. [Build](docs/building-the-images.md) an NGINX Gateway Fabric container image from source or use a pre-built image
   available
   on [GitHub Container Registry](https://github.com/nginxinc/nginx-gateway-fabric/pkgs/container/nginx-gateway-fabric).
4. Deploy various [examples](examples).
5. Read our [guides](/docs/guides).

## NGINX Gateway Fabric Releases

We publish NGINX Gateway Fabric releases on GitHub. See
our [releases page](https://github.com/nginxinc/nginx-gateway-fabric/releases).

The latest release is [1.0.0](https://github.com/nginxinc/nginx-gateway-fabric/releases/tag/v1.0.0).

The edge version is useful for experimenting with new features that are not yet published in a release. To use, choose
the *edge* version built from the [latest commit](https://github.com/nginxinc/nginx-gateway-fabric/commits/main)
from the main branch.

The table below summarizes the options regarding the images, manifests, documentation and examples and gives your links
to the correct versions:

| Version        | Description                              | Installation Manifests                                                            | Documentation and Examples                                                                                                                                             |
|----------------|------------------------------------------|-----------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Latest release | For production use                       | [Manifests](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.0.0/deploy). | [Documentation](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.0.0/docs). [Examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.0.0/examples). |
| Edge           | For experimental use and latest features | [Manifests](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/deploy).   | [Documentation](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/docs). [Examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples).     |

### Versioning

NGF uses semantic versioning for its releases. For more information, see https://semver.org.

> Major version zero `(0.Y.Z)` is reserved for development, anything MAY change at any time. The public API is not stable.

### Release Planning and Development

The features that will go into the next release are reflected in the
corresponding [milestone](https://github.com/nginxinc/nginx-gateway-fabric/milestones). Refer to
the [Issue Lifecycle](ISSUE_LIFECYCLE.md) document for information on issue creation and assignment to releases.


## Technical Specifications

The following table lists the software versions NGINX Gateway Fabric supports.

| NGINX Gateway Fabric | Gateway API | Kubernetes | NGINX OSS |
|----------------------|-------------|------------|-----------|
| Edge                 | 1.0.0       | 1.23+      | 1.25.3    |
| 1.0.0                | 0.8.1       | 1.23+      | 1.25.2    |
| 0.6.0                | 0.8.0       | 1.23+      | 1.25.2    |
| 0.5.0                | 0.7.1       | 1.21+      | 1.25.x *  |
| 0.4.0                | 0.7.1       | 1.21+      | 1.25.x *  |
| 0.3.0                | 0.6.2       | 1.21+      | 1.23.x *  |
| 0.2.0                | 0.5.1       | 1.21+      | 1.21.x *  |
| 0.1.0                | 0.5.0       | 1.19+      | 1.21.3    |

\*the installation manifests use the minor version of NGINX container image (e.g. 1.25) and the patch version is not
specified. This means that the latest available patch version is used.

## SBOM (Software Bill of Materials)

We generate SBOMs for the binaries and the Docker image.

### Binaries

The SBOMs for the binaries are available in the releases page. The SBOMs are generated
using [syft](https://github.com/anchore/syft) and are available in SPDX format.

### Docker Images

The SBOM for the Docker image is available in
the [GitHub Container](https://github.com/nginxinc/nginx-gateway-fabric/pkgs/container/nginx-gateway-fabric)
repository. The SBOM is generated using [syft](https://github.com/anchore/syft) and stored as an attestation in the
image manifest.

For example to retrieve the SBOM for `linux/amd64` and analyze it using [grype](https://github.com/anchore/grype) you
can run the following command:

```shell
docker buildx imagetools inspect ghcr.io/nginxinc/nginx-gateway-fabric:edge --format '{{ json (index .SBOM "linux/amd64").SPDX }}' | grype
```

## Troubleshooting

For troubleshooting help, see the [Troubleshooting](/docs/troubleshooting.md) document.

## Contacts

We’d like to hear your feedback! If you experience issues with our Gateway Controller, please [open a bug][bug] in
GitHub. If you have any suggestions or enhancement requests, please [open an idea][idea] on GitHub discussions. You can
contact us directly via kubernetes@nginx.com or on the [NGINX Community Slack][slack] in
the `#nginx-gateway-fabric`
channel.

[bug]:https://github.com/nginxinc/nginx-gateway-fabric/issues/new?assignees=&labels=&projects=&template=bug_report.md&title=

[idea]:https://github.com/nginxinc/nginx-gateway-fabric/discussions/categories/ideas

[slack]: https://nginxcommunity.slack.com/channels/nginx-gateway-fabric

## Contributing

Please read our [Contributing guide](CONTRIBUTING.md) if you'd like to contribute to the project.

## Support

NGINX Gateway Fabric is not covered by any support contract.


# Gateway API Compatibility

This document describes which Gateway API resources NGINX Gateway Fabric supports and the extent of that support.

## Summary

| Resource                            | Core Support Level | Extended Support Level | Implementation-Specific Support Level | API Version |
|-------------------------------------|--------------------|------------------------|---------------------------------------|-------------|
| [GatewayClass](#gatewayclass)       | Supported          | Not supported          | Not Supported                         | v1          |
| [Gateway](#gateway)                 | Supported          | Not supported          | Not Supported                         | v1          |
| [HTTPRoute](#httproute)             | Supported          | Partially supported    | Not Supported                         | v1          |
| [ReferenceGrant](#referencegrant)   | Supported          | N/A                    | Not Supported                         | v1beta1     |
| [Custom policies](#custom-policies) | Not supported      | N/A                    | Not Supported                         | N/A         |
| [TLSRoute](#tlsroute)               | Not supported      | Not supported          | Not Supported                         | N/A         |
| [TCPRoute](#tcproute)               | Not supported      | Not supported          | Not Supported                         | N/A         |
| [UDPRoute](#udproute)               | Not supported      | Not supported          | Not Supported                         | N/A         |

## Terminology

Gateway API features has three [support levels](https://gateway-api.sigs.k8s.io/concepts/conformance/#2-support-levels):
Core, Extended and Implementation-specific. We use the following terms to describe the support status for each level and
resource field:

- *Supported*. The resource or field is fully supported.
- *Partially supported*. The resource or field is supported partially or with limitations. It will become fully
  supported in future releases.
- *Not supported*. The resource or field is not yet supported. It will become partially or fully supported in future
  releases.

> Note: it might be possible that NGINX Gateway Fabric will never support some resources
> and/or fields of the Gateway API. We will document these decisions on a case by case basis.
>
> NGINX Gateway Fabric doesn't support any features from the experimental release channel.

## Resources

Below we list the resources and the support status of their corresponding fields.

For a description of each field, visit
the [Gateway API documentation](https://gateway-api.sigs.k8s.io/references/spec/).

### GatewayClass

> Support Levels:
>
> - Core: Supported.
> - Extended: Not supported.
> - Implementation-specific: Not supported.

NGINX Gateway Fabric supports only a single GatewayClass resource configured via `--gatewayclass` flag of
the [static-mode](./cli-help.md#static-mode) command.

Fields:

- `spec`
  - `controllerName` - supported.
  - `parametersRef` - not supported.
  - `description` - supported.
- `status`
  - `conditions` - supported (Condition/Status/Reason):
    - `Accepted/True/Accepted`
    - `Accepted/False/InvalidParameters`
    - `Accepted/False/GatewayClassConflict`: Custom reason for when the GatewayClass references this controller, but
          a different GatewayClass name is provided to the controller via the command-line argument.

### Gateway

> Support Levels:
>
> - Core: Supported.
> - Extended: Partially supported.
> - Implementation-specific: Not supported.

NGINX Gateway Fabric supports only a single Gateway resource. The Gateway resource must reference NGINX Gateway
Fabric's corresponding GatewayClass. See [static-mode](./cli-help.md#static-mode) command for more info.

Fields:

- `spec`
  - `gatewayClassName` - supported.
  - `listeners`
    - `name` - supported.
    - `hostname` - supported.
    - `port` - supported.
    - `protocol` - partially supported. Allowed values: `HTTP`, `HTTPS`.
    - `tls`
      - `mode` - partially supported. Allowed value: `Terminate`.
      - `certificateRefs` - The TLS certificate and key must be stored in a Secret resource of
              type `kubernetes.io/tls`. Only a single reference is supported.
      - `options` - not supported.
    - `allowedRoutes` - supported.
  - `addresses` - not supported.
- `status`
  - `addresses` - partially supported. LoadBalancer and Pod IP.
  - `conditions` - supported (Condition/Status/Reason):
    - `Accepted/True/Accepted`
    - `Accepted/True/ListenersNotValid`
    - `Accepted/False/ListenersNotValid`
    - `Accepted/False/Invalid`
    - `Accepted/False/UnsupportedValue`- custom reason for when a value of a field in a Gateway is invalid or not
          supported.
    - `Accepted/False/GatewayConflict`- custom reason for when the Gateway is ignored due to a conflicting Gateway.
          NGF only supports a single Gateway.
    - `Programmed/True/Programmed`
    - `Programmed/False/Invalid`
    - `Programmed/False/GatewayConflict`- custom reason for when the Gateway is ignored due to a conflicting
          Gateway. NGF only supports a single Gateway.
  - `listeners`
    - `name` - supported.
    - `supportedKinds` - supported.
    - `attachedRoutes` - supported.
    - `conditions` - supported (Condition/Status/Reason):
      - `Accepted/True/Accepted`
      - `Accepted/False/UnsupportedProtocol`
      - `Accepted/False/InvalidCertificateRef`
      - `Accepted/False/ProtocolConflict`
      - `Accepted/False/UnsupportedValue`- custom reason for when a value of a field in a Listener is invalid or
              not supported.
      - `Accepted/False/GatewayConflict` - custom reason for when the Gateway is ignored due to a conflicting
              Gateway. NGF only supports a single Gateway.
      - `Programmed/True/Programmed`
      - `Programmed/False/Invalid`
      - `ResolvedRefs/True/ResolvedRefs`
      - `ResolvedRefs/False/InvalidCertificateRef`
      - `ResolvedRefs/False/InvalidRouteKinds`
      - `Conflicted/True/ProtocolConflict`
      - `Conflicted/False/NoConflicts`

### HTTPRoute

> Support Levels:
>
> - Core: Supported.
> - Extended: Partially supported.
> - Implementation-specific: Not supported.

Fields:

- `spec`
  - `parentRefs` - partially supported. Port not supported.
  - `hostnames` - supported.
  - `rules`
    - `matches`
      - `path` - partially supported. Only `PathPrefix` and `Exact` types.
      - `headers` - partially supported. Only `Exact` type.
      - `queryParams` - partially supported. Only `Exact` type.
      - `method` - supported.
    - `filters`
      - `type` - supported.
      - `requestRedirect` - supported except for the experimental `path` field. If multiple filters
              with `requestRedirect` are configured, NGINX Gateway Fabric will choose the first one and ignore the
              rest.
      - `requestHeaderModifier` - supported. If multiple filters with `requestHeaderModifier` are configured,
              NGINX Gateway Fabric will choose the first one and ignore the rest.
      - `responseHeaderModifier`, `requestMirror`, `urlRewrite`, `extensionRef` - not supported.
    - `backendRefs` - partially supported. Backend ref `filters` are not supported.
- `status`
  - `parents`
    - `parentRef` - supported.
    - `controllerName` - supported.
    - `conditions` - partially supported. Supported (Condition/Status/Reason):
      - `Accepted/True/Accepted`
      - `Accepted/False/NoMatchingListenerHostname`
      - `Accepted/False/NoMatchingParent`
      - `Accepted/False/NotAllowedByListeners`
      - `Accepted/False/UnsupportedValue` - custom reason for when the HTTPRoute includes an invalid or
              unsupported value.
      - `Accepted/False/InvalidListener` - custom reason for when the HTTPRoute references an invalid listener.
      - `Accepted/False/GatewayNotProgrammed` - custom reason for when the Gateway is not Programmed. HTTPRoute
              may be valid and configured, but will maintain this status as long as the Gateway is not Programmed.
      - `ResolvedRefs/True/ResolvedRefs`
      - `ResolvedRefs/False/InvalidKind`
      - `ResolvedRefs/False/RefNotPermitted`
      - `ResolvedRefs/False/BackendNotFound`
      - `ResolvedRefs/False/UnsupportedValue` - custom reason for when one of the HTTPRoute rules has a backendRef
              with an unsupported value.
      - `PartiallyInvalid/True/UnsupportedValue`

### ReferenceGrant

> Support Levels:
>
> - Core: Supported.
> - Extended: N/A.
> - Implementation-specific: N/A

Fields:

- `spec`
  - `to`
    - `group` - supported.
    - `kind` - supports `Secret` and `Service`.
    - `name`- supported.
  - `from`
    - `group` - supported.
    - `kind` - supports `Gateway` and `HTTPRoute`.
    - `namespace`- supported.

### TLSRoute

> Status: Not supported.

### TCPRoute

> Status: Not supported.

### UDPRoute

> Status: Not supported.

### Custom Policies

> Status: Not supported.

Custom policies will be NGINX Gateway Fabric-specific CRDs that will allow supporting features like timeouts,
load-balancing methods, authentication, etc. - important data-plane features that are not part of the Gateway API spec.

While those CRDs are not part of the Gateway API, the mechanism of attaching them to Gateway API resources is part of
the Gateway API. See the [Policy Attachment doc](https://gateway-api.sigs.k8s.io/references/policy-attachment/).

