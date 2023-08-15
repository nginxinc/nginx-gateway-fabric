# Using NGINX Kubernetes Gateway to Upgrade Applications without Downtime

This guide explains how to use NGINX Kubernetes Gateway to upgrade applications without downtime.

Multiple upgrade methods are mentioned, assuming existing familiarity: this guide focuses primarily on how to use NGINX Kubernetes Gateway to accomplish them.

> See the [Architecture document](/docs/architecture.md) to learn more about NGINX Kubernetes Gateway architecture.

## NGINX Kubernetes Gateway Functionality

To understand the upgrade methods, you should be aware of the NGINX features that help prevent application downtime: graceful configuration reloads and upstream servers updates.

### Graceful Configuration Reloads

If a relevant Gateway API or built-in Kubernetes resource is changed, NGINX Kubernetes Gateway will update NGINX by regenerating the NGINX configuration. 
NGINX Kubernetes Gateway then sends a reload signal to the master NGINX process to apply the new configuration. 

We call such an operation a reload, during which client requests are not dropped - which defines it as a graceful reload.

This process is further explained in [NGINX's documentation](https://nginx.org/en/docs/control.html?#reconfiguration).

### Upstream Servers Updates

Endpoints frequently change during application upgrades: Kubernetes creates Pods for the new version of an application
and removes the old ones, creating and removing the respective Endpoints as well.

NGINX Kubernetes Gateway detects changes to Endpoints by watching their corresponding [EndpointSlices][endpoint-slices].

[endpoint-slices]:https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/

In NGINX configuration, a Service is represented as an [upstream][upstream], and an Endpoint as an
[upstream server][upstream-server].

[upstream]:https://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream

[upstream-server]:https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server

Two common cases are adding and removing Endpoints:

- If an Endpoint is added, NGINX Kubernetes Gateway adds an upstream server to NGINX that corresponds to the Endpoint,
  then reload NGINX. After that, NGINX will start proxying traffic to that Endpoint.
- If an Endpoint is removed, NGINX Kubernetes Gateway removes the corresponding upstream server from NGINX. After
  a reload, NGINX will stop proxying traffic to it. However, it will finish proxying any pending requests to that
  server before switching to another Endpoint.

As long as you have more than one ready Endpoint, the clients should not experience any downtime during upgrades.

> It is good practice to configure a [Readiness probe][readiness-probe] in the Deployment so that a Pod can advertise
> when it is ready to receive traffic. Note that NGINX Kubernetes Gateway will not add any Endpoint to NGINX that is not
> ready.

[readiness-probe]:https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/

## Before You Begin

For the upgrade methods covered in the next sections, we make the following assumptions:

- You deploy your application as a [Deployment][deployment].
- The Pods of the Deployment belong to a [Service][service] so that Kubernetes creates an [Endpoint][endpoints] for
  each Pod.
- You expose the application to the clients via an [HTTPRoute][httproute] resource that references that Service.

[deployment]:https://kubernetes.io/docs/concepts/workloads/controllers/deployment/

[service]:https://kubernetes.io/docs/concepts/services-networking/service/

[httproute]:https://gateway-api.sigs.k8s.io/api-types/httproute/

[endpoints]:https://kubernetes.io/docs/reference/kubernetes-api/service-resources/endpoints-v1/

For example, an application can be exposed using a routing rule like below:

```yaml
- matches:
  - path:
      type: PathPrefix
      value: /
  backendRefs:
  - name: my-app
    port: 80
```

> See the [Cafe example](/examples/cafe-example) for a basic example.

The upgrade methods in the next sections cover:

- Rolling Deployment Upgrades
- Blue-green Deployments
- Canary Releases

## Rolling Deployment Upgrade

To start a [rolling Deployment upgrade][rolling-upgrade], you update the Deployment to use the new version tag of
the application. As a result, Kubernetes terminates the Pods with the old version and create new ones. By default,
Kubernetes also ensures that some number of Pods always stay available during the upgrade.

[rolling-upgrade]:https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment

Such an upgrade will add new upstream servers to NGINX and remove the old ones. As long as the number
of Pods (ready Endpoints) during an upgrade does not reach zero, NGINX will be able to proxy traffic, and thus prevent
any downtime.

This method does not require you to update the HTTPRoute.

## Blue-Green Deployments

With this method, you deploy a new version of the application (blue version) as a separate Deployment,
while the old version (green) keeps running and handling client traffic. Next, you switch the traffic from the
green version to the blue. If the blue works as expected, you terminate the green. Otherwise, you switch the traffic
back to the green.

There are two ways to switch the traffic:

- Update the Service selector to select the Pods of the blue version instead of the green. As a result, NGINX Kubernetes
  Gateway removes the green upstream servers from NGINX and add the blue ones. With this approach, it is not
  necessary to update the HTTPRoute.
- Create a separate Service for the blue version and update the backend reference in the HTTPRoute to reference this
  Service, which leads to the same result as with the previous option.

## Canary Releases

To support canary releases, you can implement an approach with two Deployments behind the same Service (see
[Canary deployment][canary] in the Kubernetes documentation). However, this approach lacks precision for defining the
traffic split between the old and the canary version. You can greatly influence it by controlling the number of Pods
(for example, four Pods of the old version and one Pod of the canary). However, note that NGINX Kubernetes Gateway uses
[`random two least_conn`][random-method] load balancing method, which doesn't guarantee an exact split based on the
number of Pods (80/20 in the given example).

[canary]:https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#canary-deployment
[random-method]:https://nginx.org/en/docs/http/ngx_http_upstream_module.html#random

A more flexible and precise way to implement canary releases is to configure a traffic split in an HTTPRoute. In this
case, you create a separate Deployment for the new version with a separate Service. For example, for the rule below,
NGINX will proxy 95% of the traffic to the old version Endpoints and only 5% to the new ones.

```yaml
- matches:
  - path:
      type: PathPrefix
      value: /
  backendRefs:
  - name: my-app-old
    port: 80
    weight: 95
  - name: my-app-new
    port: 80
    weight: 5
```

> There is no stickiness for the requests from the same client. NGINX will independently split each request among
> the backend references.

By updating the rule you can further increase the share of traffic the new version gets and finally completely switch
to the new version:

```yaml
- matches:
  - path:
      type: PathPrefix
      value: /
  backendRefs:
  - name: my-app-old
    port: 80
    weight: 0
  - name: my-app-new
    port: 80
    weight: 1
```

See the [Traffic splitting example](/examples/traffic-splitting) from our repository.