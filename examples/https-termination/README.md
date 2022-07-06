# HTTPS Termination Example

In this example we expand on the simple [cafe-example](../cafe-example) by adding HTTPS termination to our routes.

## Running the Example

## 1. Deploy NGINX Kubernetes Gateway

1. Follow the [installation instructions](https://github.com/nginxinc/nginx-kubernetes-gateway/blob/main/README.md#run-nginx-gateway) to deploy NGINX Gateway.

1. Save the public IP address of NGINX Kubernetes Gateway into a shell variable:
   
   ```
   GW_IP=XXX.YYY.ZZZ.III
   ```

1. Save the HTTPS port of NGINX Kubernetes Gateway:
   
   ```
   GW_HTTPS_PORT=port
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

## 3. Configure HTTPS Termination and Routing

HTTPS termination is configured at the gateway level with listeners. You created the following gateway resource in step 1:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
  name: gateway
  namespace: nginx-gateway
  labels:
    domain: k8s-gateway.nginx.org
spec:
  gatewayClassName: nginx
  listeners:
  - name: http
    port: 80
    protocol: HTTP
  - name: https
    port: 443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: default-server-secret
        namespace: nginx-gateway
``` 

The `https` listener is configured to terminate TLS connections using the `default-server-secret` in the `nginx-gateway` namespace.
To configure HTTPS termination for our cafe application, we will bind the `https` listener to our `HTTPRoutes` in [cafe-routes.yaml](./cafe-routes.yaml) using the [`parentRef`](https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io%2fv1alpha2.ParentReference) field:

```yaml
  parentRefs:
  - name: gateway
    namespace: nginx-gateway
    sectionName: https
```

1. Create the `HTTPRoute` resources:
   
   ```
   kubectl apply -f cafe-routes.yaml
   ```

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `coffee` and `tea` services.
Since our certificate is self-signed, we'll use curl's `--insecure` option to turn off certificate verification.

To get coffee:

```
curl --resolve cafe.example.com:$GW_HTTPS_PORT:$GW_IP https://cafe.example.com:$GW_HTTPS_PORT/coffee --insecure
Server address: 10.12.0.18:80
Server name: coffee-7586895968-r26zn
```

To get tea:

```
curl --resolve cafe.example.com:$GW_HTTPS_PORT:$GW_IP https://cafe.example.com:$GW_HTTPS_PORT/tea --insecure
Server address: 10.12.0.19:80
Server name: tea-7cd44fcb4d-xfw2x
```
