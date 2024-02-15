---
docs:
---

   {{<warning>}}This will remove all corresponding custom resources in your entire cluster, across all namespaces. Double-check to make sure you don't have any custom resources you need to keep, and confirm that there are no other Gateway API implementations active in your cluster.{{</warning>}}

   To uninstall the Gateway API resources, run the following:

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml
   ```

   Alternatively, if you installed the Gateway APIs from the experimental channel, run the following:

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/experimental-install.yaml
   ```

   If you are running on Kubernetes 1.23 or 1.24, you also need to delete the validating webhook. To do so, run:

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/webhook-install.yaml
   ```
