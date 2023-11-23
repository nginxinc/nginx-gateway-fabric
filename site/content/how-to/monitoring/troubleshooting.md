---
title: "Troubleshooting"
description: "Learn how to troubleshoot common NGINX Gateway Fabric issues."
weight: 200
toc: true
docs: "DOCS-000"
---

## failed to reload NGINX: failed to send the HUP signal to NGINX main: operation not permitted

Depending on your environment's configuration, the control plane may not have the proper permissions to reload NGINX. If NGINX configuration is not applied and you see the above error in the `nginx-gateway` logs, you will need to set `allowPrivilegeEscalation` to `true`. If using Helm, you can set the `nginxGateway.securityContext.allowPrivilegeEscalation` value.
If using the manifests directly, you can update this field under the `nginx-gateway` container's `securityContext`.
