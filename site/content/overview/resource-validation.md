---
title: "Gateway API Resource Validation"
weight: 800
toc: true
docs: "DOCS-000"
---

## Overview

NGINX Gateway Fabric validates Gateway API resources for several reasons:

- _Robustness_: to gracefully handle invalid resources.
- _Security_: to prevent malicious input from propagating to the NGINX configuration.
- _Correctness_: to conform to the Gateway API specification for handling invalid resources.

The process involves four different steps, explained in detail in this document, with the goal of making sure that NGINX continues to handle traffic even if invalid Gateway API resources were created.

## Step 1 - OpenAPI Scheme validation by Kubernetes API Server

The Kubernetes API server validates Gateway API resources against the OpenAPI schema embedded in the Gateway API CRDs. For example, if you create an HTTPRoute with an invalid hostname "cafe.!@#$%example.com", the API server will reject it with the following error:

```shell
kubectl apply -f coffee-route.yaml
```

```text
The HTTPRoute "coffee" is invalid: spec.hostnames[0]: Invalid value: "cafe.!@#$%example.com": spec.hostnames[0] in body should match '^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'
```

{{< note >}}While unlikely, bypassing this validation step is possible if the Gateway API CRDs are modified to remove the validation. If this happens, Step 4 will reject any invalid values (from NGINX perspective).{{< /note >}}

## Step 2 - CEL or Webhook validation by Kubernetes

- **Kubernetes 1.25 and later - CEL validation by Kubernetes API Server**

   The Kubernetes API server validates Gateway API resources using CEL validation embedded in the Gateway API CRDs. It validates Gateway API resources using advanced rules unavailable in the OpenAPI schema validation. For example, if you create a Gateway resource with a TCP listener that configures a hostname, the CEL validation will reject it with the following error:

   ```shell
   kubectl apply -f some-gateway.yaml
   ```

   ```text
   The Gateway "some-gateway" is invalid: spec.listeners: Invalid value: "array": hostname must not be specified for protocols ['TCP', 'UDP']
   ```

   More information on CEL in Kubernetes can be found [here](https://kubernetes.io/docs/reference/using-api/cel/).


- **Kubernetes 1.23 and 1.24 - Webhook validation by Gateway API Webhook**

   The validating webhook must be [installed for these Kubernetes versions]({{<relref "installation/installing-ngf/helm/#installing-the-gateway-api-resources">}}. It validates Gateway API resources using advanced rules unavailable in the OpenAPI schema validation. For example, if you create a Gateway resource with a TCP listener that configures a hostname, the webhook will reject it with the following error:

   ```shell
   kubectl apply -f some-gateway.yaml
   ```

   ```text
   Error from server: error when creating "some-gateway.yaml": admission webhook "validate.gateway.networking.k8s.io" denied the request: spec.listeners[1].hostname: Forbidden: should be empty for protocol TCP
   ```

{{< note >}}Bypassing this validation step is possible if the webhook is not running in the cluster. If this happens, Step 3 will reject the invalid values.{{< /note >}}

## Step 3 - Webhook validation by NGINX Gateway Fabric

To ensure that the resources are validated with the webhook validation rules, even if the webhook is not running, NGINX Gateway Fabric performs the same validation. However, NGINX Gateway Fabric performs the validation _after_ the Kubernetes API server accepts the resource.

Below is an example of how NGINX Gateway Fabric rejects an invalid resource (a Gateway resource with a TCP listener that configures a hostname) with a Kubernetes event:

```shell
kubectl describe gateway some-gateway
```

```text
. . .
Events:
  Type     Reason    Age   From                            Message
  ----     ------    ----  ----                            -------
  Warning  Rejected  6s    nginx-gateway-fabric-nginx  the resource failed webhook validation, however the Gateway API webhook failed to reject it with the error; make sure the webhook is installed and running correctly; validation error: spec.listeners[1].hostname: Forbidden: should be empty for protocol TCP; NGINX Management Suite will delete any existing NGINX configuration that corresponds to the resource
```

{{< note >}}This validation step always runs and cannot be bypassed. NGINX Gateway Fabric will ignore any resources that fail the webhook validation, like in the example above. If the resource previously existed, NGINX Gateway Fabric will remove any existing NGINX configuration for that resource.{{< /note >}}

## Step 4 - Validation by NGINX Gateway Fabric

This step catches the following cases of invalid values:

- Valid values from the Gateway API perspective but not supported by NGINX Gateway Fabric yet. For example, a feature in an HTTPRoute routing rule. For the list of supported features see [Gateway API Compatibility](gateway-api-compatibility.md) doc.
- Valid values from the Gateway API perspective, but invalid for NGINX, because NGINX has stricter validation requirements for certain fields. These values will cause NGINX to fail to reload or operate erroneously.
- Invalid values (both from the Gateway API and NGINX perspectives) that were not rejected because Step 1 was bypassed. Similar to the previous case, these values will cause NGINX to fail to reload or operate erroneously.
- Malicious values that inject unrestricted NGINX config into the NGINX configuration (similar to an SQL injection attack).

Below is an example of how NGINX Gateway Fabric rejects an invalid resource. The validation error is reported via the status:

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

{{< note >}}This validation step always runs and cannot be bypassed.{{< /note >}}

## Confirm validation

To confirm that a resource is valid and accepted by NGINX Gateway Fabric, check that the **Accepted** condition in the resource status has the Status field set to **True**. For example, in a status of a valid HTTPRoute, if NGINX Gateway Fabric accepts a parentRef, the status of that parentRef will look like this:

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

{{< note>}}Make sure the reported observed generation is the same as the resource generation.{{< /note >}}
