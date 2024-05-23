---
title: "Tracing"
description: "Learn how to configure tracing in NGINX Gateway Fabric."
weight: 200
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Overview

NGINX Gateway Fabric supports tracing using [OpenTelemetry](https://opentelemetry.io/). The official [NGINX OpenTelemetry Module](https://github.com/nginxinc/nginx-otel) instruments the NGINX data plane to export traces to a configured collector. Tracing data can be exported to an OpenTelemetry Protocol (OTLP) exporter, such as the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector). This collector can then export data to one or more upstream collectors like [Jaeger](https://www.jaegertracing.io/), [DataDog](https://docs.datadoghq.com/tracing/), and many others. This particular model is called the [Agent model](https://opentelemetry.io/docs/collector/deployment/agent/).

In this guide, we are going enable tracing on our HTTPRoutes using NGINX Gateway Fabric. We will use the OpenTelemetry Collector and Jaeger to process and collect our traces.

## Installing the Collectors

The first step is to install the collectors. NGINX Gateway Fabric will be configured to export to the OpenTelemetry Collector, which is configured to export to Jaeger. This model allows us to easily swap out the visualization collector (Jaeger) for something else if we want to, or add more collectors without needing to reconfigure NGINX Gateway Fabric. It is also possible to configure NGINX Gateway Fabric to export directly to Jaeger, if desired.

First, create the namespace:

```shell
kubectl create namespace monitoring
```

Download the following files containing the configurations for the collectors:

- {{< download "otel-collector.yaml" "otel-collector.yaml" >}}
- {{< download "jaeger.yaml" "jaeger.yaml" >}}

{{< note >}}These collectors are for demo purposes and are not tuned for production use.{{< /note >}}

and install:

```shell
kubectl apply -f otel-collector.yaml -f jaeger.yaml -n monitoring
```

Ensure that the Pods are running:

```shell
kubectl -n monitoring get pods
```

```text
NAME                             READY   STATUS    RESTARTS   AGE
jaeger-8469f69b86-bfpk9          1/1     Running   0          9s
otel-collector-f786b7dfd-h2x9l   1/1     Running   0          9s
```

Once running, you can access the Jaeger dashboard by using port-forwarding in the background:

```shell
kubectl port-forward -n monitoring svc/jaeger 16686:16686 &
```

Visit [http://127.0.0.1:16686](http://127.0.0.1:16686) to view the dashboard.

## Enabling Tracing

Enabling tracing requires two pieces of configuration. The first is a resource called `NginxProxy`, which contains global settings relating to the NGINX data plane. This resource is created and managed by the [cluster operator](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/), and is referenced in the `parametersRef` field of the GatewayClass. This resource can be created and linked when we install NGINX Gateway Fabric using its helm chart, or it can be added later. In this guide we will install the resource using the helm chart, but will also show what it looks like in case you want to add it after installation.

The `NginxProxy` resource contains configuration for the collector, and applies to all Gateways and routes under the GatewayClass. It does not enable tracing, but is a prerequisite to the next piece of configuration.

The second piece of configuration is the `ObservabilityPolicy`, which is a [Policy](https://gateway-api.sigs.k8s.io/reference/policy-attachment/) that targets HTTPRoutes or GRPCRoutes. This Policy is created by the [application developer](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/) and enables tracing for a specific route or routes. It requires the `NginxProxy` resource to exist in order to complete the tracing configuration.

TODO(sberman): link to reference docs

### Installing NGINX Gateway Fabric with global tracing config

{{< note >}}Ensure that you've already [installed the Gateway API resources]({{< relref "installation/installing-ngf/helm.md#installing-the-gateway-api-resources" >}}).{{< /note >}}

Based on the collector we deployed above, we'll create the following `values.yaml` file to install NGINX Gateway Fabric:

```yaml
cat <<EOT > values.yaml
nginx:
  config:
    telemetry:
      exporter:
        endpoint: otel-collector.tracing.svc:4317
      spanAttributes:
      - key: cluster-attribute-key
        value: cluster-attribute-value
EOT
```

We've set the endpoint and added a demo attribute that will be added to all tracing spans.

To install:

```shell
helm install ngf oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric --create-namespace -n nginx-gateway -f values.yaml
```

As a result, we should see the following configurations:

```shell
kubectl get nginxproxies.gateway.nginx.org ngf-proxy-config -o yaml
```

```yaml
apiVersion: gateway.nginx.org/v1alpha1
kind: NginxProxy
metadata:
  name: ngf-proxy-config
spec:
  telemetry:
    exporter:
      endpoint: otel-collector.tracing.svc:4317
    spanAttributes:
    - key: cluster-attribute-key
      value: cluster-attribute-value
```

```shell
kubectl get gatewayclasses.gateway.networking.k8s.io nginx -o yaml
```

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: nginx
spec:
  controllerName: gateway.nginx.org/nginx-gateway-controller
  parametersRef:
    group: gateway.nginx.org
    kind: NginxProxy
    name: ngf-proxy-config
status:
  conditions:
  - lastTransitionTime: "2024-05-22T15:18:35Z"
    message: GatewayClass is accepted
    observedGeneration: 1
    reason: Accepted
    status: "True"
    type: Accepted
  - lastTransitionTime: "2024-05-22T15:18:35Z"
    message: Gateway API CRD versions are supported
    observedGeneration: 1
    reason: SupportedVersion
    status: "True"
    type: SupportedVersion
  - lastTransitionTime: "2024-05-22T15:18:35Z"
    message: parametersRef resource is resolved
    observedGeneration: 1
    reason: ResolvedRefs
    status: "True"
    type: ResolvedRefs
```

If you already had NGINX Gateway Fabric installed, then you can simply create the `NginxProxy` resource and link it in the GatewayClass `parametersRef` like shown above, using:

```shell
kubectl edit gatewayclasses.gateway.networking.k8s.io nginx
```

Next you'll want to [Expose NGINX Gateway Fabric]({{< relref "installation/expose-nginx-gateway-fabric.md" >}}) and save the public IP address and port of NGINX Gateway Fabric into shell variables:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   GW_PORT=<port number>
   ```

Now we can create our application, route, and tracing policy.

### Create the application and route

Create the basic **coffee** application:

```yaml
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coffee
spec:
  replicas: 2
  selector:
    matchLabels:
      app: coffee
  template:
    metadata:
      labels:
        app: coffee
    spec:
      containers:
      - name: coffee
        image: nginxdemos/nginx-hello:plain-text
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: coffee
spec:
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: coffee
EOF
```

Next we'll create the Gateway resource and HTTPRoute for our app:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: cafe
spec:
  gatewayClassName: nginx
  listeners:
  - name: http
    port: 80
    protocol: HTTP
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: coffee
spec:
  parentRefs:
  - name: cafe
  hostnames:
  - "cafe.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /coffee
    backendRefs:
    - name: coffee
      port: 80
EOF
```

Let's ensure that traffic can flow to our application.

{{< note >}}If you have a DNS record allocated for `cafe.example.com`, you can send the request directly to that hostname, without needing to resolve.{{< /note >}}

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
```

We should see a response from the coffee Pod.

```text
Server address: 10.244.0.69:8080
Server name: coffee-6b8b6d6486-k5w5w
URI: /coffee
```

Assuming that you have access to the [Jaeger dashboard](http://127.0.0.1:16686) from earlier in the guide, you shouldn't see any tracing information yet. This means we need to create our `ObservabilityPolicy`.

### Create the ObservabilityPolicy

To enable tracing for our coffee HTTPRoute, we create the following policy:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.nginx.org/v1alpha1
kind: ObservabilityPolicy
metadata:
  name: coffee
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: coffee
  tracing:
    strategy: ratio
    ratio: 50
    spanAttributes:
    - key: coffee-key
      value: coffee-value
EOF
```

This policy attaches to the coffee HTTPRoute and enables ratio-based tracing, where 50% of requests will be sampled. We've also included a span attribute to add extra data to the spans.

Let's check the status of the policy:

```shell
kubectl describe observabilitypolicies.gateway.nginx.org coffee
```

```text
Status:
  Ancestors:
    Ancestor Ref:
      Group:      gateway.networking.k8s.io
      Kind:       HTTPRoute
      Name:       coffee
      Namespace:  default
    Conditions:
      Last Transition Time:  2024-05-23T18:13:03Z
      Message:               Policy is accepted
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
    Controller Name:         gateway.nginx.org/nginx-gateway-controller
```

The policy is accepted, so now let's send some more traffic.

```shell
for i in $(seq 1 10); do curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee; sleep 1; done
```

This will send 10 requests. Once complete, let's refresh the Jaeger dashboard. We should now see a service entry called `ngf:default:cafe`, and a few traces. The service name by default is `ngf:<gateway-namespace>:<gateway-name>`.

{{<img src="img/jaeger-trace-overview.png" alt="">}}

<br></br>

If we click into one of the traces, we can see the attributes.

{{<img src="img/jaeger-trace-attributes.png" alt="">}}

As you can see, the trace includes the attribute from the global NginxProxy resource, set by the cluster operator, as well as the attribute from the ObservabilityPolicy, set by the application developer.

## Further Reading

TODO(sberman): link to reference docs again
