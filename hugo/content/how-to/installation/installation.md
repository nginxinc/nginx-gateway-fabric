---
title: "Installing NGINX Gateway Fabric"
description: "Learn how to install NGINX Gateway Fabric on a generic Kubernetes cluster."
weight: 100
toc: true
docs: "DOCS-000"
---

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Deploy NGINX Gateway Fabric using Helm

To deploy NGINX Gateway Fabric using Helm, please follow the instructions on [this](/deploy/helm-chart/README.md)
page.

## Deploy NGINX Gateway Fabric from Manifests

> Note: By default, NGINX Gateway Fabric (NGF) will be installed into the nginx-gateway Namespace.
> It is possible to run NGF in a different Namespace, although you'll need to make modifications to the installation
> manifests.

1. Install the Gateway API resources from the standard channel (the CRDs and the validating webhook):

   ```shell
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

1. Deploy the NGINX Gateway Fabric CRDs:

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
   ```

1. Deploy the NGINX Gateway Fabric:

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
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
This Service must live in the same Namespace as the controller. The name of this Service is provided in
the `--service` argument to the controller.

> **Important**
> The Service manifests expose NGINX Gateway Fabric on ports 80 and 443, which exposes any
> Gateway [Listener](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Listener)
> configured for those ports. If you'd like to use different ports in your listeners,
> update the manifests accordingly.
>
> Additionally, NGINX Gateway Fabric will not listen on any ports until you configure a
[Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener.

NGINX Gateway Fabric will use this Service to set the Addresses field in the Gateway Status resource. A LoadBalancer
Service sets the status field to the IP address and/or Hostname. If no Service exists, the Pod IP address is used.

### Create a NodePort Service

Create a Service with type `NodePort`:

```shell
kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/nodeport.yaml
```

A `NodePort` Service will randomly allocate one port on every Node of the cluster. To access NGINX Gateway Fabric,
use an IP address of any Node in the cluster along with the allocated port.

### Create a LoadBalancer Service

Create a Service with type `LoadBalancer` using the appropriate manifest for your cloud provider.

- For GCP or Azure:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/loadbalancer.yaml
   ```

  Lookup the public IP of the load balancer, which is reported in the `EXTERNAL-IP` column in the output of the
  following command:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

  Use the public IP of the load balancer to access NGINX Gateway Fabric.

- For AWS:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/loadbalancer-aws-nlb.yaml
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

> **Note**
> See [below](#configure-delayed-termination-for-zero-downtime-upgrades) for instructions on how to configure delayed
> termination if required for zero downtime upgrades in your environment.

### Upgrade NGINX Gateway Fabric from Manifests

1. Upgrade the Gateway Resources

   Before you upgrade, ensure the Gateway API resources are the correct version as supported by the NGINX Gateway
   Fabric - [see the Technical Specifications](/README.md#technical-specifications).
   The [release notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.8.1) of the new version of the
   Gateway API might include important upgrade-specific notes and instructions. We advise to check the release notes of
   all versions between the one you're using and the new one.

    To upgrade the Gateway resources from [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

    ```shell
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
    ```

1. Upgrade the NGINX Gateway Fabric CRDs

    Run the following command to upgrade the NGINX Gateway Fabric CRDs:

    ```shell
    kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
    ```

1. Upgrade NGINX Gateway Fabric Deployment

    Run the following command to upgrade NGINX Gateway Fabric:

    ```shell
    kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
    ```

### Upgrade NGINX Gateway Fabric using Helm

To upgrade NGINX Gateway Fabric when the deployment method is Helm, please follow the instructions
[here](/deploy/helm-chart/README.md#upgrading-the-chart).

### Configure Delayed Termination for Zero Downtime Upgrades

To achieve zero downtime upgrades (meaning clients will not see any interruption in traffic while a rolling upgrade is
being performed on NGF), you may need to configure delayed termination on the NGF Pod, depending on your environment.

> **Note**
> When proxying Websocket or any long-lived connections, NGINX will not terminate until that connection is closed
> by either the client or the backend. This means that unless all those connections are closed by clients/backends
> before or during an upgrade, NGINX will not terminate, which means Kubernetes will kill NGINX. As a result, the
> clients will see the connections abruptly closed and thus experience downtime.

#### Configure Delayed Termination Using Manifests

Edit the `nginx-gateway.yaml` to include the following:

1. Add `lifecycle` prestop hooks to both the nginx and the nginx-gateway container definitions:

   ```yaml
   <...>
   name: nginx-gateway
   <...>
   lifecycle:
   preStop:
      exec:
      command:
      - /usr/bin/gateway
      - sleep
      - --duration=40s # This flag is optional, the default is 30s
   <...>
   name: nginx
   <...>
   lifecycle:
   preStop:
      exec:
      command:
      - /bin/sleep
      - "40"
   <...>
   ```

2. Ensure the `terminationGracePeriodSeconds` matches or exceeds the `sleep` value from the `preStopHook` (the default
   is 30). This is to ensure Kubernetes does not terminate the Pod before the `preStopHook` is complete.

> **Note**
> More information on container lifecycle hooks can be found
> [here](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks) and a detailed
> description of Pod termination behavior can be found in
> [Termination of Pods](https://kubernetes.io/docs/concepts/workloads/Pods/Pod-lifecycle/#Pod-termination).

#### Configure Delayed Termination Using Helm

To configure delayed termination on the NGF Pod when the deployment method is Helm, please follow the instructions
[here](/deploy/helm-chart/README.md#configure-delayed-termination-for-zero-downtime-upgrades).

## Uninstalling NGINX Gateway Fabric

### Uninstall NGINX Gateway Fabric from Manifests

1. Uninstall the NGINX Gateway Fabric:

   ```shell
   kubectl delete -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
   ```

   ```shell
   kubectl delete -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
   ```

1. Uninstall the Gateway API resources from the standard channel (the CRDs and the validating webhook):

   >**Warning: This command will delete all the corresponding custom resources in your cluster across all namespaces!
   Please ensure there are no custom resources that you want to keep and there are no other Gateway API implementations
   running in the cluster!**

   ```shell
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

### Uninstall NGINX Gateway Fabric using Helm

To uninstall NGINX Gateway Fabric when the deployment method is Helm, please follow the instructions
[here](/deploy/helm-chart/README.md#uninstalling-the-chart).
