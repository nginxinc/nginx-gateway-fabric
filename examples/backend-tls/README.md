# Backend TLS Policy Example

In this example, we will create a Backend TLS Policy, attach it to our application service, and then configure routing
rules. The Backend TLS Policy will be picked up by NGF and the connection between NGF and the upstream server will use
HTTPS.

## Running the Example

## 1. Deploy NGINX Gateway Fabric

1. Follow the [installation instructions](https://docs.nginx.com/nginx-gateway-fabric/installation/) to deploy NGINX Gateway Fabric.
   Please note that the Gateway APIs from the experimental channel are required, and NGF must be deployed with the
   `- --experimental-features-enable` flag.

2. Save the public IP address of NGINX Gateway Fabric into a shell variable:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   ```

3. Save the HTTPS port of NGINX Gateway Fabric:

   ```text
   GW_HTTPS_PORT=<https port number>
   ```

## 2. Deploy the secure-app Application

1. Create the secure-app Deployment and Service:

   ```shell
   kubectl apply -f secure-app.yaml
   ```

1. Check that the Pods are running in the `default` namespace:

   ```shell
   kubectl -n default get pods
   ```

   ```text
   NAME                         READY   STATUS    RESTARTS   AGE
   secure-app-575785644-b6nwh   1/1     Running   0          5s
   ```

## 3. Deploy the Backend TLS Policy

1. Create the ConfigMap that holds the `ca.crt` entry for verifying our self-signed certificates:

    ```shell
   kubectl apply -f backend-certs-configmap.yaml
   ```

2. Create the Backend TLS Policy which targets our `secure-app` Service and refers to our ConfigMap created in the
   previous step:

   ```shell
   kubectl apply -f policy.yaml
   ```

## 3. Configure HTTP Termination and Routing

1. Create the Gateway resource:

   ```shell
   kubectl apply -f gateway.yaml
   ```

   This [Gateway](./gateway.yaml) configures:
    - `http` listener for HTTP traffic
    - `https` listener for HTTPS traffic. It terminates TLS connections using the `app-secret` we created in step 1.

2. Create the HTTPRoute resources:

   ```shell
   kubectl apply -f secure-app-routes.yaml
   ```

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `secure-app` Service over HTTP.

```shell
curl --resolve secure-app.example.com:$GW_PORT:$GW_IP http://secure-app.example.com:$GW_PORT/
```

```text
hello from pod secure-app-575785644-749tq
```
