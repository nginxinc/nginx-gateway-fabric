---
title: "HTTP redirects and rewrites"
weight: 400
toc: true
docs: "DOCS-1424"
---

Learn how to redirect or rewrite your HTTP traffic using NGINX Gateway Fabric.

## Overview

[HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/) filters can be used to configure HTTP redirects or rewrites. Redirects return HTTP 3XX responses to a client, instructing it to retrieve a different resource. Rewrites modify components of a client request (such as hostname and/or path) before proxying it upstream.

In this guide, we will set up the coffee application to demonstrate path URL rewriting and the tea and soda applications to showcase path-based request redirection.. For an introduction to exposing your application, we recommend that you follow the [basic guide]({{< relref "/how-to/traffic-management/routing-traffic-to-your-app.md" >}}) first.

To see an example of a redirect using scheme and port, see the [HTTPS Termination]({{< relref "/how-to/traffic-management/https-termination.md" >}}) guide.

---

## Before you begin

- [Install]({{< relref "installation/" >}}) NGINX Gateway Fabric.
- Save the public IP address and port of NGINX Gateway Fabric into shell variables:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   GW_PORT=<port number>
   ```

{{< note >}}In a production environment, you should have a DNS record for the external IP address that is exposed, and it should refer to the hostname that the gateway will forward for.{{< /note >}}

---

## HTTP rewrites and redirects examples

We will configure a common gateway for the `URLRewrite` and `RequestRedirect` filter examples mentioned below.

---

### Deploy the Gateway API resources for the applications

The [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/) resource is typically deployed by the [Cluster Operator](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#roles-and-personas_1). This Gateway defines a single listener on port 80. Since no hostname is specified, this listener matches on all hostnames. To deploy the Gateway:

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

---

## URLRewrite example

This examples demonstrates how to rewrite the traffic uri for a simple coffee application. A HTTPRoute resource is used to define two `URLRewrite` filters that will rewrite requests. You can verify the server responds with the rewritten uri.

---

### Setup

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

---

### Configure a path rewrite

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

---

### Send traffic

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
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee/mocha\?test\=v1\&test\=v2
```

```text
Server address: 10.244.0.235:8080
Server name: coffee-6db967495b-twn6x
...
URI: /beans?test=v1&test=v2
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

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/latte/prices\?test\=v1\&test\=v2
```

```text
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/latte/prices\?test\=v1\&test\=v2
Server address: 10.244.0.235:8080
Server name: coffee-6db967495b-twn6x
...
URI: /prices?test=v1&test=v2
```

---

## RequestRedirect example

This example demonstrates how to redirect the traffic to a new location for a tea and soda applications. We'll examine how request redirects works using the `RequestRedirect` filter by creating four HTTPRoutes: two for redirecting requests to a new location and other two to define the destination location blocks for the redirected traffic.

---

### Setup

Create the **tea** and **soda** application in Kubernetes by copying and pasting the following block into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tea
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tea
  template:
    metadata:
      labels:
        app: tea
    spec:
      containers:
      - name: tea
        image: nginxdemos/nginx-hello:plain-text
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: tea
spec:
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: tea
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: soda
spec:
  replicas: 1
  selector:
    matchLabels:
      app: soda
  template:
    metadata:
      labels:
        app: soda
    spec:
      containers:
      - name: soda
        image: nginxdemos/nginx-hello:plain-text
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: soda
spec:
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: soda
EOF
```

This will create the **tea** and **soda** service and a deployment. Run the following command to verify the resources were created:

```shell
kubectl get pods,svc
```

Your output should include the **tea**, **soda** pod and the **tea**, **soda** service:

```text
NAME                          READY   STATUS      RESTARTS   AGE
pod/soda-7c76d95586-dnc4n     1/1     Running   0          89m
pod/tea-7b7d6c947d-s8djx      1/1     Running   0          120m


NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/soda         ClusterIP   10.96.230.208   <none>        80/TCP    89m
service/tea          ClusterIP   10.96.151.194   <none>        80/TCP    120m
```

---

### Configure a path redirect

We will define two HTTPRoutes for **tea** application: `tea`, which specifies the destination location block to handle redirected requests, and `tea-redirect` that redirect requests as follows:

- `http://cafe.example.com/tea` to `http://cafe.example.com/organic`
- `http://cafe.example.com/tea/origin` to `http://cafe.example.com/organic/origin`

The two HTTPRoutes defined for **soda** application: `soda`, which specifies the destination location block to handle redirected requests, and `soda-redirect` that redirect requests as follows:

- `http://cafe.example.com/soda` to `http://cafe.example.com/flavors`
- `http://cafe.example.com/soda/pepsi` to `http://cafe.example.com/flavors`

To create the httproute resource, copy and paste the following into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: tea-redirect
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
    - type: RequestRedirect
      requestRedirect:
        path:
          type: ReplacePrefixMatch
          replacePrefixMatch: /organic
        port: 8080
---
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
        value: /organic
    backendRefs:
    - name: tea
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: soda-redirect
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
        value: /soda
    filters:
    - type: RequestRedirect
      requestRedirect:
        path:
          type: ReplaceFullPath
          replaceFullPath: /flavors
        port: 8080
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: soda
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
        value: /flavors
    backendRefs:
    - name: soda
      port: 80
EOF
```

---

### Send traffic

Using the external IP address and port for NGINX Gateway Fabric, we can send traffic to our tea and soda application to verify the redirect is successful. We will use curl's `--include` option to print the response headers (we are interested in the `Location` header) and `-L` to follow redirects, ensuring that the user fetches the final destination after encountering HTTP 3xx redirect response.

{{< note >}}If you have a DNS record allocated for `cafe.example.com`, you can send the request directly to that hostname, without needing to resolve.{{< /note >}}

```shell
curl -L --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea --include
```

Notice in the output that the URI has been redirected to the new location:

```text
HTTP/1.1 302 Moved Temporarily
..
Location: http://cafe.example.com:8080/organic
```

Other examples:

```shell
curl -L --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea/type --include
```

```text
HTTP/1.1 302 Moved Temporarily
..
Location: http://cafe.example.com:8080/organic/type
```

```shell
curl -L --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea/type\?test\=v1 --includ
```

```text
HTTP/1.1 302 Moved Temporarily
..
Location: http://cafe.example.com:8080/organic/type?test=v1
```

```shell
curl -L --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/soda --include
```

```text
HTTP/1.1 302 Moved Temporarily
..
Location: http://cafe.example.com:8080/flavors
```

```shell
curl -L --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/soda/pepsi --include
```

```text
HTTP/1.1 302 Moved Temporarily
..
Location: http://cafe.example.com:8080/flavors
```

```shell
curl -L --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/soda/pepsi\?test\=v1 --include
```

```text
HTTP/1.1 302 Moved Temporarily
..
Location: http://cafe.example.com:8080/flavors?test=v1
```

---

## Further reading

To learn more about redirects and rewrites using the Gateway API, see the following resource:

- [Gateway API Redirects and Rewrites](https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/)
