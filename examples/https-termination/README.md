# HTTPS Termination Example

In this example we expand on the simple [cafe-example](../cafe-example) by adding HTTPS termination to our routes and an HTTPS redirect from port 80 to 443.

## Running the Example

## 1. Deploy NGINX Kubernetes Gateway

1. Follow the [installation instructions](/docs/installation.md) to deploy NGINX Gateway.

1. Save the public IP address of NGINX Kubernetes Gateway into a shell variable:

   ```
   GW_IP=XXX.YYY.ZZZ.III
   ```

1. Save the ports of NGINX Kubernetes Gateway:

   ```
   GW_HTTP_PORT=<http port number>
   GW_HTTPS_PORT=<https port number>
   ```

## 2. Deploy the Cafe Application

1. Create the coffee and the tea Deployments and Services:

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

1. Create a Secret with a TLS certificate and key:
   ```
   kubectl apply -f cafe-secret.yaml
   ```

   The TLS certificate and key in this Secret are used to terminate the TLS connections for the cafe application.
   **Important**: This certificate and key are for demo purposes only.

1. Create the `Gateway` resource:
   ```
   kubectl apply -f gateway.yaml
   ```

   This [Gateway](./gateway.yaml) configures:
   * `http` listener for HTTP traffic
   * `https` listener for HTTPS traffic. It terminates TLS connections using the `cafe-secret` we created in the step 1.

1. Create the `HTTPRoute` resources:
   ```
   kubectl apply -f cafe-routes.yaml
   ```

   To configure HTTPS termination for our cafe application, we will bind our `coffee` and `tea` HTTPRoutes to the `https` listener in [cafe-routes.yaml](./cafe-routes.yaml) using the [`parentReference`](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.ParentReference) field:

   ```yaml
   parentRefs:
   - name: gateway
     sectionName: https
   ```

   To configure an HTTPS redirect from port 80 to 443, we will bind the special `cafe-tls-redirect` HTTPRoute with a [`HTTPRequestRedirectFilter`](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRequestRedirectFilter) to the `http` listener:

   ```yaml
   parentRefs:
   - name: gateway
     sectionName: http
   ```

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `coffee` and `tea` Services. First, we will access the application over HTTP to test that the HTTPS redirect works. Then we will use HTTPS.

### 4.1 Test HTTPS Redirect

To test that NGINX sends an HTTPS redirect, we will send requests to the `coffee` and `tea` Services on HTTP port. We will use curl's `--include` option to print the response headers (we are interested in the `Location` header).

To get a redirect for coffee:
```
curl --resolve cafe.example.com:$GW_HTTP_PORT:$GW_IP http://cafe.example.com:$GW_HTTP_PORT/coffee --include
HTTP/1.1 302 Moved Temporarily
...
Location: https://cafe.example.com:443/coffee
...
```

To get a redirect for tea:
```
curl --resolve cafe.example.com:$GW_HTTP_PORT:$GW_IP http://cafe.example.com:$GW_HTTP_PORT/tea --include
HTTP/1.1 302 Moved Temporarily
...
Location: https://cafe.example.com:443/tea
...
```

### 4.2 Access Coffee and Tea 

Now we will access the application over HTTPS. Since our certificate is self-signed, we will use curl's `--insecure` option to turn off certificate verification.

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
