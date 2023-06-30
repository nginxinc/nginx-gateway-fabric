# Example

In this example, we expand on the simple [cafe-example](../cafe-example) by using a ReferenceGrant to route to backends
in a different namespace from our HTTPRoutes.

## Running the Example

## 1. Deploy NGINX Kubernetes Gateway

1. Follow the [installation instructions](/docs/installation.md) to deploy NGINX Gateway.

1. Save the public IP address of NGINX Kubernetes Gateway into a shell variable:

   ```
   GW_IP=XXX.YYY.ZZZ.III
   ```

1. Save the port of NGINX Kubernetes Gateway:

   ```
   GW_PORT=<port number>
   ```

## 2. Deploy the Cafe Application

1. Create the cafe namespace and cafe application:

   ```
   kubectl apply -f cafe-ns-and-app.yaml
   ```

1. Check that the Pods are running in the `cafe` namespace:

   ```
   kubectl -n cafe get pods
   NAME                      READY   STATUS    RESTARTS   AGE
   coffee-6f4b79b975-2sb28   1/1     Running   0          12s
   tea-6fb46d899f-fm7zr      1/1     Running   0          12s
   ```

## 3. Configure Routing

1. Create the `Gateway`:

   ```
   kubectl apply -f gateway.yaml
   ```

1. Create the `HTTPRoute` resources:

   ```
   kubectl apply -f cafe-routes.yaml
   ```
1. Create the `ReferenceGrant`:

   ```
   kubectl apply -f reference-grant.yaml
   ```
   This ReferenceGrant allows all HTTPRoutes in the `default` namespace to reference all Services in the `cafe`
   namespace.

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `coffee` and `tea` Services.

To get coffee:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
Server address: 10.12.0.18:80
Server name: coffee-7586895968-r26zn
```

To get tea:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
Server address: 10.12.0.19:80
Server name: tea-7cd44fcb4d-xfw2x
```

## 5. Remove the ReferenceGrant

To restrict access to Services in the `cafe` Namespace, we can delete the ReferenceGrant we created in
Step 3:

```
kubectl delete -f reference-grant.yaml
```

Now, if we try to access the application over HTTPS, we will get an internal server error:
```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea

<html>
<head><title>500 Internal Server Error</title></head>
<body>
<center><h1>500 Internal Server Error</h1></center>
<hr><center>nginx/1.25.1</center>
</body>
</html>
```


You can also check the conditions of the HTTPRoutes `coffee` and `tea` to verify the that the reference is not permitted:

```
kubectl describe httproute coffee

Condtions:
      Message:               Backend ref to Service cafe/coffee not permitted by any ReferenceGrant
      Observed Generation:   1
      Reason:                RefNotPermitted
      Status:                False
      Type:                  ResolvedRefs
      Controller Name:       k8s-gateway.nginx.org/nginx-gateway-controller
```

```
kubectl describe httproute tea

Condtions:
      Message:               Backend ref to Service cafe/tea not permitted by any ReferenceGrant
      Observed Generation:   1
      Reason:                RefNotPermitted
      Status:                False
      Type:                  ResolvedRefs
      Controller Name:       k8s-gateway.nginx.org/nginx-gateway-controller
```
