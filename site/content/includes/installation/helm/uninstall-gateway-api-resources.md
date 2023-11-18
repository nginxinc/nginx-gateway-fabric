---
docs:
---

To uninstall the Gateway API resources, including the CRDs and the validating webhook, run:

   {{<warning>}}This will remove all corresponding custom resources in your entire cluster, across all namespaces. Double-check to make sure you don't have any custom resources you need to keep, and confirm that there are no other Gateway API implementations active in your cluster.{{</warning>}}

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```
