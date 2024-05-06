---
title: "HTTP Response Headers"
description: "Learn how to modify the response headers of your application using NGINX Gateway Fabric."
weight: 700
toc: true
docs: ""
---

[HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/) filters can modify the headers during the request-response lifecycle. [HTTP Header Modifiers](https://gateway-api.sigs.k8s.io/guides/http-header-modifier/?h=request#http-header-modifiers) can be used to add, modify or remove headers in incoming requests. We have two types of [filter](https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional) that can be used to instruct the Gateway for desired behaviour.

1. [ResponseHeaderModifier](https://gateway-api.sigs.k8s.io/guides/http-header-modifier/?h=request#http-response-header-modifier) to alter headers in response before responding to the downstream.


In this guide we will modify the headers for HTTP responses when client requests are made. For an introduction to exposing your application, we recommend that you follow the [basic guide]({{< relref "/how-to/traffic-management/routing-traffic-to-your-app.md" >}}) first.


## Prerequisites

- [Install]({{< relref "/installation/" >}}) NGINX Gateway Fabric.
- [Expose NGINX Gateway Fabric]({{< relref "installation/expose-nginx-gateway-fabric.md" >}}) and save the public IP
  address and port of NGINX Gateway Fabric into shell variables:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   GW_PORT=<port number>
   ```

{{< note >}}In a production environment, you should have a DNS record for the external IP address that is exposed, and it should refer to the hostname that the gateway will forward for.{{< /note >}}


## Response Header Filter

In this guide we will deploy NGINX Gateway Fabric and use HTTPRoute resources to route traffic to the server. We will use `ResponseHeaderModifier` filter to modify the HTTP response headers.


### Deploy the Headers application

Begin by deploying the `headers` applications:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.3.0/examples/http-response-header-filter/headers.yaml
   ```

Verify if the Pod is running in the `default` Namespace:

   ```shell
   kubectl -n default get pods
   ```

   ```text
   NAME                      READY   STATUS    RESTARTS   AGE
   headers-6f854c478-k9z2f   1/1     Running   0          32m
   ```

### Deploy the Gateway API Resources for the Coffee Applications


The [gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/) resource is typically deployed by the [cluster operator](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#roles-and-personas_1). To deploy the gateway:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gateway
spec:
  gatewayClassName: nginx
  listeners:
  - name: http
    port: 80
    protocol: HTTP
EOF
```

### Configure Routing

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
  - "cafe.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /headers
    filters:
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        set:
        - name: X-Header-Set
          value: overwritten-value
        add:
        - name: X-Header-Add
          value: this-is-the-appended-value
        remove:
        - X-Header-Remove
    backendRefs:
    - name: headers
      port: 80
EOF
```

This HTTPRoute has a few important properties:

- The `parentRefs` references the gateway resource that we created, and specifically defines the `http` listener to attach to, via the `sectionName` field.
- `cafe.example.com` is the hostname that is matched for all requests to the backends defined in this HTTPRoute.
- The `match` rule defines that all requests with the path prefix `/headers` are sent to the `headers` Service.
- There is `ResponseHeaderModifier` filter defined for the path prefix `/headers` to overwrite value for header `X-Header-Set`, append value to header `X-Header-Add` and remove `X-Header-Remove` header from the HTTP response.


### Send Traffic to Headers

To access the application, we will use `curl` to send requests to the `headers` endpoint.

Notice our configured header values can be seen in the `responseHeaders` section below, and that the `X-Header-Remove` header is absent. The header `X-Header-Add` gets appended with the new value and `X-Header-Set` gets overwritten to `overwritten-value` as defined in the *HttpRoute*.

```shell
curl -v -i --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/headers
```

```text
HTTP/1.1 200 OK
Server: nginx/1.25.5
Date: Mon, 06 May 2024 17:58:33 GMT
Content-Type: text/plain
Content-Length: 2
Connection: keep-alive
X-Header-Unmodified: unmodified
X-Header-Add: add-to
X-Header-Add: this-is-the-appended-value
X-Header-Set: overwritten-value

ok
```


## Further Reading

To learn more about the Gateway API and the resources we created in this guide, check out the following Kubernetes documentation resources:

- [Gateway API Overview](https://gateway-api.sigs.k8s.io/concepts/api-overview/)
- [Deploying a simple Gateway](https://gateway-api.sigs.k8s.io/guides/simple-gateway/)
- [HTTP Routing](https://gateway-api.sigs.k8s.io/guides/http-routing/)
