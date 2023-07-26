# Uninstalling NGINX Kubernetes Gateway

This guide walks you through how to uninstall NGINX Kubernetes Gateway.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Uninstall NGINX Kubernetes Gateway using Helm

To uninstall NGINX Kubernetes Gateway when the deployment method is Helm, please follow the instructions
[here](/deploy/helm-chart/README.md#uninstalling-the-chart).

## Uninstall NGINX Kubernetes Gateway from Manifests

1. Clone the repo and change into the `nginx-kubernetes-gateway` directory:

   ```shell
   git clone https://github.com/nginxinc/nginx-kubernetes-gateway.git
   cd nginx-kubernetes-gateway
   ```

1. Uninstall the NGINX Kubernetes Gateway:

   ```shell
   kubectl delete -f deploy/manifests/nginx-gateway.yaml
   ```

1. Uninstall the Gateway API resources from the standard channel (the CRDs and the validating webhook):

   >**Warning: This command will delete all the corresponding custom resources in your cluster across all namespaces!
   Please ensure there are no custom resources that you want to keep and there are no other Gateway API implementations
   running in the cluster!**

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml
   ```
