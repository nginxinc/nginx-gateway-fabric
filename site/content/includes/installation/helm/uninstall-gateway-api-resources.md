---
docs:
---

To remove the Gateway resources using the Gateway API repository, use the following command:

   {{<warning>}}This action will delete all corresponding custom resources in your cluster across all namespaces. Ensure no custom resources you want to keep or other Gateway API implementations are running in your cluster before proceeding!{{</warning>}}

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```
