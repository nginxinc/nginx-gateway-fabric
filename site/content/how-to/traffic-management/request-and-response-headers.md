---
title: "HTTP Request and Response Headers"
description: "Learn how to modify request and response headers for your HTTP Route using NGINX Gateway Fabric."
weight: 700
toc: true
docs: ""
---

[HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/) filters can modify the headers during the request-response lifecycle. [HTTP Header Modifiers](https://gateway-api.sigs.k8s.io/guides/http-header-modifier/?h=request#http-header-modifiers) can be used to add, modify or remove headers in incoming requests. We have two types of [filter](https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional) that can be used to instruct the Gateway for desired behaviour.

1. [RequestHeaderModifier](https://gateway-api.sigs.k8s.io/guides/http-header-modifier/?h=request#http-request-header-modifier) to alter headers in request before forwarding the request upstream.
1. [ResponseHeaderModifier](https://gateway-api.sigs.k8s.io/guides/http-header-modifier/?h=request#http-response-header-modifier) to alter headers in response before responding to the downstream.


In this guide we will modify the headers of HTTP request and HTTP responses from clients. For an introduction to exposing your application, we recommend that you follow the [basic guide]({{< relref "/how-to/traffic-management/routing-traffic-to-your-app.md" >}}) first.


## Prerequisites

- [Install]({{< relref "/installation/" >}}) NGINX Gateway Fabric.
- [Expose NGINX Gateway Fabric]({{< relref "installation/expose-nginx-gateway-fabric.md" >}}) and save the public IP
  address and port of NGINX Gateway Fabric into shell variables:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   GW_PORT=<port number>
   ```

{{< note >}}In a production environment, you should have a DNS record for the external IP address that is exposed, and it should refer to the hostname that the gateway will forward for.{{< /note >}}

## Echo Application

### Deploy the Headers Application

Begin by deploying the `headers` application:

```shell
kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.2.0/examples/http-header-filter/headers.yaml
```

### Deploy the Gateway API Resources for the Headers Applications

The [gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/) resource is typically deployed by the [cluster operator](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#roles-and-personas_1). To deploy the gateway:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: echo
spec:
  gatewayClassName: nginx
  listeners:
  - name: http
    port: 80
    protocol: HTTP
EOF
```

This gateway defines a single listener on port 80. Since no hostname is specified, this listener matches on all hostnames.

The [HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/) is typically deployed by the [application developer](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#roles-and-personas_1). To deploy the `headers` HTTPRoute:


```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: headers
spec:
  parentRefs:
  - name: gateway
    sectionName: http
  hostnames:
  - "echo.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /headers
    filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        set:
        - name: My-Overwrite-Header
          value: this-is-the-only-value
        add:
        - name: Accept-Encoding
          value: compress
        - name: My-cool-header
          value: this-is-an-appended-value
        remove:
        - User-Agent
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        set:
        - name: My-Overwrite-Header-Response
          value: this-is-the-only-value-response
        add:
        - name: Accept-Encoding-Response
          value: compress-response
        - name: My-cool-header-Response
          value: this-is-an-appended-value-response
        remove:
        - Accept
    backendRefs:
    - name: headers
      port: 80
EOF
```

This HTTPRoute has a few important properties:

- The `parentRefs` references the gateway resource that we created, and specifically defines the `http` listener to attach to, via the `sectionName` field.
- `echo.example.com` is the hostname that is matched for all requests to the backends defined in this HTTPRoute.
- The `match` rule defines that all requests with the path prefix `/headers` are sent to the `headers` Service.
- There are two filters defined for the path prefix `/headers`

    1. `RequestHeaderModifier` filter sets the value of header `My-Overwrite-Header` to `this-is-the-only-value`, adds headers `Accept-Encoding` and `My-cool-header` and removes `User-Agent`.
    1. `ResponseHeaderModifier` filter sets the value of header `My-Overwrite-Header-Response` to `this-is-the-only-value-Response`, adds headers `Accept-Encoding-Response` and `My-cool-header-Response` and removes `Accept`.


{{< note >}}If the request does not have the header configured by the filter, then that header will be added to the request. If the request already has the header configured by the filter, then the value of the header in the filter will be appended to the value of the header in the request.{{< /note >}}

### Send Traffic to Headers

To access the application, we will use `curl` to send requests to the `headers` Service, including sending headers with our request.

{{< note >}}If you have a DNS record allocated for `echo.example.com`, you can send the request directly to that hostname, without needing to resolve.{{< /note >}}


1. Send traffic to headers to modify request headers.

```shell
curl -s --resolve echo.example.com:$GW_PORT:$GW_IP http://echo.example.com:$GW_PORT/headers -H "My-Cool-Header:my-client-value" -H "My-Overwrite-Header:dont-see-this"
```

```text
Headers:
  header 'Accept-Encoding' is 'compress'
  header 'My-cool-header' is 'my-client-value, this-is-an-appended-value'
  header 'My-Overwrite-Header' is 'this-is-the-only-value'
  header 'Host' is 'echo.example.com:$GW_PORT'
  header 'X-Forwarded-For' is '$GW_IP'
  header 'Connection' is 'close'
  header 'Accept' is '*/*'
```

```shell
curl -v --resolve echo.example.com:$GW_PORT:$GW_IP http://echo.example.com:$GW_PORT/headers
```


1. Send traffic to headers to modify response headers.


// TODO
