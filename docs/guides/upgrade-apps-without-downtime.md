# Upgrading Applications without Downtime

This guide shows how the specific features of NGINX Kubernetes Gateway can help upgrade your applications without
downtime.

> We assume you are familiar with the upgrade methods mentioned in this guide. As a result, we will not cover
> them in depth, but instead focus on how NGINX Kubernetes Gateway supports them.
<!--- This comment silences the linter. Otherwise, it will complain about the empty line, which is intentional here-->
> By application downtime we mean the clients cannot get responses from your application. Instead,
> they get responses with an error status code like 502 from NGINX.

This guide covers:

- Rolling Deployment upgrade
- Blue-green Deployments
- Canary releases

## Assumptions

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

Before we cover the upgrade methods, we will explain some features of NGINX that help prevent application downtime
regardless of the chosen upgrade method.

## Preventing Downtime During Upgrade

NGINX graceful configuration reloads combined with how NGINX handles changes of Endpoints help prevent
application downtime.

### Graceful Configuration Reloads

If a relevant Gateway API or Kubernetes built-in resource is changed, NGINX Kubernetes Gateway will update NGINX
accordingly by regenerating NGINX configuration and then sending a reload signal to the NGINX master process to apply
the new configuration (this process is further explained in [NGINX documentation][reconfiguration]). We call such an
operation a reload. Because client requests do not get dropped during a reload, it is considered graceful.

[reconfiguration]:https://nginx.org/en/docs/control.html?#reconfiguration


> See also the [Architecture doc](/docs/architecture.md) to learn more about NGINX Kubernetes Gateway architecture.

A subset of all possible configuration changes is changes to Endpoints, which are the most frequent changes during
an application upgrade. How NGINX handles them also prevents downtime.

### Adding or Removing Endpoints

During an upgrade of an application, Kubernetes starts the Pods of the new version and brings down the old ones. It also
deletes and creates the corresponding Endpoints. NGINX Kubernetes Gateway sees the changes to the Endpoints by watching
for changes to the corresponding [EndpointSlices][endpoint-slices].

[endpoint-slices]:https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/

In NGINX configuration, a Service is represented as an [upstream][upstream], and an Endpoint as an
[upstream server][upstream-server].

[upstream]:https://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream

[upstream-server]:https://nginx.org/en/docs/http/ngx_http_upstream_module.html#server

Let's consider two cases:

- If an Endpoint is added, NGINX Kubernetes Gateway adds an upstream server to NGINX that corresponds to the Endpoint,
  and then reloads NGINX. After that, NGINX will start proxying traffic to that Endpoint.
- If an Endpoint is removed, NGINX Kubernetes Gateway removes the corresponding upstream server from NGINX. After
  a reload, NGINX will stop proxying traffic to it. However, it will finish proxying any pending requests to that
  server, which prevents downtime.

As a result, as long you have more than one ready Endpoint, the clients should not experience any downtime during
an upgrade.

> It is a good practice to configure a [Readiness probe][readiness-probe] in the Deployment so that a Pod can advertise
> when it is ready to receive traffic. Note that NGINX Kubernetes Gateway will not add any Endpoint to NGINX that is not
> ready.

[readiness-probe]:https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/

The upgrade methods from the next sections will change upstream servers in NGINX.

## Rolling Deployment Upgrade

To start a [rolling Deployment upgrade][rolling-upgrade], you update the Deployment to use the new version tag of
the application. As a result, Kubernetes terminates the Pods with the old version and create new ones. By default,
Kubernetes also ensures that some number of Pods always stay available during the upgrade.

[rolling-upgrade]:https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment

Such an upgrade will add new upstream servers to NGINX and remove the old ones. As long as the number
of Pods (ready Endpoints) during an upgrade does not reach zero, NGINX will be able to proxy traffic, and thus prevent
any downtime.

To use this method, it is not necessary to update the HTTPRoute.

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

See the [Traffic splitting example](/examples/traffic-splitting) from our repo.
