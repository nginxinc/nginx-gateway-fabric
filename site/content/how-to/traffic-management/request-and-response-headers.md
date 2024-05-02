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

## Request Header Filter

Follow this [guide](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples/http-request-header-filter/README.md) to modify request headers for a client request.


## Response Header Filter

Follow this [guide](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples/http-response-header-filter/README.md) to modify response headers for a client request.
