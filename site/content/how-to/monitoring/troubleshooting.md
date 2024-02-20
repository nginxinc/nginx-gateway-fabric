---
title: "Troubleshooting"

weight: 300
toc: true
docs: "DOCS-000"
---

{{< custom-styles >}}

This topic describes possible issues users might encounter when using NGINX Gateway Fabric. When possible, suggested workarounds are provided.

### NGINX fails to reload

#### Description

Depending on your environment's configuration, the control plane may not have the proper permissions to reload NGINX. The NGINX configuration will not be applied and you will see the following error in the _nginx-gateway_ logs:

`failed to reload NGINX: failed to send the HUP signal to NGINX main: operation not permitted`

#### Resolution

To resolve this issue you will need to set `allowPrivilegeEscalation` to `true`.

- If using Helm, you can set the `nginxGateway.securityContext.allowPrivilegeEscalation` value.
- If using the manifests directly, you can update this field under the `nginx-gateway` container's `securityContext`.

### Usage Reporting errors

#### Description

If using NGINX Gateway Fabric with NGINX Plus as the data plane, you will see the following error in the _nginx-gateway_ logs if you have not enabled Usage Reporting:

`usage reporting not enabled`

#### Resolution

To resolve this issue, enable Usage Reporting by following the [Usage Reporting]({{< relref "installation/usage-reporting.md" >}}) guide.
