# Gateway API Resource Validation

This document describes how NGINX Kubernetes Gateway (NKG) validates Gateway API resources.

## Overview

There are several reasons why NKG validates Gateway API resources:

- *Robustness*: to gracefully handle invalid resources.
- *Security*: to prevent malicious input from propagating to the NGINX configuration.
- *Correctness*: to conform to the Gateway API specification for handling invalid resources.

Ultimately, the goal is to ensure that NGINX continues to handle traffic even if invalid Gateway API resources were
created.

A Gateway API resource (a new resource or an update for the existing one) is validated by the following steps:

1. OpenAPI schema validation by the Kubernetes API server.
2. Webhook validation by the Gateway API webhook.
3. Webhook validation by NKG.
4. Validation by NKG.

To confirm that a resource is valid and accepted by NKG, check that the `Accepted` condition in the resource status
has the Status field set to `True`. For example, in a status of a valid HTTPRoute, if NKG accepts a parentRef,
the status of that parentRef will look like this:
```
Status:
  Parents:
    Conditions:
      Last Transition Time:  2023-03-30T23:18:00Z
      Message:               The route is accepted
      Observed Generation:   2
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
    Controller Name:         k8s-gateway.nginx.org/nginx-gateway-controller
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

```
kubectl apply -f coffee-route.yaml 
The HTTPRoute "coffee" is invalid: spec.hostnames[0]: Invalid value: "cafe.!@#$%example.com": spec.hostnames[0] in body should match '^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'
```

> While unlikely, bypassing this validation step is possible if the Gateway API CRDs are modified to remove the validation.
> If this happens, Step 4 will reject any invalid values (from NGINX perspective).

### Step 2 - Webhook Validation by Gateway API Webhook

The Gateway API comes with a validating webhook which is enabled by default in the Gateway API installation manifests.
It validates Gateway API resources using advanced rules unavailable in the OpenAPI schema validation. For example, if
you create a Gateway resource with a TCP listener that configures a hostname, the webhook will reject it with the
following error:

```
kubectl apply -f gateway.yaml 
Error from server: error when creating "gateway.yaml": admission webhook "validate.gateway.networking.k8s.io" denied the request: spec.listeners[1].hostname: Forbidden: should be empty for protocol TCP
```

> Bypassing this validation step is possible if the webhook is not running in the cluster.
> If this happens, Step 3 will reject the invalid values.

### Step 3 - Webhook validation by NKG

The previous step relies on the Gateway API webhook running in the cluster. To ensure that the resources are validated
with the webhook validation rules, even if the webhook is not running, NKG performs the same validation. However, NKG
performs the validation *after* the Kubernetes API server accepts the resource.

Below is an example of how NKG rejects an invalid resource (a Gateway resource with a TCP listener that configures a
hostname) with a Kubernetes event:

```
kubectl describe gateway gateway
. . .
Events:
  Type     Reason    Age   From                            Message
  ----     ------    ----  ----                            -------
  Warning  Rejected  6s    nginx-kubernetes-gateway-nginx  the resource failed webhook validation, however the Gateway API webhook failed to reject it with the error; make sure the webhook is installed and running correctly; validation error: spec.listeners[1].hostname: Forbidden: should be empty for protocol TCP; NKG will delete any existing NGINX configuration that corresponds to the resource
```

> This validation step always runs and cannot be bypassed.

> NKG will ignore any resources that fail the webhook validation, like in the example above.
> If the resource previously existed, NKG will remove any existing NGINX configuration for that resource.

### Step 4 - Validation by NKG

This step catches the following cases of invalid values:

* Values valid from the Gateway API perspective but not supported by NKG yet. For example, a certain filter in an
  HTTPRoute routing rule.
* Some values in Gateway API resources which are valid by the CRD and webhook validation, but invalid for NGINX. Such
  values will cause NGINX to fail to reload or operate erroneously.
* Invalid values in Gateway API resources that were not rejected because Step 1 was bypassed.
* Malicious values that inject unrestricted NGINX config into the NGINX configuration (similar to an SQL injection
  attack).

Below is an example of how NGK rejects an invalid resource. The validation error is reported via the status:

```
kubectl describe httproutes.gateway.networking.k8s.io coffee
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
    Controller Name:         k8s-gateway.nginx.org/nginx-gateway-controller
    Parent Ref:
      Group:         gateway.networking.k8s.io
      Kind:          Gateway
      Name:          gateway
      Namespace:     default
      Section Name:  http
```

> This validation step always runs and cannot be bypassed.
