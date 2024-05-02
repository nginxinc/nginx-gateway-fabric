# Example

In this example we will deploy NGINX Gateway Fabric and configure traffic routing for a simple echo server.
We will use HTTPRoute resources to route traffic to the echo server, using the `RequestHeaderModifier` filter to modify
headers to the request.

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

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `headers` Service, including sending headers with
our request.
Notice our configured header values can be seen in the `requestHeaders` section below, and that the `User-Agent` header
is absent.

```shell
curl -s --resolve echo.example.com:$GW_PORT:$GW_IP http://echo.example.com:$GW_PORT/headers -H "My-Cool-Header:my-client-value" -H "My-Overwrite-Header:dont-see-this"
```

```text
Headers:
  header 'Accept-Encoding' is 'compress'
  header 'My-cool-header' is 'my-client-value, this-is-an-appended-value'
  header 'My-Overwrite-Header' is 'this-is-the-only-value'
  header 'Host' is 'echo.example.com:$GW_PORT'
  header 'X-Forwarded-For' is '$GW_IP'
  header 'Connection' is 'close'
  header 'Accept' is '*/*'
```
