---
title: "HTTP redirects and rewrites"
weight: 400
toc: true
docs: "DOCS-1424"
---

Learn how to redirect or rewrite your HTTP traffic using NGINX Gateway Fabric.

## Overview

[HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/) filters can be used to configure HTTP redirects or rewrites. Redirects return HTTP 3XX responses to a client, instructing it to retrieve a different resource. Rewrites modify components of a client request (such as hostname and/or path) before proxying it upstream.

{{< note >}}NGINX Gateway Fabric currently does not support path-based redirects.{{< /note >}}

To see an example of a redirect using scheme and port, see the [HTTPS Termination]({{< relref "/how-to/traffic-management/https-termination.md" >}}) guide.

In this guide, we will be configuring a path URL rewrite.

## Before you begin

- [Install]({{< relref "installation/" >}}) NGINX Gateway Fabric.
- Save the public IP address and port of NGINX Gateway Fabric into shell variables:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   GW_PORT=<port number>
   ```

{{< note >}}In a production environment, you should have a DNS record for the external IP address that is exposed, and it should refer to the hostname that the gateway will forward for.{{< /note >}}

## Set up

Create the **coffee** application in Kubernetes by copying and pasting the following block into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coffee
spec:
  replicas: 1
  selector:
    matchLabels:
      app: coffee
  template:
    metadata:
      labels:
        app: coffee
    spec:
      containers:
      - name: coffee
        image: nginxdemos/nginx-hello:plain-text
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: coffee
spec:
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: coffee
EOF
```

This will create the **coffee** service and a deployment. Run the following command to verify the resources were created:

```shell
kubectl get pods,svc
```

Your output should include the **coffee** pod and the **coffee** service:

```text
NAME                          READY   STATUS      RESTARTS   AGE
pod/coffee-6b8b6d6486-7fc78   1/1     Running   0          40s


NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/coffee       ClusterIP   10.96.189.37   <none>        80/TCP    40s
```

## Configure a path rewrite

To create the **cafe** gateway, copy and paste the following into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: cafe
spec:
  gatewayClassName: nginx
  listeners:
  - name: http
    port: 80
    protocol: HTTP
EOF
```

The following HTTPRoute defines two filters that will rewrite requests such as the following:

- `http://cafe.example.com/coffee` to `http://cafe.example.com/beans`
- `http://cafe.example.com/coffee/flavors` to `http://cafe.example.com/beans`
- `http://cafe.example.com/latte/prices` to `http://cafe.example.com/prices`

To create the httproute resource, copy and paste the following into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: coffee
spec:
  parentRefs:
  - name: cafe
    sectionName: http
  hostnames:
  - "cafe.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /coffee
    filters:
    - type: URLRewrite
      urlRewrite:
        path:
          type: ReplaceFullPath
          replaceFullPath: /beans
    backendRefs:
    - name: coffee
      port: 80
  - matches:
    - path:
        type: PathPrefix
        value: /latte
    filters:
    - type: URLRewrite
      urlRewrite:
        path:
          type: ReplacePrefixMatch
          replacePrefixMatch: /
    backendRefs:
    - name: coffee
      port: 80
EOF
```

## Send traffic

Using the external IP address and port for NGINX Gateway Fabric, we can send traffic to our coffee application.

{{< note >}}If you have a DNS record allocated for `cafe.example.com`, you can send the request directly to that hostname, without needing to resolve.{{< /note >}}

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee/flavors
```

Notice in the output that the URI has been rewritten:

```text
Server address: 10.244.0.6:8080
Server name: coffee-6b8b6d6486-7fc78
...
URI: /beans
```

Other examples:

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
```

```text
Server address: 10.244.0.6:8080
Server name: coffee-6b8b6d6486-7fc78
...
URI: /beans
```

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/latte/prices
```

```text
Server address: 10.244.0.6:8080
Server name: coffee-6b8b6d6486-7fc78
...
URI: /prices
```

## Further reading

To learn more about redirects and rewrites using the Gateway API, see the following resource:

- [Gateway API Redirects and Rewrites](https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/)
