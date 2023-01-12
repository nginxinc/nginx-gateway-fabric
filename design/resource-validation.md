# Resource Validation

NGINX Kubernetes Gateway (NKG) must validate Gateway API resources for reliability, security, and conformity with the
Gateway API specification.

## Background

### Why Validate?

NKG transforms the Gateway API resources into NGINX configuration. Before transforming a resource, NKG needs to ensure
its validity, which is important for the following reasons:

1. *To prevent an invalid value from propagating into NGINX configuration*. For example, the URI in a path-based routing
   rule. The propagating has the following consequences:
    1. Invalid input will make NGINX fail to reload. Moreover, until the corresponding invalid config is removed from
       NGINX configuration, NKG will not be able to reload NGINX for any future configuration changes. This affects the
       reliability of NKG.
    2. Malicious input can breach the security of NKG. For example, if a malicious user can insert raw NGINX config (
       something similar to an SQL injection), they can configure NGINX to serve the files on the container filesystem.
       This affects the security of NKG.
2. *To conform to the Gateway API spec*. For example, if an HTTPRoute configures an unsupported filter, an
   implementation like NKG needs to "set Accepted Condition for the Route to `status: False`, with a Reason
   of `UnsupportedValue`".

### Validation by the Gateway API Project

To help the implementations with the validation, the Gateway API already includes:

* *The OpenAPI scheme with validation rules in the Gateway API CRDs*. It enforces the structure (for example, the field
  X must be a string) and the contents of the fields (for example, field Y only allows values 'a' and 'b').
  Additionally, it enforces the limits like max lengths on field values. Note:
  Kubernetes API server enforces this validation. To bypass it, a user needs to change the CRDs.
* *The webhook validation*. This validation is written in go and run as part of the webhook, which is included in the
  Gateway API installation files. The validation covers additional logic, not possible to implement in the CRDs. It does
  not repeat the validation from the CRDs. Note: a user can bypass this validation if the webhook is not installed.

However, the built-in validation rules do not cover all validation needs of NKG:

* The rules are not enough for NGINX. For example, the validation rule for the
  `value` of the path in a path-based routing rule allows symbols like `;`, `{`
  and `}`, which can break NGINX configuration for the
  corresponding [location](https://nginx.org/en/docs/http/ngx_http_core_module.html#location) block.
* The rules don't cover unsupported field cases. For example, the webhook does not know which filters are implemented by
  NKG, thus it cannot generate an appropriate error for NKG.

Additionally, as mentioned in [GEP-922](https://gateway-api.sigs.k8s.io/geps/gep-922/#implementers),
"implementers must not rely on webhook or CRD validation as a security mechanism. If field values need to be escaped to
secure an implementation, both webhook and CRD validation can be bypassed and cannot be relied on."

## Requirements

Design a validation mechanism for Gateway API resources.

### Personas

* *Cluster admin* who installs Gateway API (the CRDs and Webhook), installs NKG, creates Gateway and GatewayClass
  resources.
* *Application developer* who creates HTTPRoutes and other routes.

### User Stories

1. As a cluster admin, I'd like to share NKG among multiple application developers, specifically in a way that invalid
   resources of one developer do not affect on the resources of the other developers.
2. As a cluster admin/application developer, I expect that NKG rejects any invalid resources I create and I am able to
   see the reasons (errors) for that.

### Goals

* Ensure that NKG continues to work and/or fails predictably in the face of invalid input.
* Ensure that both cluster admin and application developers can see the validation errors reported about the resource
  they create (own).
* For the best UX, minimize the feedback loop: users should be able to see most of the validation errors reported by a
  Kubernetes API server during a CRUD operation on a resource.
* Ensure that the validation mechanism conforms to the Gateway API spec.

### Non-Goals

* Validation of non-Gateway API resources: Secrets, EndpointSlices. (For example, a TLS Secret resource might include a
  non-valid TLS cert that will make NGINX fail to reload).

## Design

We will introduce two validation methods to be run by NKG control plane:

1. Re-run of the Gateway API webhook validation
2. NKG-specific field validation

### Re-run of Webhook Validation

Before processing a resource, NKG will validate it using the functions from
the [validation package](https://github.com/kubernetes-sigs/gateway-api/tree/b241afc88e68c952cc0a59a5c72a51358dc2bada/apis/v1beta1/validation)
from the Gateway API. This will ensure that the webhook validation cannot be bypassed (it can be bypassed if the webhook
is not installed, misconfigured, or running a different version), and it will allow us to avoid repeating the same
validation in our code.

If a resource is invalid:

* NKG will not process it -- it will treat it as if the resource didn't exist. This also means that if the resource was
  updated from a valid to an invalid state, NKG will also ignore any previous valid state. For example, it will remove
  the generation configuration for an HTTPRoute resource.
* NKG will report the validation error as a
  Warning [Event](https://kubernetes.io/docs/reference/kubernetes-api/cluster-resources/event-v1/)
  for that resource. The Event message will describe the error and explain that the resource was ignored. We chose to
  report an Event instead of updating the status, because to update the status, NKG first needs to look inside the
  resource to determine whether it belongs to it or not. However, since the webhook validation applies to all parts of
  the spec of resource, it means NKG has to look inside the invalid resource and parse potentially invalid parts. To
  avoid that, NKG will report an Event. The owner of the resource will be able to see the Event.
* NGK will also report the validation error in the NGK logs.

### NKG-specific validation

After re-running the webhook validation, NKG will run NKG-specific validation, written in go.

NKG-specific validation will:

1. Ensure field values are considered valid by NGINX (cannot make NGINX fail to reload).
2. Ensure valid field values do not include any malicious configuration.
3. Report an error if an unsupported field is present in a resource (as the Gateway API spec prescribes).

NKG-specific validation will not include:

- *All* validation done by CRDs. NKG will only repeat the validation that addresses (1) and (2) in the list above with
  extra rules required by NGINX but missing in the CRDs. For example, NKG will not ensure the limits of field values.
- The validation done by the webhook (because it is done in the previous step).

If a resource is invalid, NKG will report the error in its status.

### Summary of Validation

The table below summarizes the validation methods NKG will use. Any Gateway API resource will be validated by the
following methods in order of their appearance in the table.

| Name                         | Type    | Component             | Scope                   | Feedback loop for errors                                                         | Can be bypassed?                                                                     |
|------------------------------|---------|-----------------------|-------------------------|----------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| CRD validation               | OpenAPI | Kubernetes API server | Structure, field values | Kubernetes API server returns any errors a response for an API call.             | Yes, if the CRDs are modified.                                                       |
| Webhook validation           | Go code | Gateway API webhook   | Field values            | Kubernetes API server returns any errors a response for an API call.             | Yes, if the webhook is not installed, misconfigured, or running a different version. |
| Re-run of webhook validation | Go code | NKG control plane     | Field values            | Errors are reported as Event for the resource.                                   | No                                                                                   |
| NGK-specific validation      | Go code | NKG control plane     | Field values            | Errors are reported in the status of a resource after its creation/modification. | No                                                                                   |

Notes:

* The amount and the extent of the validation should allow multiple application developers to share a single NKG (User
  story 1).
* We expect that most of the validation problems will be caught by CRD and webhook validation and reported quickly to
  users as a response to a Kubernetes API call (User story 2).

### Evolution

NKG will support more resources:

- More Gateway API resources. For those, NGK will use the four validation methods from the table in the previous
  section.
- Introduce NKG resources. For those, NKG will use CRD validation (the rules of which are fully controlled by us). The
  CRD validation will include the validation to prevent invalid NGINX configuration values and malicious values. Because
  the CRD validation can be bypassed, NKG control plane will need to run the same validation rules. In addition to that,
  NKG control plane will run any extra validation not possible to define via CRDs.

We will not introduce any NKG webhook in the cluster (it adds operational complexity for the cluster admin and is a
source of potential downtime -- a webhook failure disables CRUD operations on the relevant resources) unless we find
good reasons for that.

### Upgrades

Since NKG will use the validation package from the Gateway API project, when a new release happens, we will need to
upgrade the dependency and release a new version of NKG, provided that the validation code changed. However, if it did
not change, we do not need to release a new version. Note: other things from a new Gateway API release might prompt us
to release a new version like supporting a new field. See also
[GEP-922](https://gateway-api.sigs.k8s.io/geps/gep-922/#).

### Reliability

NGK processes two kinds of transactions:

* *Data plane transactions*. NGINX handles requests from clients that want to connect to applications exposed through
  NKG.
* *Control plane transactions*. NKG handles configuration requests (ex. a new HTTPRoute is created) from NKG users.

Invalid user input makes NGINX config invalid, which means NGINX will fail to reload, which will prevent any new control
plane transactions until that invalid value is fixed or removed. The proposed design addresses this issue by preventing
NKG from generating invalid NGINX configuration.

However, in case of bugs in the NKG validation code, NKG might still generate an invalid NGINX config. When that
happens, NGINX will fail to reload, but it will continue to use the last known valid config, so that the data plane
transactions will not be stopped. This situation must be reported to both the cluster admin and the app developers.
However, this is out of the scope of this design doc.

### Security

The proposed design ensures that the configuration values are properly validated before reaching NGINX config, which
will prevent a malicious user from misusing them. For example, it will not be possible to inject NGINX configuration
which can turn it into a web server serving the contents of the NKG data plane container file system.

## Alternatives Considered

### Utilize CRD Validation

It is [possible](https://github.com/hasheddan/k8s-cr-validator) to run CRD validation from Go code. However, this will
require NKG to be shipped with the Gateway API CRDs, which will increase the coupling between NKG and the Gateway API
version.

Additionally, the extra benefits are not clear: the validation proposed in this design document should adequately
address reliability and security issues. Also, disabling CRD validation in the API server is not easy for an application
developer -- they need to be a cluster admin to update the CRDs in the cluster.

At the same time, if a [convenient validation package](https://github.com/kubernetes-sigs/gateway-api/issues/926)
that includes CRD validation is developed, we will revisit the design.

### Write NKG-specific Validation Rules in Validation Language

It is possible to define validation rules in an expression language like [CEL](https://github.com/google/cel-spec). NKG
can load those rules, compile and run them.

Because (1) we need to define validation rules only to parts of Gateway API resources and (2) it is not necessary to
load them on the fly, the approach will not provide any benefits over defining those rules in go.

At the same time, we might use CEL for validating future NGK CRDs (CEL
is [supported](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#validation-rules)
in CRDs).
