# Advanced Routing

In this example we will deploy NGINX Kubernetes Gateway and configure advanced routing rules for a simple cafe application. 
We will use `HTTPRoute` resources to route traffic to the cafe application based on a combination of the request method, headers, and query parameters.

## Running the Example

## 1. Deploy NGINX Kubernetes Gateway

1. Follow the [installation instructions](https://github.com/nginxinc/nginx-kubernetes-gateway/blob/main/README.md#run-nginx-gateway) to deploy NGINX Gateway.

1. Save the public IP address of NGINX Kubernetes Gateway into a shell variable:
   
   ```
   GW_IP=XXX.YYY.ZZZ.III
   ```

1. Save the port of NGINX Kubernetes Gateway:
   
   ```
   GW_PORT=<port number>
   ```

## 2. Deploy the Cafe Application  

1. Create the coffee and the tea deployments and services:
   
   ```
   kubectl apply -f cafe.yaml
   ```

1. Check that the Pods are running in the `default` namespace:

   ```
   kubectl -n default get pods
   NAME                      READY   STATUS    RESTARTS   AGE
   coffee-6f4b79b975-2sb28   1/1     Running   0          12s
   tea-6fb46d899f-fm7zr      1/1     Running   0          12s
   ```

## 3. Configure Routing

1. Create the `HTTPRoute` resources:

   ```
   kubectl apply -f cafe-routes.yaml
   ```

## 4. Test the Application

We will use `curl` to send requests to the `coffee` and `tea` services.

### 4.1 Access coffee

Send a `POST` request to the path `/coffee` with the headers `X-Demo-Header:Demo-X1` and `version:v1`:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee -X POST -H "X-Demo-Header:Demo-X1" -H "version:v1"
Server address: 10.12.0.18:80
Server name: coffee-7586895968-r26zn
```

Header keys are case-insensitive, so we can also access coffee with the following request:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee -X POST -H "X-DEMO-HEADER:Demo-X1" -H "Version:v1"
Server address: 10.12.0.18:80
Server name: coffee-7586895968-r26zn
```

Only `POST` requests to the path `/coffee` with the headers `X-Demo-Header:Demo-X1` and `version:v1` will be able to access coffee.
For example, try sending the following `GET` request:
```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee -H "X-Demo-Header:Demo-X1" -H "version:v1"
```

NGINX Kubernetes Gateway returns a 405 since the request method does not match the method defined in the routing rule for `/coffee`.

### 4.2 Access tea

Send a request to the path `/tea` with the query parameter `Great=Example`:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea?Great=Example
Server address: 10.12.0.19:80
Server name: tea-7cd44fcb4d-xfw2x
```

Query parameters are case-sensitive, so the case must match what you specify in the `HTTPRoute` resource.

Only requests to the path `/tea` with the query parameter `Great=Example` will be able to access tea. 
For example, try sending the following request:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
```

NGINX Kubernetes Gateway returns a 404 since the request does not satisfy the routing rule configured for `/tea`.
