# Gateway API Resource Validation

This document describes how NGINX Gateway Fabric (NGF) validates Gateway API resources.

## Overview

There are several reasons why NGF validates Gateway API resources:

- *Robustness*: to gracefully handle invalid resources.
- *Security*: to prevent malicious input from propagating to the NGINX configuration.
- *Correctness*: to conform to the Gateway API specification for handling invalid resources.

Ultimately, the goal is to ensure that NGINX continues to handle traffic even if invalid Gateway API resources were
created.

A Gateway API resource (a new resource or an update for the existing one) is validated by the following steps:

### For Kubernetes 1.25+

1. OpenAPI schema validation by the Kubernetes API server.
2. CEL validation by the Kubernetes API server.
3. Webhook validation by NGF.
4. Validation by NGF.

### For Kubernetes 1.23 and 1.24

1. OpenAPI schema validation by the Kubernetes API server.
2. Webhook validation by the Gateway API webhook.
3. Webhook validation by NGF.
4. Validation by NGF.

To confirm that a resource is valid and accepted by NGF, check that the `Accepted` condition in the resource status
has the Status field set to `True`. For example, in a status of a valid HTTPRoute, if NGF accepts a parentRef,
the status of that parentRef will look like this:

```text
Status:
  Parents:
    Conditions:
      Last Transition Time:  2023-03-30T23:18:00Z
      Message:               The route is accepted
      Observed Generation:   2
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
    Controller Name:         gateway.nginx.org/nginx-gateway-controller
    Parent Ref:
      Group:         gateway.networking.k8s.io
      Kind:          Gateway
      Name:          gateway
      Namespace:     default
      Section Name:  http
```

> Make sure the reported observed generation is the same as the resource generation.

The remaining part of this document describes each step in detail with examples of how validation errors are reported.

### Step 1 - OpenAPI Scheme Validation by Kubernetes API Server

The Kubernetes API server validates Gateway API resources against the OpenAPI schema embedded in the Gateway API CRDs.
For example, if you create an HTTPRoute with an invalid hostname `cafe.!@#$%example.com`, the API server will reject it
with the following error:

```shell
kubectl apply -f coffee-route.yaml
```

```text
The HTTPRoute "coffee" is invalid: spec.hostnames[0]: Invalid value: "cafe.!@#$%example.com": spec.hostnames[0] in body should match '^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'
```

> While unlikely, bypassing this validation step is possible if the Gateway API CRDs are modified to remove the validation.
> If this happens, Step 4 will reject any invalid values (from NGINX perspective).

### Step 2 - For Kubernetes 1.25+ - CEL Validation by Kubernetes API Server

The Kubernetes API server validates Gateway API resources using CEL validation embedded in the Gateway API CRDs.
It validates Gateway API resources using advanced rules unavailable in the OpenAPI schema validation.
For example, if you create a Gateway resource with a TCP listener that configures a hostname, the CEL validation will
reject it with the following error:


```shell
kubectl apply -f some-gateway.yaml
```

```text
The Gateway "some-gateway" is invalid: spec.listeners: Invalid value: "array": hostname must not be specified for protocols ['TCP', 'UDP']
```

More information on CEL in Kubernetes can be found [here](https://kubernetes.io/docs/reference/using-api/cel/).

### Step 2 - For Kubernetes 1.23 and 1.24 - Webhook Validation by Gateway API Webhook

The Gateway API comes with a validating webhook which is enabled by default in the Gateway API installation manifests.
It validates Gateway API resources using advanced rules unavailable in the OpenAPI schema validation. For example, if
you create a Gateway resource with a TCP listener that configures a hostname, the webhook will reject it with the
following error:

```shell
kubectl apply -f some-gateway.yaml
```

```text
Error from server: error when creating "some-gateway.yaml": admission webhook "validate.gateway.networking.k8s.io" denied the request: spec.listeners[1].hostname: Forbidden: should be empty for protocol TCP
```

> Bypassing this validation step is possible if the webhook is not running in the cluster.
> If this happens, Step 3 will reject the invalid values.

### Step 3 - Webhook validation by NGF
To ensure that the resources are validated with the webhook validation rules, even if the webhook is not running,
NGF performs the same validation. However, NGF performs the validation *after* the Kubernetes API server accepts
the resource.

Below is an example of how NGF rejects an invalid resource (a Gateway resource with a TCP listener that configures a
hostname) with a Kubernetes event:

```shell
kubectl describe gateway some-gateway
```

```text
. . .
Events:
  Type     Reason    Age   From                            Message
  ----     ------    ----  ----                            -------
  Warning  Rejected  6s    nginx-gateway-fabric-nginx  the resource failed webhook validation, however the Gateway API webhook failed to reject it with the error; make sure the webhook is installed and running correctly; validation error: spec.listeners[1].hostname: Forbidden: should be empty for protocol TCP; NGF will delete any existing NGINX configuration that corresponds to the resource
```

> This validation step always runs and cannot be bypassed.
> NGF will ignore any resources that fail the webhook validation, like in the example above.
> If the resource previously existed, NGF will remove any existing NGINX configuration for that resource.

### Step 4 - Validation by NGF

This step catches the following cases of invalid values:

- Valid values from the Gateway API perspective but not supported by NGF yet. For example, a feature in an
  HTTPRoute routing rule. Note: for the list of supported features,
  see [Gateway API Compatibility](gateway-api-compatibility.md) doc.
- Valid values from the Gateway API perspective, but invalid for NGINX, because NGINX has stricter validation
  requirements for certain fields. Such values will cause NGINX to fail to reload or operate erroneously.
- Invalid values (both from the Gateway API and NGINX perspectives) that were not rejected because Step 1 was bypassed.
  Similarly to the previous case, such values will cause NGINX to fail to reload or operate erroneously.
- Malicious values that inject unrestricted NGINX config into the NGINX configuration (similar to an SQL injection
  attack).

Below is an example of how NGF rejects an invalid resource. The validation error is reported via the status:

```shell
kubectl describe httproutes.gateway.networking.k8s.io coffee
```

```text
. . .
Status:
  Parents:
    Conditions:
      Last Transition Time:  2023-03-30T22:37:53Z
      Message:               All rules are invalid: spec.rules[0].matches[0].method: Unsupported value: "CONNECT": supported values: "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"
      Observed Generation:   1
      Reason:                UnsupportedValue
      Status:                False
      Type:                  Accepted
    Controller Name:         gateway.nginx.org/nginx-gateway-controller
    Parent Ref:
      Group:         gateway.networking.k8s.io
      Kind:          Gateway
      Name:          prod-gateway
      Namespace:     default
      Section Name:  http
```

> This validation step always runs and cannot be bypassed.
