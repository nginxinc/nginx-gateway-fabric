# Directing Traffic to Your Application

In this guide, you will learn how to direct external traffic to your Kubernetes applications using the Gateway API and
NGINX Kubernetes Gateway. Whether you're managing a web application or a REST backend API, you can use NGINX Kubernetes
Gateway to expose your application outside the cluster.

## Prerequisites

- NGINX Kubernetes Gateway is [installed](/docs/installation.md) on your cluster.
- [Expose NGINX Kubernetes Gateway](/docs/installation.md#expose-nginx-kubernetes-gateway) and save the public IP
  address and port of NGINX Kubernetes Gateway into shell variables:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   GW_PORT=<port number>
   ```

## The Application

The application we are going to use in this guide is a simple coffee application comprised of one Service and two Pods:

![coffee app](/docs/images/direct-all-traffic-app.png)

With this architecture, the coffee application is not accessible outside the cluster. We want to expose this application
on the hostname `cafe.example.com` so that users outside the cluster can access it.

To do this, we will install NGINX Kubernetes Gateway and create two Gateway API resources: a Gateway and an HTTPRoute.
With these resources, we will configure a simple routing rule to match all HTTP traffic with the
hostname `cafe.example.com` and direct it to the coffee Service.

## Setup

Create the coffee application in Kubernetes by copying and pasting the following into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coffee
spec:
  replicas: 2
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

This will create the coffee Service and a Deployment with two Pods. Run the following command to verify the resources
were created:

```shell
kubectl get pods,svc
```

Your output should include 2 coffee Pods and the coffee Service:

```text
NAME                          READY   STATUS      RESTARTS   AGE
pod/coffee-7dd75bc79b-cqvb7   1/1     Running     0          77s
pod/coffee-7dd75bc79b-dett3   1/1     Running     0          77s


NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/coffee       ClusterIP   10.96.75.77     <none>        80/TCP    77s
```

## Application Architecture with NGINX Kubernetes Gateway

To direct traffic to the coffee application, we will create a Gateway and HTTPRoute. The following diagram shows the
configuration we'll be creating in the next step:

![Configuration](/docs/images/direct-all-traffic-config.png)

We need a Gateway to create an access point for HTTP traffic coming into the cluster. The `cafe` Gateway we are going to
create will open an access point to the cluster on port 80 for HTTP traffic.

To direct HTTP traffic from the Gateway to the coffee Service, we need to create an HTTPRoute named `coffee` and attach
to the Gateway. This HTTPRoute will have a single routing rule that directs all traffic to the
hostname `cafe.example.com` from the Gateway to the coffee Service.

Once NGINX Kubernetes Gateway parses the `cafe` Gateway and `coffee` HTTPRoute, it will configure its dataplane, NGINX,
to route all HTTP requests to `cafe.example.com` to the Pods that the `coffee` Service targets:

![Traffic Flow](/docs/images/direct-all-traffic-flow.png)

The coffee Service is omitted from the diagram above because the NGINX Kubernetes Gateway routes directly to the Pods
that the coffee Service targets.

> **Note**
> In the diagrams above, all resources that are the responsibility of the cluster operator are shown in green.
> The orange resources are the responsibility of the application developers.
> See the [roles and personas](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#roles-and-personas_1)
> Gateway API document for more information on these roles.

## Create the Gateway API Resources

To create the `cafe` Gateway, copy and paste the following into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1beta1
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

This Gateway is associated with the NGINX Kubernetes Gateway through the `gatewayClassName` field. The default
installation of NGINX Kubernetes Gateway creates a GatewayClass with the name `nginx`. NGINX Kubernetes Gateway will
only configure Gateways with a `gatewayClassName` of `nginx` unless you change the name via the `--gatewayclass`
[command-line flag](/docs/cli-help.md#static-mode).

We specify a listener on the Gateway to open an access point on the cluster. In this case, since the coffee application
accepts HTTP requests, we create an HTTP listener, named `http`, that listens on port 80.

By default, Gateways only allow routes (such as HTTPRoutes) to attach if they are in the same namespace as the Gateway.
If you want to change this behavior, you can set
the [`allowedRoutes`](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.AllowedRoutes)
field.

Now, let's create the HTTPRoute by copying and pasting the following into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: coffee
spec:
  parentRefs:
  - name: cafe
  hostnames:
  - "cafe.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: coffee
      port: 80
EOF
```

To attach the `coffee` HTTPRoute to the `cafe` Gateway, we specify the Gateway name in the `parentRefs` field. The
attachment will succeed if the hostnames and protocol in the HTTPRoute are allowed by at least one of the
Gateway's listeners.

The `hostnames` field allows you to list the hostnames that the HTTPRoute matches. In this case, incoming requests
handled by the `http` listener with the HTTP host header `cafe.example.com` will match this HTTPRoute and will be routed
according to the rules in the spec.

The `rules` field defines routing rules for the HTTPRoute. A rule is selected if the request satisfies one of the
rule's `matches`. To forward traffic for all paths to the coffee Service we specify a match with the PathPrefix `/` and
target the coffee Service using the `backendRef` field.

## Test the Configuration

To test the configuration, we will send a request to the public IP and port of NGINX Kubernetes Gateway that you saved
in the [prerequisites](#prerequisites) section and verify that the response comes from one of the coffee Pods.

> **Note**
> In a production environment, the coffee application would be accessible through the hostname `cafe.example.com`.
> However, in this guide, we will be accessing the application through the public IP and port of the NGINX Kubernetes
> Gateway Service and with curl's `--resolve` option.


First, let's send a request to the path `/`:

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/
```

We should get a response from one of the coffee Pods:

```text
Server address: 10.12.0.18:8080
Server name: coffee-7dd75bc79b-cqvb7
```

Since the `cafe` HTTPRoute routes all traffic on any path to the coffee application, the following requests should also
be handled by the coffee Pods:

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/some-path
```

```text
Server address: 10.12.0.18:8080
Server name: coffee-7dd75bc79b-cqvb7
```

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/some/path
```

```text
Server address: 10.12.0.19:8080
Server name: coffee-7dd75bc79b-dett3
```

## Troubleshooting

If you have any issues while testing the configuration, try the following to debug your configuration and setup:

- Make sure you set the shell variables $GW_IP and $GW_PORT to the public IP and port of the NGINX Kubernetes Gateway
  Service. Instructions for finding those values are [here](/docs/installation.md#expose-nginx-kubernetes-gateway).

- Check the status of the Gateway:

  ```shell
  kubectl describe gateway cafe
  ```

  The Gateway status should look similar to this:

  ```text
  Status:
  Addresses:
    Type:   IPAddress
    Value:  10.244.0.85
  Conditions:
    Last Transition Time:  2023-08-15T20:57:21Z
    Message:               Gateway is accepted
    Observed Generation:   1
    Reason:                Accepted
    Status:                True
    Type:                  Accepted
    Last Transition Time:  2023-08-15T20:57:21Z
    Message:               Gateway is programmed
    Observed Generation:   1
    Reason:                Programmed
    Status:                True
    Type:                  Programmed
  Listeners:
    Attached Routes:  1
    Conditions:
      Last Transition Time:  2023-08-15T20:57:21Z
      Message:               Listener is accepted
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2023-08-15T20:57:21Z
      Message:               Listener is programmed
      Observed Generation:   1
      Reason:                Programmed
      Status:                True
      Type:                  Programmed
      Last Transition Time:  2023-08-15T20:57:21Z
      Message:               All references are resolved
      Observed Generation:   1
      Reason:                ResolvedRefs
      Status:                True
      Type:                  ResolvedRefs
      Last Transition Time:  2023-08-15T20:57:21Z
      Message:               No conflicts
      Observed Generation:   1
      Reason:                NoConflicts
      Status:                False
      Type:                  Conflicted
    Name:                    http
  ```

  Check that the conditions match and that the attached routes for the `http` listener equals 1. If it is 0, there may
  be an issue with the HTTPRoute.

- Check the status of the HTTPRoute:

  ```shell
  kubectl describe httproute coffee
  ```

  The HTTPRoute status should look similar to this:

  ```text
  Status:
    Parents:
      Conditions:
        Last Transition Time:  2023-08-15T20:57:21Z
        Message:               The route is accepted
        Observed Generation:   1
        Reason:                Accepted
        Status:                True
        Type:                  Accepted
        Last Transition Time:  2023-08-15T20:57:21Z
        Message:               All references are resolved
        Observed Generation:   1
        Reason:                ResolvedRefs
        Status:                True
        Type:                  ResolvedRefs
      Controller Name:         gateway.nginx.org/nginx-gateway-controller
      Parent Ref:
        Group:      gateway.networking.k8s.io
        Kind:       Gateway
        Name:       cafe
        Namespace:  default
  ```

  Check for any error messages in the conditions.

- Check the generated nginx config:

  ```shell
  kubectl exec -it -n nginx-gateway <nginx gateway Pod> -c nginx -- nginx -T
  ```

  The config should contain the following config blocks:

  ```nginx configuration
  upstream default_coffee_80 {
    random two least_conn;
    zone default_coffee_80 512k;

    server 10.12.0.18:8080; # these should be the Pod IPs of the coffee Pods
    server 10.12.0.19:8080;
  }
  ...

  server {
    listen 80;

    server_name cafe.example.com;

    location / {

        proxy_set_header Host $gw_api_compliant_host;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_pass http://default_coffee_80$request_uri;
    }
  }
  ```

If your issue persists, open an [issue](https://github.com/nginxinc/nginx-kubernetes-gateway/issues/new/choose) in the
NGINX Kubernetes Gateway repo.

## Further Reading

To learn more about the Gateway API and the resources we created in this guide, check out the following resources:

- [Gateway API Overview](https://gateway-api.sigs.k8s.io/concepts/api-overview/)
- [Deploying a simple Gateway](https://gateway-api.sigs.k8s.io/guides/simple-gateway/)
- [HTTP Routing](https://gateway-api.sigs.k8s.io/guides/http-routing/)
