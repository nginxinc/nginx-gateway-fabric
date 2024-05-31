---
title: "Troubleshooting"
weight: 400
toc: true
docs: "DOCS-1419"
---

{{< custom-styles >}}

This topic describes possible issues users might encounter when using NGINX Gateway Fabric. When possible, suggested workarounds are provided.

### **General Troubleshooting**

When attempting to diagnose a problem or get support, there are a few important data points that can be collected to help with understanding what issues may exist.

#### Resource Status

To get the status of a resource, use `kubectl describe`. For example, to check the status of the `coffee` HTTPRoute:

```shell
kubectl describe httproutes.gateway.networking.k8s.io coffee [-n namespace]
```

```text
...
Status:
  Parents:
    Conditions:
      Last Transition Time:  2024-05-31T16:22:26Z
      Message:               The route is accepted
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-05-31T16:22:26Z
      Message:               All references are resolved
      Observed Generation:   1
      Reason:                ResolvedRefs
      Status:                True
      Type:                  ResolvedRefs
    Controller Name:         gateway.nginx.org/nginx-gateway-controller
    Parent Ref:
      Group:         gateway.networking.k8s.io
      Kind:          Gateway
      Name:          gateway
      Namespace:     default
      Section Name:  http
```

If a resource has any errors relating to its configuration or relation to other resources, it is likely that those errors will be contained within the status.

#### Events

Events may be created by NGINX Gateway Fabric or other Kubernetes components that could indicate system or configuration issues. To see events:

```shell
kubectl get events [-n namespace]
```

#### Logs

Logs of the NGINX Gateway Fabric control plane and data plane can contain information that isn't otherwise reported in status or events. These could include errors in processing or passing traffic.

To see logs for the control plane container (replacing the name of the deployment if necessary):

```shell
kubectl -n nginx-gateway logs deployments/nginx-gateway-fabric -c nginx-gateway
```

To see logs for the data plane container (replacing the name of the deployment if necessary):

```shell
kubectl -n nginx-gateway logs deployments/nginx-gateway-fabric -c nginx
```

You can also see the logs of a container that has crashed or been killed, by specifying the `-p` flag with the above commands.

### **NGINX fails to reload**

#### Description

Depending on your environment's configuration, the control plane may not have the proper permissions to reload NGINX. The NGINX configuration will not be applied and you will see the following error in the _nginx-gateway_ logs:

`failed to reload NGINX: failed to send the HUP signal to NGINX main: operation not permitted`

#### Resolution

To resolve this issue you will need to set `allowPrivilegeEscalation` to `true`.

- If using Helm, you can set the `nginxGateway.securityContext.allowPrivilegeEscalation` value.
- If using the manifests directly, you can update this field under the `nginx-gateway` container's `securityContext`.

### **Usage Reporting errors**

#### Description

If using NGINX Gateway Fabric with NGINX Plus as the data plane, you will see the following error in the _nginx-gateway_ logs if you have not enabled Usage Reporting:

`usage reporting not enabled`

#### Resolution

To resolve this issue, enable Usage Reporting by following the [Usage Reporting]({{< relref "installation/usage-reporting.md" >}}) guide.
