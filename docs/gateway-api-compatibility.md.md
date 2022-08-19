# Gateway API Compatibility

This document describes which Gateway API resources NGINX Kubernetes Gateway supports and the extent of that support.

## Summary

| Resource | Support Status |
|-|-|
| [GatewayClass](#gatewayclass) | Partially supported |
| [Gateway](#gateway) | Partially supported |
| [HTTPRoute](#httproute) | Partially supported |
| [TLSRoute](#tlsroute) | Not supported |
| [TCPRoute](#tcproute) | Not supported |
| [UDPRoute](#udproute) | Not supported |
| [ReferenceGrant](#referencegrant) |  Not supported |
| [Custom policies](#custom-policies) | Not supported |

## Terminology

The support status is described using the following words:
- *Supported*. The resource or field is fully supported and conformant to the Gateway API specification.
- *Partially supported*. The resource or field is supported partially or with limitations. It will become fully supported in future releases.
- *Not supported*. The resource or field is not yet supported. It will become partially or fully supported in the future releases.

Note: it might be possible that NGINX Kubernetes Gateway will never support some resources and/or fields of the Gateway API. If that happens, that will be documented.

## Resources

Below we list the resources and the support status of their corresponding fields. 

Note that we don't describe the meaning of each field. For that, please visit the [Gateway API documentation](https://gateway-api.sigs.k8s.io/references/spec/). 

### GatewayClass 

> Status: Partially supported. 

NGINX Kubernetes Gateway supports only a single GatewayClass resource configured via `--gatewayclass` [cli argument](./cli-args.md).

Fields:
* `spec`
	* `controllerName` - supported.
	* `parametersRef` - not supported.
	* `description` - supported.
* `status` - partially supported.

### Gateway

> Status: Partially supported.

NGINX Kubernetes Gateway supports only a single Gateway resource: it select the Gateway resource among all resources that belong to
NGINX Kubernetes Gateway (i.e. reference NGINX Kubernetes Gateway corresponding GatewayClass). In case of multiple Gateway resources, NGINX Kubernetes Gateway will use a
deterministic conflict resolution strategy: it will choose the oldest resource based on the creation timestamp. If the timestamps are equal, NGINX Kubernetes Gateway will chose the resource that appears alphabetically first based on its namespace and then name. We might support multiple Gateway resources. If you're interested in that support, please share your use case with us.

Fields:
* `spec`
	* `gatewayClassName` - supported.
	* `listeners`
		* `name` - supported.
		* `hostname` - partially supported. Wildcard hostnames like `*.example.com` are not yet supported.
		* `port` - partially supported. Allowed values: `80` for HTTP listeners and `443` for HTTPS listeners.
		* `protocol` - partially supported. Allowed values: `HTTP`, `HTTPS`.
		* `tls`
		  * `mode` - partially supported. Allowed value: `Terminate`.
		  * `certificateRefs` - partially supported. The TLS certificate and key must be stored in a Secret resource of type `kubernetes.io/tls` in the same namespace as the Gateway resource. Only a single reference is supported. The Secret must be deployed before the Gateway resource. Secret rotation (watching for updates) is not supported.
		  * `options` - not supported.
		* `allowedRoutes` - not supported. 
	* `addresses` - not supported.
* `status` - partially supported.

### HTTPRoute

> Status: Partially supported.

Fields:
* `spec`
  * `parentRefs` - partially supported. `sectionName` must always be set. 
  * `hostnames` - partially supported. Wildcard binding is not supported: a hostname like `example.com` will not bind to a listener with the hostname `*.example.com`. However, `example.com` will bind to a listener with the empty hostname.
  * `rules`
	* `matches`
	  * `path` - partially supported. Only `PathPrefix` type.
	  * `headers` - partially supported. Only `Exact` type.
	  * `queryParams` - partially supported. Only `Exact` type. 
	  * `method` -  supported.
	* `filters` - not supported.
	* `backendRefs` - partially supported. Only a single backend ref. Backend ref `filters` are not supported. NGINX Kubernetes Gateway will use the IP of the Service as a backend, not the IPs of the corresponding Pods. Watching for Service updates is not supported.
* `status` -  partially supported.

### TLSRoute

> Status: Not supported.

### TCPRoute

> Status: Not supported.

### UDPRoute

> Status: Not supported.

### ReferenceGrant

> Status: Not supported.

### Custom Policies

> Status: Not supported.

Custom policies will be NGINX Kubernetes Gateway-specific CRDs that will allow supporting features like timeouts, load-balancing methods, authentication, etc - important data-plane features that are not part of the Gateway API spec.

While those CRDs are not part of the Gateway API, the mechanism of attaching them to Gateway API resources resources is part of the Gateway API. See the [Policy Attachment doc](https://gateway-api.sigs.k8s.io/references/policy-attachment/).