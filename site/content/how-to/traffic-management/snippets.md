---
title: "Use the SnippetsFilter API"
weight: 800
toc: true
docs: "DOCS-000"
---

This topic introduces Snippets and how to implement them using the `SnippetsFilter` API. It provides an example of how to use a Snippet for rate limiting, and how to investigate errors caused by misconfiguration.

---

## Overview

Snippets allow users to insert NGINX configuration into different contexts of the
NGINX configurations that NGINX Gateway Fabric generates.

Snippets are for advanced NGINX users who need more control over the generated NGINX configuration,
and can be used in cases where Gateway API resources or NGINX extensions policies don't apply.

Users can configure Snippets through the `SnippetsFilter` API. `SnippetsFilter` is an [HTTPRouteFilter](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter)
that attaches to an HTTP/GRPCRoute rule and is intended to modify NGINX configuration specifically for that Route rule.

---

## Disadvantages of Snippets

{{< warning >}} We recommend managing NGINX configuration through Gateway API resources, [first-class policies]({{< relref "overview/custom-policies.md" >}}), and other existing [NGINX extensions]({{< relref "data-plane-configuration.md" >}})
before using Snippets. {{< /warning >}}

Snippets are configured using the `SnippetsFilter` API, but are disabled by default due to their complexity and security implications.

Snippets have the following disadvantages:

- _Complexity_. Snippets require you to:
  - Understand NGINX configuration primitives to implement correct NGINX configuration.
  - Understand how NGINX Gateway Fabric generates NGINX configuration so that a Snippet doesn’t interfere with the other features in the configuration.
- _Decreased robustness_. An incorrect Snippet can invalidate NGINX configuration, causing reload failures. Until the snippet is fixed, it will prevent any new configuration updates, including updates for the other Gateway resources.
- _Security implications_. Snippets give access to NGINX configuration primitives, which are not validated by NGINX Gateway Fabric. For example, a Snippet can configure NGINX to serve the TLS certificates and keys used for TLS termination for Gateway resources.

{{< note >}} If the NGINX configuration includes an invalid Snippet, NGINX will continue to operate with the last valid configuration. {{< /note >}}

{{< note >}} If you end up using Snippets and run into situations where an NGINX directive fails to be applied, please create an issue in the
[NGINX Gateway Fabric Github repository](https://github.com/nginxinc/nginx-gateway-fabric). {{< /note >}}

## Setup

- To enable Snippets, [install]({{< relref "/installation/" >}}) NGINX Gateway Fabric with these modifications:
  - Using Helm: set the `nginxGateway.snippetsFilters.enable` command line argument to true.
  - Using Kubernetes manifests: set the `--snippets-filters` flag in the nginx-gateway container argument, add `snippetsfilters` to the RBAC
    rules with verbs `list` and `watch`, and add `snippetsfilters/status` to the RBAC rules with verb `update`. See this [example manifest](https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/main/deploy/snippets-filters/deploy.yaml) for clarification.

- Save the public IP address and port of NGINX Gateway Fabric into shell variables:

   ```text
  GW_IP=XXX.YYY.ZZZ.III
  GW_PORT=<port number>
  ```

  {{< note >}} In a production environment, you should have a DNS record for the external IP address that is exposed, and it should refer to the hostname that the gateway will forward for. {{< /note >}}

- Create the coffee and tea example applications:

  ```yaml
  kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.4.0/examples/snippets-filter/app.yaml
  ```

- Create a Gateway:

  ```yaml
  kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.4.0/examples/snippets-filter/gateway.yaml
   ```

- Create HTTPRoutes for the coffee and tea applications:

  ```yaml
  kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.4.0/examples/snippets-filter/httproutes.yaml
   ```

- Test the configuration:

  You can send traffic to the coffee and tea applications using the external IP address and port for NGINX Gateway Fabric.

  Send a request to coffee:

  ```shell
  curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
  ```

  Send a request to tea:

  ```shell
  curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
  ```

  Both requests should receive this response from the backend Pod:

  ```text
  <html>
  <head><title>500 Internal Server Error</title></head>
  <body>
  <center><h1>500 Internal Server Error</h1></center>
  <hr><center>nginx</center>
  </body>
  </html>
  ```

  Use `kubectl describe` to investigate HTTPRoutes for the error:

  ```shell
  kubectl describe httproutes.gateway.networking.k8s.io
  ```

  You should see the following conditions:

  ```text
   Conditions:
      Last Transition Time:  2024-10-22T23:43:11Z
      Message:               The route is accepted
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-10-22T23:43:11Z
      Message:               spec.rules[0].filters[0].extensionRef: Not found: v1.LocalObjectReference{Group:"gateway.nginx.org", Kind:"SnippetsFilter", Name:"coffee-rate-limiting-sf"}
      Observed Generation:   1
      Reason:                InvalidFilter
      Status:                False
      Type:                  ResolvedRefs
  .
  .
  .
   Conditions:
      Last Transition Time:  2024-10-22T23:43:14Z
      Message:               The route is accepted
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-10-22T23:43:14Z
      Message:               spec.rules[0].filters[0].extensionRef: Not found: v1.LocalObjectReference{Group:"gateway.nginx.org", Kind:"SnippetsFilter", Name:"tea-rate-limiting-sf"}
      Observed Generation:   1
      Reason:                InvalidFilter
      Status:                False
      Type:                  ResolvedRefs
  ```

  The HTTPRoutes created earlier both reference `SnippetsFilter` resources that do not currently
  exist, creating a 500 error code response returned on requests that are processed by these HTTPRoutes.
  This issue will be resolved using SnippetsFilters.

---

## Configure rate limiting to the coffee application

Configure rate limiting to the coffee application by adding the following `SnippetsFilter`:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.nginx.org/v1alpha1
kind: SnippetsFilter
metadata:
  name: coffee-rate-limiting-sf
spec:
  snippets:
    - context: http
      value: limit_req_zone $binary_remote_addr zone=coffeezone:10m rate=1r/s;
    - context: http.server.location
      value: limit_req zone=coffeezone burst=3 nodelay;
EOF
```

This `SnippetsFilter` is already referenced by the HTTPRoute created during setup, so it will immediately apply
to the HTTPRoute. The Snippet uses the NGINX `limit_req_module` to configure rate limiting for this HTTPRoute and the
backend coffee application. This snippet will limit the request processing rate to 1 request per second, and if there
are more than 3 requests in queue, it will throw a 503 error.

Verify that the `SnippetsFilter` is Accepted:

```shell
kubectl describe snippetsfilters.gateway.nginx.org coffee-rate-limiting-sf
```

You should see the following status:

```text
Status:
  Controllers:
    Conditions:
      Last Transition Time:  2024-10-21T22:20:22Z
      Message:               SnippetsFilter is accepted
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
    Controller Name:         gateway.nginx.org/nginx-gateway-controller
Events:                      <none>
```

Verify that the coffee `HTTPRoute` which had an `InvalidFilter` condition earlier, no longer has that condition.

```shell
kubectl describe httproutes.gateway.networking.k8s.io coffee
```

You should see the following conditions:

```text
Conditions:
      Last Transition Time:  2024-10-23T00:33:08Z
      Message:               The route is accepted
      Observed Generation:   2
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-10-23T00:33:08Z
      Message:               All references are resolved
      Observed Generation:   2
      Reason:                ResolvedRefs
      Status:                True
      Type:                  ResolvedRefs
```

Test that the `SnippetsFilter` is configured and has successfully applied the rate limiting NGINX configuration changes.

Send a request to coffee:

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
```

This request should receive a response from the coffee Pod:

```text
Server address: 10.244.0.9:8080
Server name: coffee-76c7c85bbd-cf8nz
```

When processing a single request, the rate limiting configuration has no noticeable effect. Try to exceed the
set rate limit with a script that sends multiple requests.

```shell
for i in `seq 1 10`; do curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee; done
```

You should see some successful responses from the coffee Pod, however there should be multiple `503` responses such as:

```text
Request ID: 890c17df930ef1ef573feed3c6e81290
<html>
<head><title>503 Service Temporarily Unavailable</title></head>
<body>
<center><h1>503 Service Temporarily Unavailable</h1></center>
<hr><center>nginx</center>
</body>
</html>
```

This is the default error response given by NGINX when the rate limit burst is exceeded, meaning our `SnippetsFilter`
correctly applied our rate limiting NGINX configuration changes.

## Configure rate limiting to the tea application

Configure a different set of rate limiting rules to the tea application by adding the following `SnippetsFilter`:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.nginx.org/v1alpha1
kind: SnippetsFilter
metadata:
  name: tea-rate-limiting-sf
spec:
  snippets:
    - context: http
      value: limit_req_zone $binary_remote_addr zone=teazone:10m rate=1r/s;
    - context: http.server.location
      value: limit_req zone=teazone burst=3;
EOF
```

This `SnippetsFilter` is the same as the one applied to the coffee HTTPRoute, however it removes the `nodelay` setting
on the `limit_req` directive. This forces a delay on the incoming requests to match the rate set in `limit_req_zone`.

Verify that the `SnippetsFilter` is Accepted:

```shell
kubectl describe snippetsfilters.gateway.nginx.org tea-rate-limiting-sf
```

You should see the following status:

```text
Status:
  Controllers:
    Conditions:
      Last Transition Time:  2024-10-21T22:20:24Z
      Message:               SnippetsFilter is accepted
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
    Controller Name:         gateway.nginx.org/nginx-gateway-controller
Events:                      <none>
```

Verify that the tea `HTTPRoute` which had an `InvalidFilter` condition earlier, no longer has that condition.

```shell
kubectl describe httproutes.gateway.networking.k8s.io tea
```

You should see the following conditions:

```text
Conditions:
      Last Transition Time:  2024-10-23T00:33:08Z
      Message:               The route is accepted
      Observed Generation:   2
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-10-23T00:33:08Z
      Message:               All references are resolved
      Observed Generation:   2
      Reason:                ResolvedRefs
      Status:                True
      Type:                  ResolvedRefs
```

Test that the `SnippetsFilter` is configured and has successfully applied the rate limiting NGINX configuration changes.

Send a request to tea:

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
```

This request should receive a response from the tea Pod:

```text
Server address: 10.244.0.7:8080
Server name: tea-76c7c85bbd-cf8nz
```

When processing a single request, the rate limiting configuration has no noticeable effect. Try sending
multiple requests.

```shell
for i in `seq 1 10`; do curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea; done
```

You should see all successful responses from the tea Pod, but they should be spaced apart roughly one second each as
expected through the rate limiting configuration.

You've now used the `SnippetFilter` resource to configure two distinct rate limiting rules to different backend applications through the use of Snippets.
For an alternative method of modifying the NGINX configuration NGINX Gateway Fabric generates through Gateway API resources, check out
our supported [first-class policies]({{< relref "overview/custom-policies.md" >}}) which don't carry many of the aforementioned disadvantages of Snippets.

---

## Further reading

- [API reference]({{< relref "reference/api.md" >}}): all configuration fields for the `SnippetsFilter` API.
