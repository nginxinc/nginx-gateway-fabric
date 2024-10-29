---
title: "Use the SnippetsFilter API"
weight: 800
toc: true
docs: "DOCS-000"
---

This topic introduces Snippets, how to implement them using the `SnippetsFilter` API, and provides an example of how to use `SnippetsFilter` for rate limiting.

---

## Overview

Snippets allow users to insert NGINX configuration into different contexts of the
NGINX configurations that NGINX Gateway Fabric generates.

Snippets are for advanced NGINX users who need more control over the generated NGINX configuration,
and can be used in cases where Gateway API resources or NGINX extension policies don't apply.

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

---

## Setup

- To enable Snippets, [install]({{< relref "/installation/" >}}) NGINX Gateway Fabric with these modifications:
  - Using Helm: set the `nginxGateway.snippetsFilters.enable=true` Helm value.
  - Using Kubernetes manifests: set the `--snippets-filters` flag in the nginx-gateway container argument, add `snippetsfilters` to the RBAC
    rules with verbs `list` and `watch`, and add `snippetsfilters/status` to the RBAC rules with verb `update`. See this [example manifest](https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/main/deploy/snippets-filters/deploy.yaml) for clarification.

- Save the public IP address and port of NGINX Gateway Fabric into shell variables:

   ```text
  GW_IP=<ip address>
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

  This request should receive a response from the coffee Pod:

  ```text
  Server address: 10.244.0.7:8080
  Server name: coffee-76c7c85bbd-cf8nz
  ```

  Send a request to tea:

  ```shell
  curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
   ```

  This request should receive a response from the tea Pod:

  ```text
  Server address: 10.244.0.6:8080
  Server name: tea-76c7c85bbd-cf8nz
  ```

  Before we enable rate limiting, try sending multiple requests to coffee:

  ```shell
  for i in `seq 1 10`; do curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee; done
  ```

  You should see all successful responses in quick succession as we configured any rate limiting rules yet.

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
      value: limit_req_zone \$binary_remote_addr zone=coffeezone:10m rate=1r/s;
    - context: http.server.location
      value: limit_req zone=coffeezone burst=3 nodelay;
EOF
```

The Snippet uses the NGINX `limit_req_module` to configure rate limiting for this HTTPRoute and the
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

To use the `SnippetsFilter`, update the coffee HTTPRoute to reference it:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: coffee
spec:
  parentRefs:
    - name: gateway
      sectionName: http
  hostnames:
    - "cafe.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /coffee
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.nginx.org
            kind: SnippetsFilter
            name: coffee-rate-limiting-sf
      backendRefs:
        - name: coffee
          port: 80
EOF
```

Verify that the coffee HTTPRoute has been configured correctly:

```shell
kubectl describe httproutes.gateway.networking.k8s.io coffee
```

You should see the following conditions:

```text
Conditions:
      Last Transition Time:  2024-10-28T00:33:08Z
      Message:               The route is accepted
      Observed Generation:   2
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-10-28T00:33:08Z
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
Server address: 10.244.0.7:8080
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
      value: limit_req_zone \$binary_remote_addr zone=teazone:10m rate=1r/s;
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

Update the tea HTTPRoute to reference the `SnippetsFilter`:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: tea
spec:
  parentRefs:
    - name: gateway
      sectionName: http
  hostnames:
    - "cafe.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /tea
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.nginx.org
            kind: SnippetsFilter
            name: tea-rate-limiting-sf
      backendRefs:
        - name: tea
          port: 80
EOF
```

Verify that the tea HTTPRoute has been configured correctly:

```shell
kubectl describe httproutes.gateway.networking.k8s.io tea
```

You should see the following conditions:

```text
Conditions:
      Last Transition Time:  2024-10-28T00:33:08Z
      Message:               The route is accepted
      Observed Generation:   2
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-10-28T00:33:08Z
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
Server address: 10.244.0.6:8080
Server name: tea-76c7c85bbd-cf8nz
```

When processing a single request, the rate limiting configuration has no noticeable effect. Try sending
multiple requests.

```shell
for i in `seq 1 10`; do curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea; done
```

You should see all successful responses from the tea Pod, but they should be spaced apart roughly one second each as
expected through the rate limiting configuration.

This indicates that you've successfully used Snippets with the `SnippetsFilter` resource to configure two distinct rate limiting rules to different backend applications.
For an alternative method of modifying the NGINX configuration NGINX Gateway Fabric generates through Gateway API resources, check out
our supported [first-class policies]({{< relref "overview/custom-policies.md" >}}) which don't carry many of the aforementioned disadvantages of Snippets.

---

## See also

- [API reference]({{< relref "reference/api.md" >}}): all configuration fields for the `SnippetsFilter` API.
