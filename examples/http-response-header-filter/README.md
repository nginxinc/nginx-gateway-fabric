# Example

In this example we will deploy NGINX Gateway Fabric and configure traffic routing for a simple echo server.
We will use HTTPRoute resources to route traffic to the echo server, using the `RequestHeaderModifier` filter to modify
the response headers.

## Running the Example

## 1. Deploy NGINX Gateway Fabric

1. Follow the [installation instructions](https://docs.nginx.com/nginx-gateway-fabric/installation/) to deploy NGINX Gateway Fabric.

1. Save the public IP address of NGINX Gateway Fabric into a shell variable:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   ```

1. Save the port of NGINX Gateway Fabric:

   ```text
   GW_PORT=<port number>
   ```

## 2. Deploy the Headers Application

1. Create the headers Deployment and Service:

   ```shell
   kubectl apply -f headers.yaml
   ```

1. Check that the Pod is running in the `default` Namespace:

   ```shell
   kubectl -n default get pods
   ```

   ```text
   NAME                      READY   STATUS    RESTARTS   AGE
   headers-6f4b79b975-2sb28   1/1     Running   0          12s
   ```

## 3. Configure Routing

1. Create the Gateway:

   ```shell
   kubectl apply -f gateway.yaml
   ```

1. Create the HTTPRoute resources:

   ```shell
   kubectl apply -f echo-route.yaml
   ```

This HTTPRoute has a few important properties:

- The `parentRefs` references the gateway resource that we created, and specifically defines the `http` listener to attach to, via the `sectionName` field.
- `echo.example.com` is the hostname that is matched for all requests to the backends defined in this HTTPRoute.
- The `match` rule defines that all requests with the path prefix `/headers` are sent to the `headers` Service.
- There is `ResponseHeaderModifier` filter defined for the path prefix `/headers` to set header `Response-Overwrite-Header` and add header `My-cool-header`.

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `headers` Service, including sending headers with our request.


Notice our configured header values can be seen in the `responseHeaders` section below, and that the `Header-to-remove` header is absent. The header `My-cool-header` gets the appended with value `respond-with-this` from the `responseHeaderModifier` filter and the value of header `Response-Overwrite-Header` gets overwritten as defined in the *HttpRoute*.

```shell
curl -s -i --resolve echo.example.com:$GW_PORT:$GW_IP http://echo.example.com:$GW_PORT/headers -H "Header-to-remove:remove-this-header"
```

```text
HTTP/1.1 200 OK
Server: nginx/1.25.5
Date: Thu, 02 May 2024 01:11:37 GMT
Content-Type: text/plain
Content-Length: 2
Connection: keep-alive
My-cool-header: this-is-the-value
My-cool-header: respond-with-this
Response-Overwrite-Header: overwritten-value
```
