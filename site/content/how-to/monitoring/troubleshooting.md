---
title: "Troubleshooting"

weight: 200
toc: true
docs: "DOCS-000"
---

{{< custom-styles >}}

This topic describes possible issues users might encounter when using Instance Manager. When possible, suggested workarounds are provided.

## NGINX fails to reload

#### Description
 
Depending on your environment's configuration, the control plane may not have the proper permissions to reload NGINX. The NGINX configuration will not be applied and you will see the following error in the _nginx-gateway_ logs: `failed to reload NGINX: failed to send the HUP signal to NGINX main: operation not permitted`

#### Resolution
To resolve this issue you will need to set `allowPrivilegeEscalation` to `true`. 

- If using Helm, you can set the `nginxGateway.securityContext.allowPrivilegeEscalation` value.
- If using the manifests directly, you can update this field under the `nginx-gateway` container's `securityContext`.
