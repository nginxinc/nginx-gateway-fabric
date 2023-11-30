---
docs:
---

   {{<warning>}}This will remove all corresponding custom resources in your entire cluster, across all namespaces. Double-check to make sure you don't have any custom resources you need to keep, and confirm that there are no other Gateway API implementations active in your cluster.{{</warning>}}

   To uninstall the Gateway API resources, including the CRDs and the validating webhook, run the following:

   **Stable release**

   If you were running the latest stable release version of NGINX Gateway Fabric:

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

   **Edge version**

   If you were running the edge version of NGINX Gateway Fabric from the **main** branch:

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml
   ```


   If you are running on Kubernetes 1.23 or 1.24, you also need to delete the validating webhook. To do so, run:

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/webhook-install.yaml
   ```
