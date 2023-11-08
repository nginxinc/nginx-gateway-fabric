# Example

In this example we deploy NGINX Gateway Fabric, a simple web application, configure NGINX Gateway Fabric
to route traffic to that application using HTTPRoute resources, and configure NGINX Gateway Fabric to enable tracing on
all requests using the [NGINX OTel module](https://nginx.org/en/docs/ngx_otel_module.html).

## Running the Example

## Prerequisites

### Deploy Jaegar operator and instance

1. Follow the instructions to deploy [Jaegar operator](https://www.jaegertracing.io/docs/latest/operator/) into your
   environment.

1. Deploy a simple [all-in-one instance](https://www.jaegertracing.io/docs/latest/operator/#quick-start---deploying-the-allinone-image).
   A sample manifest is provided:

   ```shell
   kubectl apply -f jaegar.yaml
   ```

1. Get the local DNS name of the all in one Jaegar collector service. It will follow the standard
   [Kubernetes DNS naming conventions](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#a-aaaa-records)
   e.g. `simplest-collector.default.svc.cluster.local`.
   Also get the number of either the `grpc-otlp` port or the `http-otlp` port of the collector service, e.g.

   ```shell
   k describe svc simplest-collector
   Name:              simplest-collector
   Namespace:         default
   <...>
   Port:              grpc-otlp  4317/TCP
   TargetPort:        4317/TCP
   Endpoints:         10.244.0.16:4317
   Port:              http-otlp  4318/TCP
   TargetPort:        4318/TCP
   Endpoints:         10.244.0.16:4318
   <...>
   ```

## 1. Deploy NGINX Gateway Fabric

1. Follow the [installation instructions](/docs/installation.md) to deploy NGINX Gateway Fabric.

1. Save the public IP address of NGINX Gateway Fabric into a shell variable:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   ```

1. Save the port of NGINX Gateway Fabric:

   ```text
   GW_PORT=<port number>
   ```

## 2. Add trace configuration to NginxProxy

> For this step, the name `nginx-proxy-config` should be updated to the name of the resource that
> was created by your installation.

1. Update the configuration to add the tracing configuration.

   ```shell
   kubectl -n nginx-gateway edit nginxproxies nginx-proxy-config
   ```

   This will open the configuration in your default editor. You can then update and save the configuration, which is
   applied automatically to the data plane. The `endpoint` value should be the Service name and the
   `grpc-otlp` port or the `http-otlp` port from earlier concatenated with a colon,
   e.g. `simplest-collector.default.svc.cluster.local:4317`.

   ```yaml
   <...>
   spec:
      http:
         telemetry:
            tracing:
            enabled: true # default false
            endpoint: simplest-collector.default.svc.cluster.local:4317 # required
            interval: "5s" # default
            batchSize: 512 # default
            batchCount: 4 # default
   <...>
   ```

2. View the status of the configuration.

   ```shell
   kubectl -n nginx-gateway describe nginxproxies nginx-proxy-config
   ```

   You should see something like the following:

   ```yaml
   Name:         nginx-proxy-config
   Namespace:    nginx-gateway
   <...>
   Spec:
   Http:
      Telemetry:
         Tracing:
         Batch Count:  4
         Batch Size:   512
         Enabled:      true
         Endpoint:     simplest-collector.default.svc.cluster.local:4317
         Interval:     5s
   Status:
   Conditions:
      Last Transition Time:  2023-11-08T16:37:17Z
      Message:               NginxProxy is accepted
      Observed Generation:   2
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2023-11-08T16:37:17Z
      Message:               NginxProxy is programmed
      Observed Generation:   2
      Reason:                Programmed
      Status:                True
      Type:                  Programmed
   Events:                    <none>
   ```


## 2. Deploy the Cafe Application

1. Create the coffee and the tea Deployments and Services:

   ```shell
   kubectl apply -f cafe.yaml
   ```

1. Check that the Pods are running in the `default` Namespace:

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

1. Create the HTTPRoute resources:

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

## 5. View tracing

Log into your Jaegar instance and view the traces for your requests. The service will follow the naming pattern
`gateway-class-name:ngf`, so for this example it will be `nginx:ngf`.
