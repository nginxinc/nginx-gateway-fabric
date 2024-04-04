# Example

In this example we deploy NGINX Gateway Fabric, two simple web applications named "coffee" and "tee",
and then configure NGINX Gateway Fabric to route traffic to those application using HTTPRoute resources.
In addition, we will specify a mirror filter for the `/coffee` route of the coffee application and direct it
to mirror incoming requests to the `tea` application.

## Running the Example

## 1. Deploy NGINX Gateway Fabric

1. Follow the [installation instructions](https://docs.nginx.com/nginx-gateway-fabric/installation/) to deploy NGINX Gateway Fabric.

2. Save the public IP address of NGINX Gateway Fabric into a shell variable:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   ```

3. Save the port of NGINX Gateway Fabric:

   ```text
   GW_PORT=<port number>
   ```

## 2. Deploy the Cafe Application

1. Create the coffee and the tea Deployments and Services:

   ```shell
   kubectl apply -f cafe.yaml
   ```

2. Check that the Pods are running in the `default` Namespace:

   ```shell
   kubectl -n default get pods
   ```

   ```text
   NAME                      READY   STATUS    RESTARTS   AGE
   coffee-6f4b79b975-2sb28   1/1     Running   0          12s
   tea-6fb46d899f-fm7zr      1/1     Running   0          12s
   ```

## 3. Configure Routing

1. Create the Gateway:

   ```shell
   kubectl apply -f gateway.yaml
   ```

2. Create the HTTPRoute resources:

   ```shell
   kubectl apply -f cafe-routes.yaml
   ```

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `coffee` and `tea` Services.

To get coffee:

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
```

```text
Server address: 10.12.0.18:80
Server name: coffee-7586895968-r26zn
```

To get tea:

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
```

```text
Server address: 10.12.0.19:80
Server name: tea-7cd44fcb4d-xfw2x
```

To validate that calling coffee also calls tee (mirroring), check the logs of the tea pod:

- first call coffee:

  ```shell
   curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
   ```

- then check the logs of tee pod to see that the request was indeed mirrored to the tea pod:

  ```shell
   kubectl logs tea-6fb46d899f-fm7zr
   ```

## 5. Using different hostnames

Traffic is allowed to `cafe.example.com` because the Gateway listener's hostname allows `*.example.com`. You can
change an HTTPRoute's hostname to something that matches this wildcard and still pass traffic.

For example, run the following command to open your editor and change the HTTPRoute's hostname to `foo.example.com`.

```shell
kubectl -n default edit httproute tea
```

Once changed, update the `curl` command above for the `tea` service to use the new hostname. Traffic should still pass
successfully.

Likewise, if you change the Gateway listener's hostname to something else, you can prevent the HTTPRoute's traffic from
passing successfully.

For example, run the following to open your editor and change the Gateway listener's hostname to `bar.example.com`:

```shell
kubectl -n default edit gateway gateway
```

Once changed, try running the same `curl` requests as above. They should be denied with a `404 Not Found`.
