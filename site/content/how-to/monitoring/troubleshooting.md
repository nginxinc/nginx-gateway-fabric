---
title: "Troubleshooting"
weight: 400
toc: true
docs: "DOCS-1419"
---

{{< custom-styles >}}

This topic describes general troubleshooting techniques when users run into problems using NGINX Gateway Fabric. When possible, suggested workarounds are provided.


### General troubleshooting

When investigating a problem or requesting help, there are important data points that can be collected to help understand what issues may exist.

#### Resource status

To check the status of a resource, use `kubectl describe`. This example checks the status of the `coffee` HTTPRoute, which has an error:

```shell
kubectl describe httproutes.gateway.networking.k8s.io coffee -n nginx-gateway
```

```text
...
Status:
  Parents:
    Conditions:
      Last Transition Time:  2024-05-31T17:20:51Z
      Message:               The route is accepted
      Observed Generation:   4
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-05-31T17:20:51Z
      Message:               spec.rules[0].backendRefs[0].name: Not found: "bad-backend"
      Observed Generation:   4
      Reason:                BackendNotFound
      Status:                False
      Type:                  ResolvedRefs
    Controller Name:         gateway.nginx.org/nginx-gateway-controller
    Parent Ref:
      Group:         gateway.networking.k8s.io
      Kind:          Gateway
      Name:          gateway
      Namespace:     default
      Section Name:  http
```

If a resource has errors relating to its configuration or relationship to other resources, they can likely be read in the status. The `ObservedGeneration` in the status should match the `ObservedGeneration` of the resource. Otherwise, this could mean that the resource hasn't been processed yet or that the status failed to update.

If no `Status` is written on the resource, further debug by checking if the referenced resources exists and belong to NGINX Gateway Fabric.

#### Events

Events created by NGINX Gateway Fabric or other Kubernetes components could indicate system or configuration issues. To see events:

```shell
kubectl get events -n nginx-gateway
```

For example, a warning event when the NginxGateway configuration CRD is deleted:

```text
kubectl -n nginx-gateway get event
LAST SEEN   TYPE      REASON              OBJECT                                           MESSAGE
5s          Warning   ResourceDeleted     nginxgateway/ngf-config                          NginxGateway configuration was deleted; using defaults
```

#### Get shell access to NGINX container

Getting shell access to containers allows developers and operators to view the environment of a running container, see its logs or diagnose any problems. To get shell access to the NGINX container, use `kubectl exec`:

```shell
kubectl exec -it -n nginx-gateway  <ngf-pod-name> -c nginx /bin/sh
```

#### Logs

Logs from the NGINX Gateway Fabric control plane and data plane can contain information that isn't available to status or events. These can include errors in processing or passing traffic. You can see logs for a crashed or killed container by adding the `-p` flag to the above commands.

1. Container Logs

    To see logs for control plane container:

    ```shell
    kubectl -n nginx-gateway logs <ngf-pod-name> -c nginx-gateway
    ```

    To see logs for data plane container:

    ```shell
    kubectl -n nginx-gateway logs <ngf-pod-name> -c nginx
    ```

1. Error Logs

    For _nginx-gateway_ container, you can `grep` for the word `error` or change the log level to `error` by following steps in Modify log levels (Section 4).

    ```shell
    kubectl -n nginx-gateway logs <ngf-pod-name> -c nginx-gateway | grep error
    ```

    For example, an error message when telemetry is not enabled for NGINX Plus installations:

    ```text
    kubectl logs -n nginx-gateway nginx-gateway-nginx-gateway-fabric-77f8746996-j6z6v | grep error
    Defaulted container "nginx-gateway" out of: nginx-gateway, nginx
    {"level":"error","ts":"2024-06-13T18:22:16Z","logger":"usageReporter","msg":"Usage reporting must be enabled when using NGINX Plus; redeploy with usage reporting enabled","error":"usage reporting not enabled","stacktrace":"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static.createUsageWarningJob.func1\n\tgithub.com/nginxinc/nginx-gateway-fabric/internal/mode/static/manager.go:616\nk8s.io/apimachinery/pkg/util/wait.JitterUntilWithContext.func1\n\tk8s.io/apimachinery@v0.30.1/pkg/util/wait/backoff.go:259\nk8s.io/apimachinery/pkg/util/wait.BackoffUntil.func1\n\tk8s.io/apimachinery@v0.30.1/pkg/util/wait/backoff.go:226\nk8s.io/apimachinery/pkg/util/wait.BackoffUntil\n\tk8s.io/apimachinery@v0.30.1/pkg/util/wait/backoff.go:227\nk8s.io/apimachinery/pkg/util/wait.JitterUntil\n\tk8s.io/apimachinery@v0.30.1/pkg/util/wait/backoff.go:204\nk8s.io/apimachinery/pkg/util/wait.JitterUntilWithContext\n\tk8s.io/apimachinery@v0.30.1/pkg/util/wait/backoff.go:259\ngithub.com/nginxinc/nginx-gateway-fabric/internal/framework/runnables.(*CronJob).Start\n\tgithub.com/nginxinc/nginx-gateway-fabric/internal/framework/runnables/cronjob.go:53\nsigs.k8s.io/controller-runtime/pkg/manager.(*runnableGroup).reconcile.func1\n\tsigs.k8s.io/controller-runtime@v0.18.4/pkg/manager/runnable_group.go:226"}
    ```

    For _nginx_ container you can `grep` for various [error](https://nginx.org/en/docs/ngx_core_module.html#error_log) log levels. For example, to search for all logs logged at the `emerg` level:

    ```shell
    kubectl -n nginx-gateway logs <ngf-pod-name> -c nginx | grep emerg
    ```

    For example, if a variable is too long, NGINX may display such an error message:

    ```text
    kubectl logs -n nginx-gateway ngf-nginx-gateway-fabric-bb8598998-jwk2m -c nginx | grep emerg
    2024/06/13 20:04:17 [emerg] 27#27: too long parameter, probably missing terminating """ character in /etc/nginx/conf.d/http.conf:78
    ```

1. Access Logs

    NGINX access logs record all requests processed by the NGINX server. These logs provide detailed information about each request, which can be useful for troubleshooting and analyzing web traffic.
    Access logs can be viewed with the above method of using `kubectl logs`, or by viewing the access log file directly. To do that, get shell access to your NGINX container using these [steps](#get-shell-access-to-nginx-container). The access logs are located in the file `/var/log/nginx/access.log` in the NGINX container.

1. Modify Log Levels

    To see debug logs for control plane in NGINX Gateway Fabric, enable verbose logging by editing the `NginxGateway` configuration. This can be done either before or after deploying NGINX Gateway Fabric. Refer to this [guide](https://docs.nginx.com/nginx-gateway-fabric/how-to/configuration/control-plane-configuration) to do so.

1. Debug Logs

    Once you have modified log levels for NGINX Gateway Fabric, you can `grep` for the word `debug` to check debug logs for further investigation.


#### Understanding the generated NGINX config

Understanding the NGINX configuration is key for fixing issues because it shows how NGINX handles requests. This helps tweak settings to make sure NGINX behaves the way you want it to for your application. To see your current configuration, you can open a shell in the _nginx_ container by following these [steps](#get-shell-access-to-nginx-container) and run `nginx -T`. To understand the usage of NGINX directives in the configuration file, consult this list of [NGINX directives](https://nginx.org/en/docs/dirindex.html).

In this section, we will see how the configuration gets updated as we configure different Services, Deployments and HTTPRoutes with NGINX Gateway Fabric. In the configuration file, you'll often find several server blocks, each assigned to specific ports and server names. NGINX selects the appropriate server for a request and evaluates the URI against the location directives within that block.
When only a Gateway resource is defined, but no Services or HTTPRoutes are configured, NGINX generates a basic configuration. This includes a default server listening on the ports specified in the Gateway listeners, handling all incoming requests. Additionally, there are blocks to manage errors with status codes 500 or 502.

This is a default `server` block listening on port 80:

```text
server {
    listen 80 default_server;

    default_type text/html;
    return 404;
}
```

Once a HTTPRoute with path matches and rules are defined, the nginx.conf is updated accordingly to determine which location block will manage incoming requests. To demonstrate how `nginx.conf` is changed, let's create some resources:

1. A Gateway with single listener with the hostname `*.example.com` on port 80.
2. A simple `coffee` application.
3. An HTTPRoute that exposes the `coffee` application outside the cluster using the listener created in step 1. The path and rule matches create different location blocks in `nginx.conf` to route requests as needed.

For example, this `coffee` route matches requests with path `/coffee` and type `prefix`. Lets see how the `nginx.conf` is modified.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: coffee
spec:
  parentRefs:
  - name: gateway
    sectionName: http
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
```

The modified `nginx.conf`:

```shell
server {
    listen 80 default_server;

    default_type text/html;
    return 404;
}

server {
    listen 80;

    server_name cafe.example.com;


    location /coffee/ {
        proxy_set_header Host "$gw_api_compliant_host";
        proxy_set_header X-Forwarded-For "$proxy_add_x_forwarded_for";
        proxy_set_header Upgrade "$http_upgrade";
        proxy_set_header Connection "$connection_upgrade";
        proxy_http_version 1.1;
        proxy_pass http://default_coffee_80$request_uri;
    }

    location = /coffee {
        proxy_set_header Host "$gw_api_compliant_host";
        proxy_set_header X-Forwarded-For "$proxy_add_x_forwarded_for";
        proxy_set_header Upgrade "$http_upgrade";
        proxy_set_header Connection "$connection_upgrade";
        proxy_http_version 1.1;
        proxy_pass http://default_coffee_80$request_uri;
    }

    location / {
        return 404 "";
    }

}
upstream default_coffee_80 {
    random two least_conn;
    zone default_coffee_80 512k;

    server 10.244.0.13:8080;
}
```

Some key things to note here:

1. A new `server` block is created with the hostname of the HTTPRoute. When a request is sent to this hostname, it will be handled by this `server` block.
2. Within the `server` block, three new `location` blocks are added for *coffee*, each with distinct prefix and exact paths. Requests directed to the *coffee* application with a path prefix `/coffee/hello` will be managed by the first location block, while those with an exact path `/coffee` will be handled by the second location block. Any other requests not recognized by the server block for this hostname will default to the third location block, returning a 404 Not Found status.
3. Each `location` block has headers and directives that configure the NGINX proxy to forward requests to the `/coffee` path correctly, preserving important client information and ensuring compatibility with the upstream server.
4. The `upstream` block in the given NGINX configuration defines a group of backend servers and configures how NGINX should load balance requests among them.

Now let's check the behaviour when curl request is sent to the `coffee` application:

Matches location /coffee/ block

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee/hello
Handling connection for 8080
Server address: 10.244.0.13:8080
Server name: coffee-56b44d4c55-hwpkp
Date: 13/Jun/2024:22:51:52 +0000
URI: /coffee/hello
Request ID: 21fc2baad77337065e7cf2cd57e04383
```

Matches location = /coffee block

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
Handling connection for 8080
Server address: 10.244.0.13:8080
Server name: coffee-56b44d4c55-hwpkp
Date: 13/Jun/2024:22:51:40 +0000
URI: /coffee
Request ID: 4d8d719e95063303e290ad74ecd7339f
```

Matches location / block

```shell
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/
Handling connection for 8080
<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx/1.25.4</center>
</body>
```

{{< warning >}}
The configuration may change in future releases. This configuration is valid for version 1.3.
{{< /warning >}}

#### Metrics for Troubleshooting

Metrics can be useful to identify performance bottlenecks and pinpoint areas of high resource consumption within NGINX Gateway Fabric. To set up metrics collection, refer to this [guide]({{< relref "prometheus.md" >}}). The metrics dashboard will help you understand problems with the way NGINX Gateway Fabric is set up or potential issues that could show up with time.

For example, metrics `nginx_reloads_total` and `nginx_reload_errors_total` offer valuable insights into the system's stability and reliability. A high `nginx_reloads_total` value indicates frequent updates or configuration changes, while a high `nginx_reload_errors_total` value suggests issues with the configuration or other problems preventing successful reloads. Monitoring these metrics helps identify and resolve configuration errors, ensuring consistent service reliability.

In such situations, it's advisable to review the logs of both NGINX and NGINX Gateway containers for any potential error messages. Additionally, verify the configured resources to ensure they are in a valid state.

#### Access the N+ Dashboard

If you have NGINX Gateway Fabric installed with NGINX Plus, you can access the N+ dashboard at `http://localhost:8080/dashboard.html`.
Note that, the port number i.e `8080` matches the port number you have port-forwarded your NGINX Gateway Fabric Pod. For further details, see the [dashboard guide]({{< relref "dashboard.md" >}})

### Common Errors

{{< bootstrap-table "table table-striped table-bordered" >}}
| Problem Area                 | Symptom                                | Troubleshooting Method                                                                                              | Common Cause                                                                                                                            |
|------------------------------|----------------------------------------|---------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| Startup                      | NGINX Gateway Fabric fails to start.   | Check logs for _nginx_ and _nginx-gateway_ container.                                                               | Readiness probe failed.                                                                                                                 |
| Resources not configured     | Status missing on resources.           | Check referenced resources.                                                                                         | Referenced resources do not belong to NGINX Gateway Fabric.                                                                             |
| NGINX errors                 | Reload failures on NGINX               | Fix permissions for control plane.                                                                                  | Security context not configured.                                                                                                        |
| Usage reporting              | Errors logs related to usage reporting | Enable usage reporting. Refer to [Usage Reporting]({{< relref "installation/usage-reporting.md" >}})                | Usage reporting disabled.                                                                                                               |
| Client Settings              | Request entity too large error         | Adjust client settings. Refer to [Client Settings Policy]({{< relref "../traffic-management/client-settings.md" >}})   | Payload is greater than the [`client_max_body_size`](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size) value.|
{{< /bootstrap-table >}}

##### NGINX fails to reload

NGINX reload errors can occur for various reasons, including syntax errors in configuration files, permission issues, and more. To determine if NGINX has failed to reload, check logs for your _nginx-gateway_ and _nginx_ containers.
You will see the following error in the _nginx-gateway_  logs: `failed to reload NGINX:`, followed by the reason for the failure. Similarly, error logs in _nginx_ container start with `emerg`. For example, `2024/06/12 14:25:11 [emerg] 12345#0: open() "/var/run/nginx.pid" failed (13: Permission denied)` shows a critical error, such as a permission problem preventing NGINX from accessing necessary files.

To debug why your reload has failed, start with verifying the syntax of your configuration files by opening a shell in the NGINX container following these [steps](#get-shell-access-to-nginx-container) and running `nginx -T`. If there are errors in your configuration file, the reload will fail and specify the reason for it.

##### NGINX Gateway Fabric Pod is not running or ready

To understand why the NGINX Gateway Fabric Pod has not started running or is not ready, check the state of the Pod to get detailed information about the current status and events happening in the Pod. To do this, use `kubectl describe`:

```shell
kubectl describe pod <ngf-pod-name> -n nginx-gateway
```

The Pod description includes details about the image name, tags, current status, and environment variables. Verify that these details match your setup and cross-check with the events to ensure everything is functioning as expected. For example, the Pod below has two containers that are running and the events reflect the same.

```text
Containers:
  nginx-gateway:
    Container ID:  containerd://06c97a9de938b35049b7c63e251418395aef65dd1ff996119362212708b79cab
    Image:         nginx-gateway-fabric
    Image ID:      docker.io/library/import-2024-06-13@sha256:1460d63bd8a352a6e455884d7ebf51ce9c92c512cb43b13e44a1c3e3e6a08918
    Ports:         9113/TCP, 8081/TCP
    Host Ports:    0/TCP, 0/TCP
    State:          Running
      Started:      Thu, 13 Jun 2024 11:47:46 -0600
    Ready:          True
    Restart Count:  0
    Readiness:      http-get http://:health/readyz delay=3s timeout=1s period=1s #success=1 #failure=3
    Environment:
      POD_IP:          (v1:status.podIP)
      POD_NAMESPACE:  nginx-gateway (v1:metadata.namespace)
      POD_NAME:       ngf-nginx-gateway-fabric-66dd665756-zh7d7 (v1:metadata.name)
  nginx:
    Container ID:   containerd://c2f3684fd8922e4fac7d5707ab4eb5f49b1f76a48893852c9a812cd6dbaa2f55
    Image:          nginx-gateway-fabric/nginx
    Image ID:       docker.io/library/import-2024-06-13@sha256:c9a02cb5665c6218373f8f65fc2c730f018d0ca652ae827cc913a7c6e9db6f45
    Ports:          80/TCP, 443/TCP
    Host Ports:     0/TCP, 0/TCP
    State:          Running
      Started:      Thu, 13 Jun 2024 11:47:46 -0600
    Ready:          True
    Restart Count:  0
    Environment:    <none>
Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  40s   default-scheduler  Successfully assigned nginx-gateway/ngf-nginx-gateway-fabric-66dd665756-zh7d7 to kind-control-plane
  Normal  Pulled     40s   kubelet            Container image "nginx-gateway-fabric" already present on machine
  Normal  Created    40s   kubelet            Created container nginx-gateway
  Normal  Started    39s   kubelet            Started container nginx-gateway
  Normal  Pulled     39s   kubelet            Container image "nginx-gateway-fabric/nginx" already present on machine
  Normal  Created    39s   kubelet            Created container nginx
  Normal  Started    39s   kubelet            Started container nginx
```

##### Insufficient Privileges errors

Depending on your environment's configuration, the control plane may not have the proper permissions to reload NGINX. The NGINX configuration will not be applied and you will see the following error in the _nginx-gateway_ logs:

`failed to reload NGINX: failed to send the HUP signal to NGINX main: operation not permitted`

To **resolve** this issue you will need to set `allowPrivilegeEscalation` to `true`.

- If using Helm, you can set the `nginxGateway.securityContext.allowPrivilegeEscalation` value.
- If using the manifests directly, you can update this field under the `nginx-gateway` container's `securityContext`.

##### Usage Reporting errors

If using NGINX Gateway Fabric with NGINX Plus as the data plane, you will see the following error in the _nginx-gateway_ logs if you have not enabled Usage Reporting:

`usage reporting not enabled`

To **resolve** this issue, enable Usage Reporting by following the [Usage Reporting]({{< relref "installation/usage-reporting.md" >}}) guide.

##### 413 Request Entity Too Large

If you receive the following error:

```text
<html>
<head><title>413 Request Entity Too Large</title></head>
<body>
<center><h1>413 Request Entity Too Large</h1></center>
<hr><center>nginx/1.25.5</center>
</body>
</html>
```

Or view the following error message in the NGINX logs:

```text
2024/05/30 21:48:22 [error] 138#138: *43 client intended to send too large body: 112 bytes, client: 127.0.0.1, server: cafe.example.com, request: "POST /coffee HTTP/1.1", host: "cafe.example.com:8080"
```

The request body exceeds the [client_max_body_size](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size).
To **resolve** this, you can configure the `client_max_body_size` using the `ClientSettingsPolicy` API. Read the [Client Settings Policy]({{< relref "how-to/traffic-management/client-settings.md" >}}) documentation for more information.


### Further Reading

You can check out the [Kubernetes Troubleshooting Guide](https://kubernetes.io/docs/tasks/debug/debug-application/) for further assistance
