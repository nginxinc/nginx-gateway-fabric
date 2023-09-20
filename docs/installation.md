# Installation

This guide walks you through how to install NGINX Gateway Fabric on a generic Kubernetes cluster.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Deploy NGINX Gateway Fabric using Helm

To deploy NGINX Gateway Fabric using Helm, please follow the instructions on [this](/deploy/helm-chart/README.md)
page.

## Deploy NGINX Gateway Fabric from Manifests

> Note: By default, NGINX Gateway Fabric (NGF) will be installed into the nginx-gateway Namespace.
> It is possible to run NGF in a different Namespace, although you'll need to make modifications to the installation
> manifests.

1. Clone the repo and change into the `nginx-gateway-fabric` directory:

   ```shell
   git clone https://github.com/nginxinc/nginx-gateway-fabric.git
   cd nginx-gateway-fabric
   ```

1. Check out the latest tag (unless you are installing the `edge` version from the `main` branch):

   ```shell
   git fetch --tags
   latestTag=$(git describe --tags `git rev-list --tags --max-count=1`)
   git checkout $latestTag
   ```

1. Install the Gateway API resources from the standard channel (the CRDs and the validating webhook):

   ```shell
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.0/standard-install.yaml
   ```

1. Deploy the NGINX Gateway Fabric CRDs:

   ```shell
   kubectl apply -f deploy/manifests/crds
   ```

1. Deploy the NGINX Gateway Fabric:

   ```shell
   kubectl apply -f deploy/manifests/nginx-gateway.yaml
   ```

1. Confirm the NGINX Gateway Fabric is running in `nginx-gateway` namespace:

   ```shell
   kubectl get pods -n nginx-gateway
   ```

   ```text
   NAME                             READY   STATUS    RESTARTS   AGE
   nginx-gateway-5d4f4c7db7-xk2kq   2/2     Running   0          112s
   ```

## Expose NGINX Gateway Fabric

You can gain access to NGINX Gateway Fabric by creating a `NodePort` Service or a `LoadBalancer` Service.

> Important
>
> The Service manifests expose NGINX Gateway Fabric on ports 80 and 443, which exposes any
> Gateway [Listener](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Listener)
> configured for those ports. If you'd like to use different ports in your listeners,
> update the manifests accordingly.

### Create a NodePort Service

Create a Service with type `NodePort`:

```shell
kubectl apply -f deploy/manifests/service/nodeport.yaml
```

A `NodePort` Service will randomly allocate one port on every Node of the cluster. To access NGINX Gateway Fabric,
use an IP address of any Node in the cluster along with the allocated port.

### Create a LoadBalancer Service

Create a Service with type `LoadBalancer` using the appropriate manifest for your cloud provider.

- For GCP or Azure:

   ```shell
   kubectl apply -f deploy/manifests/service/loadbalancer.yaml
   ```

  Lookup the public IP of the load balancer, which is reported in the `EXTERNAL-IP` column in the output of the
  following command:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

  Use the public IP of the load balancer to access NGINX Gateway Fabric.

- For AWS:

   ```shell
   kubectl apply -f deploy/manifests/service/loadbalancer-aws-nlb.yaml
   ```

  In AWS, the NLB DNS name will be reported by Kubernetes in lieu of a public IP in the `EXTERNAL-IP` column. To get the
  DNS name run:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

  In general, you should rely on the NLB DNS name, however for testing purposes you can resolve the DNS name to get the
  IP address of the load balancer:

   ```shell
   nslookup <dns-name>
   ```

## Upgrading NGINX Gateway Fabric

### Upgrade NGINX Gateway Fabric from Manifests

1. Upgrade the Gateway Resources

    Before you upgrade, ensure the Gateway API resources are the correct version as supported by the
    NGINX Gateway Fabric - [see the Technical Specifications](/README.md#technical-specifications).:

    To upgrade the Gateway resources from [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

    ```shell
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.0/standard-install.yaml
    ```

1. Upgrade the NGINX Gateway Fabric CRDs

    Run the following command to upgrade the NGINX Gateway Fabric CRDs:

    ```shell
    kubectl apply -f deploy/manifests/crds
    ```

    The following warning is expected and can be ignored:

    ```text
    Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply.
    ```

1. Upgrade NGINX Gateway Fabric Deployment

    Run the following command to upgrade NGINX Gateway Fabric:

    ```shell
    kubectl apply -f deploy/manifests/nginx-gateway.yaml
    ```

### Upgrade NGINX Gateway Fabric using Helm

To upgrade NGINX Gateway Fabric when the deployment method is Helm, please follow the instructions
[here](/deploy/helm-chart/README.md#upgrading-the-chart).

## Uninstalling NGINX Gateway Fabric

### Uninstall NGINX Gateway Fabric from Manifests

1. Uninstall the NGINX Gateway Fabric:

   ```shell
   kubectl delete -f deploy/manifests/nginx-gateway.yaml
   ```

   ```shell
   kubectl delete -f deploy/manifests/crds
   ```

1. Uninstall the Gateway API resources from the standard channel (the CRDs and the validating webhook):

   >**Warning: This command will delete all the corresponding custom resources in your cluster across all namespaces!
   Please ensure there are no custom resources that you want to keep and there are no other Gateway API implementations
   running in the cluster!**

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.0/standard-install.yaml
   ```

### Uninstall NGINX Gateway Fabric using Helm

To uninstall NGINX Gateway Fabric when the deployment method is Helm, please follow the instructions
[here](/deploy/helm-chart/README.md#uninstalling-the-chart).
