
# Enhancement Proposal #1411: Extensions for NGINX Features

- Issue: #1566
- Status: Provisional

## Summary

NGINX is highly configurable and offers rich features that can benefit our users. We want to expose this native NGINX configuration to our users through Gateway API extension points -- such as Policies and Filters. This Enhancement Proposal aims to identify the set of NGINX directives and parameters we will expose first, group them according to Gateway API role(s), NGINX contexts, and use cases, and propose the type of extension point for each group.

## Table of Contents

<!-- TOC -->
* [Enhancement Proposal #1411: Extensions for NGINX Features](#enhancement-proposal-1411--extensions-for-nginx-features)
  * [Summary](#summary)
  * [Table of Contents](#table-of-contents)
  * [Goals](#goals)
  * [Non-Goals](#non-goals)
  * [Gateway API Extension](#gateway-api-extension)
    * [GatewayClass Parameters Ref](#gatewayclass-parameters-ref)
      * [Issues with `parametersRef`](#issues-with-parametersref)
    * [Infrastructure API](#infrastructure-api)
    * [TLS Options](#tls-options)
    * [Filters](#filters)
    * [BackendRef](#backendref)
    * [Policy](#policy)
      * [Direct Policy Attachment](#direct-policy-attachment)
      * [Inherited Policy Attachment](#inherited-policy-attachment)
        * [Hierarchy](#hierarchy)
      * [Challenges of Policy Attachment](#challenges-of-policy-attachment)
  * [Prioritized NGINX Features](#prioritized-nginx-features)
    * [High-Priority Features](#high-priority-features)
    * [Medium-Priority Features](#medium-priority-features)
    * [Low-Priority Features](#low-priority-features)
    * [Features with Active Gateway API Enhancement Proposals (GEPs)](#features-with-active-gateway-api-enhancement-proposals--geps-)
  * [Grouping the Features](#grouping-the-features)
  * [API](#api)
    * [Gateway Settings](#gateway-settings)
      * [Future Work](#future-work)
      * [Alternatives](#alternatives)
    * [Response Modification](#response-modification)
      * [Future Work](#future-work-1)
      * [Alternatives](#alternatives-1)
    * [TLS Settings](#tls-settings)
      * [Future Work](#future-work-2)
      * [Alternatives](#alternatives-2)
    * [Client Settings](#client-settings)
      * [Future Work](#future-work-3)
      * [Alternatives](#alternatives-3)
    * [Upstream Settings](#upstream-settings)
      * [Alternatives](#alternatives-4)
    * [Authentication](#authentication)
      * [Future Work](#future-work-4)
      * [Alternatives](#alternatives-5)
    * [Observability](#observability)
      * [Future Work](#future-work-5)
      * [Alternatives](#alternatives-6)
    * [Proxy Settings](#proxy-settings)
      * [Future Work](#future-work-6)
      * [Alternatives](#alternatives-7)
    * [Circuit Breaker/ Backup service](#circuit-breaker-backup-service)
      * [Alternatives](#alternatives-8)
  * [Testing](#testing)
  * [Security Considerations](#security-considerations)
  * [Alternatives Considered](#alternatives-considered)
  * [References](#references)
<!-- TOC -->als

- Identify the set of NGINX features to deliver to users.
- Group these features to reduce the number of APIs and improve user experience.
- For each group, identify the type of Gateway API extension to use and the applicable Gateway API role(s).

## Non-Goals

- Design the API of every extension. This design work will be completed for each extension before implementation.
- Design an API that will allow the user to insert raw NGINX config into the NGINX config that NGINX Gateway Fabric generates (e.g. [snippets in NGINX Ingress Controller](https://docs.nginx.com/nginx-ingress-controller/configuration/ingress-resources/advanced-configuration-with-snippets/)).

## Gateway API Extension

The Gateway API provides many extension that implementations can leverage to deliver features that the general-purpose API cannot address. This section provides a summary of these extension.

### GatewayClass Parameters Ref

_Role(s)_: Infrastructure Provider

_Resource(s)_: GatewayClass

_Status_: Completed

_Channel_: Standard

_Conformance Level_: Implementation-specific

_Example(s)_: [EnvoyProxy CRD](https://gateway.envoyproxy.io/v0.6.0/user/customize-envoyproxy/)

First-class API field on the `GatewayClass.spec`. The field refers to a resource containing the configuration parameters corresponding to the GatewayClass. The resource referenced can be a standard Kubernetes resource, a CRD, or a ConfigMap. While GatewayClasses are cluster-scoped, resources referenced by `GatewayClass.spec.parametersRef` can be namespaced-scoped or cluster-scoped. The configuration parameters contained in the referenced resource are applied to all Gateways attached to the GatewayClass.

Example:

```yaml
kind: GatewayClass
metadata:
  name: internet
spec:
  controllerName: "example.net/gateway-controller"
  parametersRef:
    group: example.net/v1alpha1
    kind: Config
    name: internet-gateway-config
---
apiVersion: example.net/v1alpha1
kind: Config
metadata:
  name: internet-gateway-config
spec:
  ip-address-pool: internet-vips
```

#### Issues with `parametersRef`

[GEP-1867](https://gateway-api.sigs.k8s.io/geps/gep-1867/) raises the following operational challenges with `parameterRefs`:

1. _Permissions_: To make declarative changes to a Gateway, the Gateway owner (who has RBAC permissions to a specific Gateway) must have access to GatewayClass, a cluster-scoped resource.
2. _Scope_: If a change is made on a GatewayClass, _all_ Gateways are affected by that change. This will become an issue once we [support multiple Gateways](https://github.com/nginxinc/nginx-gateway-fabric/issues/1443).
3. _Dynamic Changes_: GatewayClasses are meant to be templates, so changes made to the GatewayClass are not expected to change deployed Gateways. This means the configuration is not dynamic.

### Infrastructure API

_Role(s)_: Cluster Operator, and Infrastructure Provider

_Resource(s)_: Gateway, GatewayClass (planned)

_Status_: Experimental

_Channel_: Experimental

_Conformance Level_: Core

First-class infrastructure API on Gateway and GatewayClass. Several GEPs are related to this API: [GEP-1867](https://gateway-api.sigs.k8s.io/geps/gep-1867/), [GEP-1762](https://gateway-api.sigs.k8s.io/geps/gep-1762/), and [GEP-1651](https://gateway-api.sigs.k8s.io/geps/gep-1651/).

Gateways represent a piece of infrastructure that often needs vendor-specific config (e.g., size or version of the infrastructure to provide). The goal of the infrastructure API is to provide a way to define both implementation-specific and standard attributes on a _specific_ Gateway.

So far, the infrastructure API has been implemented on the Gateway, and there are only two fields: `annotations` and `labels`:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: my-gateway
spec:
  infrastructure:
    labels:
      foo: bar
    annotations:
      name: my-annotation
```

Infrastructure labels and annotations should be applied to any resources created in response to the Gateway. This only applies to _automated deployments_ (i.e., provisioner mode), implementations that automatically deploy the data plane based on a Gateway.
Other use cases for this API are Service type, Service IP, CPU memory requests, affinity rules, and Gateway routability (public, private, and cluster).

There are plans to add a `parametersRef` field to the infrastructure API on the Gateway. This field would function similarly to the GatewayClass `parametersRef` but would not have the same issues highlighted in the section above.

### TLS Options

_Role(s)_: Cluster Operator

_Resource(s)_: Gateway

_Status_: Completed

_Channel_: Standard

_Conformance Level_: Implementation-specific

_Example(s)_: [GKE pre-shared certs](https://cloud.google.com/kubernetes-engine/docs/how-to/gatewayclass-capabilities#spec-listeners-tls-options)

TLS options are a list of key/value pairs to enable extended TLS configuration for an implementation. This field is part of the [GatewayTLSConfig](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io%2fv1.GatewayTLSConfig) defined on a Gateway Listener. Currently, there are no standard keys for TLS options, but the API may define a set of standard keys in the future.

Possible use cases: minimum TLS version and supported cipher suites.

Example:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: my-gateway
spec:
  listeners:
  - name: https
    port: 443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: my-secret
        namespace: certificate
      options:
        example.com/my-custom-option: custom-value
```

### Filters

_Role(s)_: Application Developer

_Resource(s)_: HTTPRoute, GRPCRoute

_Status_: Completed

_Channel_: Standard

_Conformance Level_: Implementation-specific

_Example(s)_: [Easegress Filters](https://megaease.com/blog/2023/12/05/enhancing-k8s-gateway-api-with-easegress-without-changing-a-single-line-of-code/).

Filters define processing steps that must be completed during the request or response lifecycle. They can be applied on the `route.rule` or the `route.rule.backendRef`. If applied on the `backendRef`, the filter should be executed if and only if the request is being forwarded.

Possible use cases: request/response modification, authentication, rate-limiting, and traffic shaping.

Example:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: custom
spec:
  hostnames:
  - "example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    filters:
    - type: ExtensionRef
      extensionRef:
        group: my.group.io
        kind: MyCustomFilter
        name: example
    backendRefs:
    - name: backend
      port: 80
---
apiVersion: example.com/v1alpha1
kind: MyCustomFilter
metadata:
  name: example
spec:
  // some values here that are applied during the request/response lifecycle
```

### BackendRef

_Role(s)_: Application Developer

_Resource(s)_: xRoute

_Status_: Completed

_Channel_: Standard

_Conformance Level_: Implementation-specific

_Example(s)_: [ServiceImport GKE](https://cloud.google.com/kubernetes-engine/docs/how-to/gatewayclass-capabilities#spec-rules-backendrefs), [ServiceImport Envoy Gateway](https://gateway.envoyproxy.io/v0.6.0/user/multicluster-service/).

BackendRefs defines the backend(s) where matching requests should be sent. The Gateway API supports BackendRefs of type Kubernetes Service. An implementation can add support for other types of backends. This extension point should be used for forwarding traffic to network endpoints other than Kubernetes Services.

Possible use cases: S3 bucket, lambda function, file-server.

Example:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: custom-backend
spec:
  hostnames:
  - "example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: my.group.io
      kind: MyKind
      name: backend
      port: 80
```

### Policy

_Role(s)_: All roles

_Resource(s)_: All resources

_Status_: Experimental

_Channel_: Experimental

_Conformance Level_: Extended/Implementation-specific

_Example(s)_: [Envoy Gateway BackendTrafficPolicy](https://gateway.envoyproxy.io/v0.6.0/api/extension_types/#backendtrafficpolicy), [GKE HealthCheckPolicy](https://cloud.google.com/kubernetes-engine/docs/how-to/configure-gateway-resources), [BackendTLSPolicy](https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/)

Policies are a Kubernetes object that augments the behavior of an object in a standard way. Policies can be attached to one object ("Direct Policy Attachment") or objects in a hierarchy ("Inherited Policy Attachment"). In both cases, Policies are implemented as CRDs, and they must include a `TargetRef` struct in the `spec` to identify how and where to apply the Policy. Policies can use the `sectioName` field of the `TargetRef` to target specific matches within nested objects, such as a Listener within a Gateway or a specific Port on a Service. There's also an open [GEP issue](https://github.com/kubernetes-sigs/gateway-api/issues/995) to a name field to HTTPRouteRule and HTTPRouteMatch structs to allow users to identify different routes. Once implemented, Policies can target a specific rule or match in an HTTPRoute.

Policies do not need to support attaching to all resource Kinds. Implementations can choose which resources the Policy can be attached to.

#### Direct Policy Attachment

A Direct Policy Attachment is a Policy that references a single object -- such as a Gateway or HTTPRoute. It is tightly bound to one instance of a particular Kind within a single Namespace or an instance of a single Kind at the cluster-scope. It only modifies the behavior of the object that it's bound to.

Example:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: BackendTLSPolicy
metadata:
  name: backend-tls
spec:
  targetRef:
    group: '' #
    kind: Service
    name: secure-app
    namespace: default
  tls:
    caCertRefs:
    - name: backend-cert
      group: '' # Empty string means core - this is a standard convention
      kind: ConfigMap
    hostname: secure-app.example.com
```

This BackendTLSPolicy is an example of a Direct Policy Attachment. It targets the `secure-app` Service in the `default` Namespace. The TLS configuration in the `backend-tls` Policy will be applied for all Routes that reference the `secure-app` Service.

When to use Direct Policy Attachment:

- The number or scope of objects that need to be modified is _limited_ or _singular_.
- No transitive information. The modifications only affect the single object that the Policy is bound to.
- The status should be reasonably easy to set in 1-2 Conditions.
- SHOULD only be used to target objects in the same Namespace as the Policy.

Direct Policy Attachment is simple. A Policy references the resource it wants to apply to. Access is granted with RBAC; anyone who has access to the Policy in the given Namespace can attach it to any resource within that Namespace.

The Policy defines which resource Kinds it can be attached to.

#### Inherited Policy Attachment

Inherited Policy Attachment is designed to allow settings to flow down a hierarchy. Inherited Policies must have the following fields in their specs:

- `defaults`: set the default value. Can be overridden by _lower_ objects.
  Ex. A connection timeout default policy on a Gateway may be overridden by a connection timeout policy on an HTTPRoute.
- `overrides`: cannot be overridden by _lower_ objects.
  Ex. Setting a max client timeout to some non-infinite value at the Gateway to stop HTTPRoute owners from leaking connections over time.

When to use an Inherited Policy:

- The settings are bound to one object but affect other objects attached to it. For example, it affects HTTPRoutes attached to a single Gateway.
- The settings need to be able to be defaulted but can be overridden on a per-object basis.
- The settings must be enforced by one persona and not modifiable or removable by a lesser-privileged persona.
- Reporting (in terms of status), if the Policy is attached, is easy, but reporting what resources the Policy is being applied to is not and will require careful design.

Example:

```yaml
kind: CDNCachingPolicy
metadata:
  name: gw-policy
spec:
  override:
    cdn:
      enabled: true
  default:
    cdn:
      cachePolicy:
        includeHost: true
        includeProtocol: true
        includeQueryString: true
  targetRef:
    kind: Gateway
    name: example
---
kind: CDNCachingPolicy
metadata:
  name: route-policy
spec:
  default:
    cdn:
      cachePolicy:
        includeQueryString: false
  targetRef:
    kind: HTTPRoute
    name: example
```

The Policy attached to the Gateway has the following effects:

- It _requires_ cdn to be enabled for all routes connected to the Gateway. This setting can not be overridden by the CDNCachingPolicy attached to the HTTPRoute.
- It provides a default configuration for the cdn. These settings _can_ be overridden by the CDNCachingPolicy attached to the HTTPRoute.

The policy attached to the HTTPRoute changes the value of `cdn.cachePolicy.includeQueryString`. All other default cdn configuration set on the Gateway policy remain unchanged. The effective policy on the HTTPRoute is:

```yaml
cdn:
  enabled: true
  cachePolicy:
    includeHost: true
    includeProtocol: true
    includeQueryString: false
```

##### Hierarchy

The following image depicts the hierarchy of `defaults` and `overrides`:

![hierarchy](/docs/images/hierarchy.png)

Override values are given precedence from the top down. An override value attached to a GatewayClass will have the highest precedence overall. A lower object cannot override it. This provides users with a mechanism to enforce/require settings.

Default values are given precedence from the bottom up. A default attached to a Backend will have the highest precedence among _default_ values.

Inherited Policies do not need to support policy attachment to each resource shown in the hierarchy. Implementations can choose which resources the Policy can attach to.

#### Challenges of Policy Attachment

_Status and the Discoverability problem_

Policy Attachment adds a new persona -- Policy Admin, whose responsibilities cut across all levels of the Gateway API. The Policy Admins actions need to be discoverable by all other Gateway API personas. For example, the Application Developer needs to be able to look at their routes and know what Policy settings are applied to them. The Cluster Operator needs to be able to determine the set of objects affected by a Policy so they can diagnose and fix cluster-wide issues. Finally, the Policy Admin needs to be able to validate their Policies and understand where the Policies are applied and what the effects are.

Ideally, implementations can write the complete resultant set of Policy available in the status of objects that the Policy affects. Unfortunately, this introduces a new set of challenges around writing status, which include status getting too large, the difficulty of building a common data representation for Policy, and the fan-out problem where updating one Policy object could trigger a waterfall of status updates for all affected objects.

[The Metaresources and Policy Attachment GEP](https://gateway-api.sigs.k8s.io/geps/gep-713/#solution-cookbook) proposes a range of solutions that help address this problem but does not eliminate it.

_Hard to implement_

Policies are complex to implement because they can affect many objects in various ways -- i.e., all objects in a Namespace or all Routes attached to a Gateway. Inherited Policies introduce more complexity through the hierarchical nature of overrides and defaults. Computing the effective set of policy for a given object can be computationally challenging. In addition, we will need to define how different Policy types interact.

_Subject to change_

The Metaresources and Policy Attachment GEP is experimental. There's no guarantee that Policy Attachment will stay the same as the Gateway API receives feedback from implementors on the feasibility of Policies.

## Prioritized NGINX Features

To identify the set of NGINX directives and parameters NGINX Gateway Fabric should implement first, we first identified the highest priority features using the following criteria:

1. Customer requested features.
2. Most popular features of the NGINX Ingress Controller.
3. Features without active Gateway Enhancement Proposals (GEPs).

From this list of features, we prioritized them into four categories: high-priority, medium-priority, low-priority, and active GEPs.

### High-Priority Features

| Features                                                                                         | Requires NGINX Plus |
|--------------------------------------------------------------------------------------------------|---------------------|
| Log level and format                                                                             |                     |
| Passive health checks                                                                            |                     |
| Slow start. Prevent a recently recovered server from being overwhelmed on connections by restart | X                   |
| Active health checks                                                                             | X                   |
| JSON Web Token (JWT) authentication                                                              | X                   |
| OpenID Connect (OIDC) authentication. Allow users to authenticate with single-sign-on (SSO)      | X                   |
| Customize client settings. For example, `client_max_body_size`                                   |                     |
| Upstream keepalives                                                                              |                     |
| Client keepalives                                                                                |                     |
| TLS settings. For example, TLS protocols and server ciphers                                      |                     |
| OpenTelemetry tracing                                                                            |                     |

### Medium-Priority Features

| Features                                                                               | Requires NGINX Plus |
|----------------------------------------------------------------------------------------|---------------------|
| Circuit breaker                                                                        |                     |
| Load-balancing method                                                                  |                     |
| Load-balancing method (least time)                                                     | X                   |
| Limit connections to a server                                                          |                     |
| Request queuing. Queue requests if servers are unavailable                             | X                   |
| Basic authentication. User/Password                                                    |                     |
| PROXY protocol                                                                         |                     |
| Upstream zone size                                                                     |                     |
| Custom response. Return a custom response for a given path                             |                     |
| External routing. Route to services outside the cluster                                |                     |
| HTTP Strict Transport Security (HSTS)                                                  |                     |
| API Key authentication                                                                 |                     |
| Next upstream retries. NGINX retries by trying the _next_ server in the upstream group |                     |
| Proxy Buffering                                                                        |                     |
| Connection timeouts                                                                    |                     |

### Low-Priority Features

| Features                                   | Requires NGINX Plus |
|--------------------------------------------|---------------------|
| Pass/hide headers to/from the client       |                     |
| Custom error pages                         |                     |
| Access control                             |                     |
| Windows Technology LAN Manager (NTLM)      | X                   |
| Backend client certificate verification    |                     |
| NGINX worker settings                      |                     |
| HTTP/2 client requests                     |                     |
| Set server name size                       |                     |
| JSON Web Token (JWT) content-based routing | X                   |
| QUIC/HTTP3                                 |                     |
| Remote authentication request              |                     |
| Timeout for unresponsive clients           |                     |

### Features with Active Gateway API Enhancement Proposals (GEPs)

The status field in the table describes the status of the GEP using the following terms:

- Experimental: GEP is implemented and in the experimental channel.
- Provisional: GEP is written but has yet to be implemented.
- Open: An issue exists, but a GEP still needs to be written.

| Functionality                                 | Status       | GEP/Issue                                                                |
|-----------------------------------------------|--------------|--------------------------------------------------------------------------|
| Per-request timeouts                          | Experimental | [GEP-1742](https://gateway-api.sigs.k8s.io/geps/gep-1742/)               |
| Retries. This is different from NGINX retries | Open         | [GEP-1731](https://github.com/kubernetes-sigs/gateway-api/issues/1731)   |
| Session persistence                           | Provisional  | [GEP-1619](https://gateway-api.sigs.k8s.io/geps/gep-1619/)               |
| Rate-limiting                                 | Open         | [Issue 326](https://github.com/kubernetes-sigs/gateway-api/issues/326)   |
| Cross-Origin Resource Sharing (CORS)          | Open         | [Issue 1767](https://github.com/kubernetes-sigs/gateway-api/issues/1767) |
| Gateway client certificate verification       | Provisional  | [GEP-91](https://gateway-api.sigs.k8s.io/geps/gep-91/)                   |
| Authentication                                | Open         | [Issue 1494](https://github.com/kubernetes-sigs/gateway-api/issues/1494) |

## Grouping the Features

To reduce the number of CRDs NGINX Gateway Fabric must maintain and users have to create, we grouped the high and medium-priority features into configuration categories. Once the features are grouped, we can begin implementation by focusing on the high-priority features of each group. Then, once the high-priority features are complete, we can move on to the medium-priority features. The low-priority features are out of scope for the rest of the Enhancement Proposal but may be revisited once we make progress on the higher-priority features. In addition, most features with active GEPs are not included in the groups as we want to help move the GEPs forward toward standardization instead of creating our bespoke solutions.

When grouping the features, we considered the following factors:

1. Use cases. Are there features that can be grouped by the use case they satisfy? For example, authentication or observability.
2. Gateway API roles. Which features are the domain of the Cluster Operator? Which features are Application Developers responsible for?
3. The NGINX contexts. Which contexts are the NGINX directives available in? For example, http, server, location, upstream, etc.

The following picture shows the nine groups we came up with:

![groups](/docs/images/nginx-functionality-groups.png)

## API

The following section proposes an extension type (Policy, Filter, etc.), extension point (GatewayClass, HTTPRoute, etc.), and Gateway API role for all the groups of features. The goal of this Enhancement Proposal is not to design each extension but to propose a strategy for delivering critical NGINX features to our customers.

### Gateway Settings

_Extension type:_ ParametersRef

_Resource type:_ CRD

_Role(s):_ Cluster Operator

_Attaches to:_ GatewayClass or Gateway

_NGINX Context(s):_ main, http, stream

Features:

- Error log level
- Access log settings: format and disable
- PROXY protocol
- OTEL Tracing: global settings such as exporter configuration and global enable/disable.
- External Routing: enable/disable routing to ExternalName Services and set DNS resolver addresses.

NGINX directives:

- [`proxy_protocol`](https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html#proxy_protocol)
- [`access_log`](https://nginx.org/en/docs/http/ngx_http_log_module.html#access_log)
- [`error_log`](https://nginx.org/en/docs/ngx_core_module.html#error_log)
- [`log_format`](https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format)
- [`otel_exporter`](https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter)
- [`otel_service_name`](https://nginx.org/en/docs/ngx_otel_module.html#otel_service_name)
- [`otel_trace`](https://nginx.org/en/docs/ngx_otel_module.html#otel_trace)
- [`resolver`](https://nginx.org/en/docs/http/ngx_http_core_module.html#resolver)


These features are grouped because they are all the responsibility of the Cluster Operator and should not be set or changed by Application Developers.

Currently, `parametersRef` is only available on the GatewayClass, which means all Gateways attached to the GatewayClass will inherit these settings. This is acceptable since NGINX Gateway Fabric supports a single Gateway per GatewayClass. However, once we [support multiple Gateways](https://github.com/nginxinc/nginx-gateway-fabric/issues/144) and `parametersRef` is added to the Gateway Infrastructure API, then we will want to support referencing this CRD on individual Gateways.

#### Future Work

Add support for:

- Stream access log settings: format and disable
- Worker settings: number of worker processes, connections, etc.
- Resolver settings: timeout, ipv6 on/off, cache timeout, etc.

#### Alternatives

- ParametersRef with ConfigMap: A ConfigMap is another resource type where a user can provide configuration options. However, unlike CRDs, ConfigMaps do not have built-in schema validation, versioning, or conversion webhooks.
- Direct Policy: A Direct Policy may also work for Gateway Settings. It can be attached to a Gateway and scoped to Cluster Operators through RBAC. However, `parametersRef` provides better visibility as it is directly referenced in the Gateway resource and will be easier to manage.

### Response Modification

_Extension type:_ Filter

_Resource type:_ CRD

_Role(s):_ Application Developer

_Attaches to:_ HTTPRoute

_NGINX Context(s):_ location

Features:

- Custom Responses. Return a preconfigured response.

NGINX directives:

- [`return`](https://nginx.org/en/docs/http/ngx_http_rewrite_module.html#return)

#### Future Work

- We could contribute this filter to the Gateway API if it is portable among enough implementations.
- Allow attaching to GRPCRouteRule once we add support for GRPCRoutes.

#### Alternatives

- Pursue standardization in Gateway API through the GEP process before implementing our API. While it's possible the Gateway API would accept a GEP for this Filter, it will take a significant amount of time and is not guaranteed. After implementing it ourselves, we can always contribute this Filter back to the API.

### TLS Settings

_Extension type:_ TLS Options

_Resource type:_ N/A. We will need to document supported key-values.

_Role(s):_ Cluster Operator

_Attaches to:_ Gateway

_NGINX Context(s):_ server

Features:

- TLS Settings: SSL protocols, preferred ciphers, DH parameters file
- HSTS

NGINX Directives:

- [ssl_protocols](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols)
- [ssl_prefer_server_ciphers](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_prefer_server_ciphers)
- [ssl_dhparam](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam)
- [add_header](https://nginx.org/en/docs/http/ngx_http_headers_module.html)

These features are grouped because they are all TLS-related settings. TLS options are a good fit because they allow configuring these settings for Gateway Listeners and are restricted to the Cluster Operator role.

#### Future Work

- Contribute back to Gateway API by standardizing the keys.

#### Alternatives

- Policy: A TLS or Security Policy that attaches to Gateway listeners may also work. However, that would involve maintaining an additional CRD, whereas with TLS options, we only need to document and validate a few key-value pairs.
- ParametersRef: While ParametersRef is also scoped to the ClusterOperator, it applies to all Gateways attached to the GatewayClass. Gateways can have multiple listeners with different protocols, and the TLS settings may not apply to all listeners. In addition, ClusterOperators may want to enforce different TLS settings for different TLS listeners.

### Client Settings

_Extension type:_ Inherited Policy

_Resource type:_ CRD

_Role(s):_ Cluster Operator

_Attaches to:_ Gateway, Gateway Listener, HTTPRoute

_NGINX context(s)_: http, server, location

Features:

- Client max body size
- Client keepalive

NGINX directives:

- [`client_max_body_size`](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size): http, server, location
- [`client_body_timeout`](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout): http, server, location
- [`keepalive_requests`](https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_requests): http, server, location
- [`keepalive_time`](https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_time): http, server, location
- [`keepalive_timeout`](https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout): http, server, location
- [`keepalive_disable`](https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_disable): http, server, location

These features are grouped because they all deal with client traffic. An Inherited Policy fits this group best because there is a use case where the ClusterOperator sets a sane default for client max body size or client keepalives that has to be overridden by an Application Developer because of the unique attributes of the service they own.

#### Future Work

- Can add support for more `client_*` directives: `client_body_buffer_size`, `client_header_buffer_size`, etc.

#### Alternatives

- Direct Policy: A Direct Policy isn't a good fit for client settings because it does not support overrides and defaults.

### Upstream Settings

_Extension type:_ Direct Policy

_Resource type:_ CRD

_Role(s):_ Application Developer

_Attaches to:_ Backend

_NGINX Context(s):_ upstream, location (for active health checks)

OSS Features:

- Upstream zone size
- Upstream keepalives
- Load-balancing method (all except `least_time`)
- Limit connections to server
- Passive health checks

OSS NGINX directives/parameters:

- [`zone`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#zone)
- [`keepalive`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive)
- [`keepalive_requests`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive_requests)
- [`keepalive_time`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive_time)
- [`keepalive_timeout`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive_timeout)
- [`ip_hash`, `least_conn`, `random`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#ip_hash)
- [`max_conns`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server)
- [`fail_timeout`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server)
- [`max_fails`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server)

NGINX Plus Features:

- Request queueing
- Slow start
- Active health checks
- Load-balancing method (`least_time`)

NGINX Plus directives/parameters:

- [`slow_start`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server)
- [`queue`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#queue)
- [`health_check`](https://nginx.org/en/docs/http/ngx_http_upstream_hc_module.html#health_check)
- [`least_time`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#ip_hash)

These features are grouped because they all apply to the upstream context and make sense when attached to a Backend.

#### Alternatives

- Split out health checks into a separate Policy: The upstream settings group is large and could be split into two Policies based on use cases. However, we want to limit the number of CRDs we maintain and reduce the overlap between CRDs. Two Policies that set similar directives and parameters in the upstream context could be complex to implement.

### Authentication

_Extension type:_ Inherited Policy

_Resource type:_ CRD

_Role(s):_ Cluster Operator, Application Developer

_Attaches to:_ Gateway, Gateway Listener, HTTPRoute Rule (once it's supported)

_NGINX context(s):_ http, server, location

OSS Features:

- Basic Authentication (User/Password)
- API Key Authentication

OSS NGINX directives:

- [`auth_basic`](https://nginx.org/en/docs/http/ngx_http_auth_basic_module.html#auth_basic)
- [`auth_basic_user_file`](https://nginx.org/en/docs/http/ngx_http_auth_basic_module.html#auth_basic_user_file)

NGINX Plus Features:

- JSON Web Token (JWT) Authentication
- OIDC Authentication

NGINX Plus directives:

- [`auth_jwt`](https://nginx.org/en/docs/http/ngx_http_auth_jwt_module.html#auth_jwt)
- [`auth_jwt_type`](https://nginx.org/en/docs/http/ngx_http_auth_jwt_module.html#auth_jwt_type)
- [`auth_jwt_key_file`](https://nginx.org/en/docs/http/ngx_http_auth_jwt_module.html#auth_jwt_key_file)
- [`auth_jwt_key_request`](https://nginx.org/en/docs/http/ngx_http_auth_jwt_module.html#auth_jwt_key_request)

#### Future Work

- Add support for remote authentication or other authentication strategies.

#### Alternatives

- Begin with a Filter: Rather than beginning with an Inherited Policy, which is complex, we could start by adding an authentication Filter. This would allow us to roll out an authentication solution quickly to support the core use case for Application Developers. The use case for Cluster Operators applying authentication Policies is less well-known than the Application Developer's use case. Instead of leading with the Cluster Operator use case, we can begin with a Filter and wait for user feedback. Later, we can add a Policy and define how the Filter and Policy interact.
- Wait for Gateway API GEP on authentication: There is an open [GEP issue](https://github.com/kubernetes-sigs/gateway-api/issues/1494) for authentication in the Gateway API repository. Instead of rolling out our own authentication solution, we could wait or even champion this GEP. The advantages of waiting are that if the GEP is implemented, we won't have to deprecate our bespoke authentication Policy in favor of the GEP's implementation, and we won't have to maintain an additional Policy. However, we don't know how long it will take the GEP to move from open to implementable. Authentication is a critical feature, and it may not be prudent to wait.

### Observability

_Extension type:_ Direct Policy

_Resource type:_ CRD

_Role(s):_ Application Developer

_Attaches to:_ HTTPRoute, HTTPRoute rule

_NGINX context(s):_ http, server, location

Features:

- OTEL Tracing: enable/disable by Route or routing rule. Set span name, attributes, and context.

NGINX directives:

- [`otel_trace`](https://nginx.org/en/docs/ngx_otel_module.html#otel_trace)
- [`otel_trace_context`](https://nginx.org/en/docs/ngx_otel_module.html#otel_trace_context)
- [`otel_span_name`](https://nginx.org/en/docs/ngx_otel_module.html#otel_span_name)
- [`otel_span_attr`](https://nginx.org/en/docs/ngx_otel_module.html#otel_span_attr)

#### Future Work

Add support for:

- OTEL logging
- Metrics

#### Alternatives

- Combine with OTEL settings in Gateway Settings for one OTEL Policy: Rather than splitting tracing across two Policies, we could create a single tracing Policy. The issue with this approach is that some tracing settings -- such as exporter endpoint -- should be restricted to Cluster Operators, while settings like attributes should be available to Application Developers. If we combine these settings, RBAC will not be sufficient to restrict access across the settings. We will have to disallow certain fields based on the resource the Policy is attached to. This is a bad user experience.
- Inherited Policy: An Inherited Policy would be useful if there is a use case for the Cluster Operator enforcing or defaulting the OTEL tracing settings included in this policy.

### Proxy Settings

_Extension type:_ Inherited Policy

_Resource type:_ CRD

_Role(s):_ Cluster Operator, Application Developer

_Attaches to:_ Gateway, HTTPRoute, HTTPRoute Rule

_NGINX context(s):_ http, server, location

Features:

- Proxy buffering
- Connection timeouts
- Next upstream retries

NGINX directives:

- [`proxy_connect_timeout`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout)
- [`proxy_read_timeout`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout)
- [`proxy_send_timeout`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout)
- [`proxy_next_upstream`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream)
- [`proxy_next_upstream_timeout`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream_timeout)
- [`proxy_next_upstream_retries`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream_tries)
- [`proxy_buffering`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering)

#### Future Work

- Add other `proxy_*` directives

#### Alternatives

- Direct Policy: If there's no strong use case for the Cluster Operator setting sane defaults for these settings, then we can use a Direct Policy. The Direct Policy could attach to an HTTPRoute or HTTPRoute Rule, and the NGINX contexts would be server and location.

### Circuit Breaker/ Backup service

_Extension type:_ Direct Policy/Filter

_Resource type:_ CRD

_Role(s):_ Application Developer

_Attaches to:_ Backend/HTTPRoute Rule

_NGINX context(s):_ upstream

Features:

- Backup service/circuit breaker

NGINX upstream server parameters:

- [`backup`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server)
- [`fail_timeout`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server)
- [`max_fails`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server)

This [NGINX blog post](https://www.nginx.com/blog/microservices-reference-architecture-nginx-circuit-breaker-pattern/) describes the circuit breaker pattern and how to implement it with NGINX plus. The solution involves active health checks, rate-limiting, caching, and the `backup` directive. While this might be a complete solution, it is likely too heavyweight for NGINX Gateway Fabric.

The NGINX Kubernetes projects, NGINX Service Mesh and NGINX Ingress Controller offer two alternatives to circuit breaker.

NGINX Ingress Controller supports setting a [backup Service](https://github.com/nginxinc/kubernetes-ingress/tree/release-3.4/examples/custom-resources/backup-directive/virtual-server) that will be used when the primary servers are unavailable. The backup service must be of type ExternalName.

NGINX Service Mesh implements circuit breaking with the following [CRD](https://github.com/nginxinc/nginx-service-mesh/blob/main/pkg/apis/specs/v1alpha1/circuit_breaker.go):

```yaml
apiVersion: specs.smi.nginx.com/v1alpha1
kind: CircuitBreaker
metadata:
  name: circuit-breaker-example
  namespace: default
spec:
  destination:
    kind: Service
    name: target-svc
    namespace: default
  errors: 3
  timeoutSeconds: 30
  fallback:
    service: default/target-backup
    port: 80
```

The `spec.destination` field is similar to `targetRef` in that it attaches the CircuitBreaker setting to an object. This example configures a circuit breaker that will trip when there are three errors within 30 seconds. If the circuit breaker trips, requests to the destination Service will be routed to the fallback Service. After 30 seconds, the circuit breaker will reset, and the requests will be routed to the destination Service again. This implementation uses the NGINX `max_fails`, `fail_timeout`, and `backup` upstream server parameters.

NGINX Gateway Fabric could choose one of these approaches or design something new. It's difficult to propose an extension type for circuit breaker without more discussion and design. A Direct Policy that attaches to Backends or an HTTPRoute Filter could both work. Which one we choose will depend on the design of the circuit breaker and whether circuit breakers should be applied per Backend or Route.

#### Alternatives

- Include in the Upstream Settings Policy: The `fail_timeout` and `max_fails` directives are also used in passive health checks, which is a part of the Upstream Settings Policy. Including circuit breaking -- or just the `backup` directive -- might make more sense in this Policy.

## Testing

Each extension will be tested with a combination of unit and system tests. The details of the tests are out of scope for this Enhancement Proposal.

## Security Considerations

Each extension will need validation to prevent malicious or invalid NGINX configuration. The details of this validation are outside the scope of this Enhancement Proposal.

## Alternatives Considered

- Data Plane CRD or ConfigMap: Rather than adding nine different extensions APIs, we could add a single data plane CRD or ConfigMap that contains all the NGINX features described in this proposal. This would allow us to quickly expose NGINX functionality to the user and simplify our API. However, this object would become large and unfocused, a laundry list of NGINX directives and parameters. Additionally, with a single resource, there's no way to restrict access to certain fields with RBAC, which doesn't fit well with the role-based nature of the Gateway API.
- Add support for NGINX configuration injection: Adding nine extension APIs will take time. If we first add support for the injection of any NGINX configuration (see NGINX Ingress Controller [snippets](https://docs.nginx.com/nginx-ingress-controller/configuration/ingress-resources/advanced-configuration-with-snippets/) and [templates](https://docs.nginx.com/nginx-ingress-controller/configuration/global-configuration/custom-templates/) for examples), then we would immediately support all the NGINX functionality included in this proposal. However, this feature comes with a couple of problems:

  1. Security. The configuration cannot be validated before it is applied. This means that invalid or even malicious configuration can be injected. As a result, not every Cluster Operator would want this functionality turned on. So, for some users, this functionality would not be available.
  2. Requires NGINX knowledge. Without a first-class API exposing and abstracting away NGINX configuration, the users must be knowledgeable about NGINX. This will be great for NGINX power users but much more complicated for those unfamiliar with NGINX.
  3. Lack of status. No helpful status is set, and users must parse NGINX error messages—potential for frustration and bad user experiences.
     Problems aside, this feature will still be useful for our users and is something we should implement. Still, it does not eliminate the need for the extension APIs proposed in this document and should not be a higher priority.

## References

- [Policy and Metaresources GEP](https://gateway-api.sigs.k8s.io/geps/gep-713/)
- [Gateway API Extension Points](https://gateway-api.sigs.k8s.io/concepts/api-overview/?h=extension#extension-points)
- [TLS Extensions](https://gateway-api.sigs.k8s.io/guides/tls/?h=extensions#extensions)
