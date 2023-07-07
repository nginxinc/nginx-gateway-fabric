# Gateway API Compatibility

This document describes which Gateway API resources NGINX Kubernetes Gateway supports and the extent of that support.

## Summary

| Resource                            | Support Status      |
|-------------------------------------|---------------------|
| [GatewayClass](#gatewayclass)       | Partially supported |
| [Gateway](#gateway)                 | Partially supported |
| [HTTPRoute](#httproute)             | Partially supported |
| [TLSRoute](#tlsroute)               | Not supported       |
| [TCPRoute](#tcproute)               | Not supported       |
| [UDPRoute](#udproute)               | Not supported       |
| [ReferenceGrant](#referencegrant)   | Supported           |
| [Custom policies](#custom-policies) | Not supported       |

## Terminology

We use the following words to describe support status:

- *Supported*. The resource or field is fully supported and conformant to the Gateway API specification.
- *Partially supported*. The resource or field is supported partially or with limitations. It will become fully
  supported in future releases.
- *Not supported*. The resource or field is not yet supported. It will become partially or fully supported in future
  releases.

Note: it might be possible that NGINX Kubernetes Gateway will never support some resources and/or fields of the Gateway
API. We will document these decisions on a case by case basis.

## Resources

Below we list the resources and the support status of their corresponding fields.

For a description of each field, visit
the [Gateway API documentation](https://gateway-api.sigs.k8s.io/references/spec/).

### GatewayClass

> Status: Partially supported.

NGINX Kubernetes Gateway supports only a single GatewayClass resource configured via `--gatewayclass` flag
of the [static-mode](./cli-help.md#static-mode) command.

Fields:

* `spec`
    * `controllerName` - supported.
    * `parametersRef` - not supported.
    * `description` - supported.
* `status`
    * `conditions` - supported (Condition/Status/Reason):
        * `Accepted/True/Accepted`
        * `Accepted/False/InvalidParameters`
        * `Accepted/False/GatewayClassConflict`: Custom reason for when the GatewayClass references this controller, but
          a different GatewayClass name is provided to the controller via the command-line argument.

### Gateway

> Status: Partially supported.

NGINX Kubernetes Gateway supports only a single Gateway resource. The Gateway resource must reference NGINX Kubernetes
Gateway's corresponding GatewayClass.
See [static-mode](./cli-help.md#static-mode) command for more info.

Fields:

* `spec`
    * `gatewayClassName` - supported.
    * `listeners`
        * `name` - supported.
        * `hostname` - supported.
        * `port` - supported.
        * `protocol` - partially supported. Allowed values: `HTTP`, `HTTPS`.
        * `tls`
            * `mode` - partially supported. Allowed value: `Terminate`.
            * `certificateRefs` - The TLS certificate and key must be stored in a Secret resource of
              type `kubernetes.io/tls`. Only a single reference is supported. You must deploy the Secret before the
              Gateway resource.
            * `options` - not supported.
        * `allowedRoutes` - supported.
    * `addresses` - not supported.
* `status`
    * `addresses` - Pod IPAddress supported.
    * `conditions` - Supported (Condition/Status/Reason):
        * `Accepted/True/Accepted`
        * `Accepted/True/ListenersNotValid`
        * `Accepted/False/ListenersNotValid`
        * `Accepted/False/Invalid`
        * `Accepted/False/UnsupportedValue`: Custom reason for when a value of a field in a Gateway is invalid or not
          supported.
        * `Accepted/False/GatewayConflict`: Custom reason for when the Gateway is ignored due to a conflicting Gateway.
          NKG only supports a single Gateway.
        * `Programmed/True/Programmed`
        * `Programmed/False/Invalid`
        * `Programmed/False/GatewayConflict`: Custom reason for when the Gateway is ignored due to a conflicting
          Gateway. NKG only supports a single Gateway.
    * `listeners`
        * `name` - supported.
        * `supportedKinds` - not supported.
        * `attachedRoutes` - supported.
        * `conditions` - Supported (Condition/Status/Reason):
            * `Accepted/True/Accepted`
            * `Accepted/False/UnsupportedProtocol`
            * `Accepted/False/InvalidCertificateRef`
            * `Accepted/False/ProtocolConflict`
            * `Accepted/False/UnsupportedValue`: Custom reason for when a value of a field in a Listener is invalid or
              not supported.
            * `Accepted/False/GatewayConflict`: Custom reason for when the Gateway is ignored due to a conflicting
              Gateway. NKG only supports a single Gateway.
            * `Programmed/True/Programmed`
            * `Programmed/False/Invalid`
            * `ResolvedRefs/True/ResolvedRefs`
            * `ResolvedRefs/False/InvalidCertificateRef`
            * `ResolvedRefs/False/InvalidRouteKinds`
            * `Conflicted/True/ProtocolConflict`
            * `Conflicted/False/NoConflicts`

### HTTPRoute

> Status: Partially supported.

Fields:

* `spec`
    * `parentRefs` - partially supported. Port not supported.
    * `hostnames` - supported.
    * `rules`
        * `matches`
            * `path` - partially supported. Only `PathPrefix` and `Exact` types.
            * `headers` - partially supported. Only `Exact` type.
            * `queryParams` - partially supported. Only `Exact` type.
            * `method` - supported.
        * `filters`
            * `type` - supported.
            * `requestRedirect` - supported except for the experimental `path` field. If multiple filters
              with `requestRedirect` are configured, NGINX Kubernetes Gateway will choose the first one and ignore the
              rest.
            * `requestHeaderModifier` - supported. If multiple filters with `requestHeaderModifier` are configured,
              NGINX Kubernetes Gateway will choose the first one and ignore the rest.
            * `responseHeaderModifier`, `requestMirror`, `urlRewrite`, `extensionRef` - not supported.
        * `backendRefs` - partially supported. Backend ref `filters` are not supported.
* `status`
    * `parents`
        * `parentRef` - supported.
        * `controllerName` - supported.
        * `conditions` - partially supported. Supported (Condition/Status/Reason):
            * `Accepted/True/Accepted`
            * `Accepted/False/NoMatchingListenerHostname`
            * `Accepted/False/NoMatchingParent`
            * `Accepted/False/NotAllowedByListeners`
            * `Accepted/False/UnsupportedValue`: Custom reason for when the HTTPRoute includes an invalid or unsupported
              value.
            * `Accepted/False/InvalidListener`: Custom reason for when the HTTPRoute references an invalid listener.
            * `Accepted/False/GatewayNotProgrammed`: Custom reason for when the Gateway is not Programmed. HTTPRoute may
              be valid and configured, but will maintain this status as long as the Gateway is not Programmed.
            * `ResolvedRefs/True/ResolvedRefs`
            * `ResolvedRefs/False/InvalidKind`
            * `ResolvedRefs/False/RefNotPermitted`
            * `ResolvedRefs/False/BackendNotFound`
            * `ResolvedRefs/False/UnsupportedValue`: Custom reason for when one of the HTTPRoute rules has a backendRef
              with an unsupported value.

### TLSRoute

> Status: Not supported.

### TCPRoute

> Status: Not supported.

### UDPRoute

> Status: Not supported.

### ReferenceGrant

> Status: Supported.
> Support Level: Core

Fields:

* `spec`
    * `to`
        * `group` - supported.
        * `kind` - supports `Secret` and `Service`.
        * `name`- supported.
    * `from`
        * `group` - supported.
        * `kind` - supports `Gateway` and `HTTPRoute`.
        * `namespace`- supported.

### Custom Policies

> Status: Not supported.

Custom policies will be NGINX Kubernetes Gateway-specific CRDs that will allow supporting features like timeouts,
load-balancing methods, authentication, etc. - important data-plane features that are not part of the Gateway API spec.

While those CRDs are not part of the Gateway API, the mechanism of attaching them to Gateway API resources is part of
the Gateway API. See the [Policy Attachment doc](https://gateway-api.sigs.k8s.io/references/policy-attachment/).
