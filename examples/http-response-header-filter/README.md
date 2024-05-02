# Example

In this example we will deploy NGINX Gateway Fabric and configure traffic routing.
We will use HTTPRoute resources to route traffic to the server, using the `ResponseHeaderModifier` filter to modify
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
   kubectl apply -f http-route.yaml
   ```

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `headers` endpoint.


Notice our configured header values can be seen in the `responseHeaders` section below, and that the `X-Server-Version` header is absent. The header `My-cool-header` gets the appended with value `this-is-the-appended-value` from the `responseHeaderModifier` filter and the value of header `Response-Overwrite-Header` gets overwritten to `overwritten-value` as defined in the *HttpRoute*.

```shell
curl -v -i --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/headers
```

```text
HTTP/1.1 200 OK
Server: nginx/1.25.5
Date: Thu, 02 May 2024 20:58:54 GMT
Content-Type: text/plain
Content-Length: 2
Connection: keep-alive
X-Custom-Header: this-stays-unchanged
My-cool-header: this-is-the-appended-value
My-cool-header: this-is-the-value
Response-Overwrite-Header: overwritten-value
```
