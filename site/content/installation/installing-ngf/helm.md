---
title: "Installation with Helm NGINX"
description: "Learn how to install NGINX Gateway Fabric on a generic Kubernetes cluster."
weight: 100
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Prerequisites

- Install [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Deploying NGINX Gateway Fabric 

{{<tabs name="deploy-ngf">}}

{{%tab name="Helm"%}}

### Deploying NGINX Gateway Fabric using Helm


Follow the instructions [to deploy NGINX Gateway Fabric using Helm](/deploy/helm-chart/README.md).


{{%/tab%}}

{{%tab name="Manifest"%}}

### Deploy NGINX Gateway Fabric from Manifests

{{<note>}}NGINX Gateway Fabric installs into the **nginx-gateway** namespace by default. To run NGINX Gateway Fabric in a different namespace, modify the installation manifests.{{</note>}}

1. Install the Gateway API resources (Custom Resource Definitions [CRDs] and validating webhook) from the standard channel:

   ```shell
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

2. Deploy the NGINX Gateway Fabric CRDs:

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
   ```

3. Deploy NGINX Gateway Fabric:

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
   ```

4. Verify NGINX Gateway Fabric is running in the `nginx-gateway` namespace:

   ```shell
   kubectl get pods -n nginx-gateway
   ```

   Expected output (note that `5d4f4c7db7-xk2kq` is a randomly generated string and will vary):

   ```text
   NAME                             READY   STATUS    RESTARTS   AGE
   nginx-gateway-5d4f4c7db7-xk2kq   2/2     Running   0          112s
   ```



### Upgrade NGINX Gateway Fabric from Manifests

Upgrading NGINX Gateway Fabric from manifests involves several steps to ensure all components are updated to their latest versions.

1. **Upgrade Gateway Resources:**
   - Check that the **Gateway API** resources are compatible with your version of NGINX Gateway Fabric ([Technical Specifications](/README.md#technical-specifications)).
   - Review the [release notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.8.1) for any important upgrade-specific information.
   - To upgrade the gateway resources, run:

     ```shell
     kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
     ```

2. **Upgrade NGINX Gateway Fabric CRDs:**
   - To upgrade the Custom Resource Definitions (CRDs), run:

     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
     ```

3. **Upgrade NGINX Gateway Fabric Deployment:**
   - To upgrade the deployment, run:

     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
     ```

{{%/tab%}}

{{</tabs>}}







## Expose NGINX Gateway Fabric

Gain access to NGINX Gateway Fabric by creating either a **NodePort** service or a **LoadBalancer** service in the same namespace as the controller. The service name is specified in the `--service` argument of the controller.

{{<important>}}
The service manifests configure NGINX Gateway Fabric on ports `80` and `443`, affecting any Gateway [Listeners](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Listener) on these ports. To use different ports, update the manifests. NGINX Gateway Fabric requires a configured [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener to listen on any ports.
{{</important>}}

NGINX Gateway Fabric uses this service to update the **Addresses** field in the **Gateway Status** resource. A **LoadBalancer** service sets this field to the IP address and/or hostname. Without a service, the Pod IP address is used.

This gateway is associated with the NGINX Gateway Fabric through the **gatewayClassName** field. The default installation of NGINX Gateway Fabric creates a **GatewayClass** with the name **nginx**. NGINX Gateway Fabric will only configure gateways with a **gatewayClassName** of **nginx** unless you change the name via the `--gatewayclass` [command-line flag](/docs/cli-help.md#static-mode).

### Create a NodePort service

To create a **NodePort** service:

```shell
kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/nodeport.yaml
```

A **NodePort** service allocates a port on every cluster node. Access NGINX Gateway Fabric using any node's IP address and the allocated port.

### Create a LoadBalancer Service

To create a **LoadBalancer** service, use the appropriate manifest for your cloud provider:

- For GCP (Google Cloud Platform) or Azure:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/loadbalancer.yaml
   ```

  Lookup the public IP of the load balancer, which is reported in the `EXTERNAL-IP` column in the output of the following command:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

  Use the public IP of the load balancer to access NGINX Gateway Fabric.

- For AWS (Amazon Web Services):

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/loadbalancer-aws-nlb.yaml
   ```

  In AWS, the NLB (Network Load Balancer) DNS (directory name system) name will be reported by Kubernetes instead of a public IP in the `EXTERNAL-IP` column. To get the DNS name, run:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

  Generally, use the NLB DNS name, but for testing purposes, you can resolve the DNS name to get the IP address of the load balancer:

   ```shell
   nslookup <dns-name>
   ```


---

## Upgrading NGINX Gateway Fabric

This section provides guidelines for upgrading your NGINX Gateway Fabric deployment to ensure you are using the latest features and improvements.

{{<tip>}}For zero-downtime upgrades, follow the instructions to [configure a delayed pod termination](#configure-delayed-pod-termination-for-zero-downtime-upgrades) for the NGINX Gateway Fabric pod.{{</tip>}}



### Upgrade NGINX Gateway Fabric Using Helm

For Helm-managed deployments, follow the [Helm upgrade instructions](/deploy/helm-chart/README.md#upgrading-the-chart).

### Configure Delayed Pod Termination {#configure-delayed-pod-termination-for-zero-downtime-upgrades}

Configuring delayed pod termination is crucial for ensuring zero downtime during upgrades, particularly in environments handling persistent or long-lived connections. The configuration settings required are the same for both Helm- and Manifest-based deployments, although the specific file to update will depend on your deployment type.

{{<note>}}NGINX won't shut down until all websocket or long-lived connections are closed. Keeping these connections open during an upgrade can lead to Kubernetes forcibly shutting down NGINX, potentially causing downtime for clients.{{</note>}}

#### For Helm-Based Deployments

To configure delayed pod termination, follow these steps:

1. Depending on your deployment type, update the `values.yaml` file for Helm-based deployments or the `nginx-gateway.yaml` file for Manifest-based deployments.

1. Add `lifecycle: preStop` hooks to both `nginx` and `nginx-gateway` container definitions. These hooks delay the shutdown process to allow time for connections to close gracefully.

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

1. Ensure `terminationGracePeriodSeconds` is equal to or greater than the `sleep` duration in the preStop hook (default is `30`). This is to ensure Kubernetes does not terminate the pod before the `preStop` Hook is complete. To do so, update your `values.yaml` or `nginx-gateway.yaml` file to include the following (update the value to what is required in your environment):

   ```yaml
   terminationGracePeriodSeconds: 50
   ```

#### Using Helm to Configure Delayed Pod Termination

For Helm-based deployments, follow the [Helm-specific instructions for extending the termination period](/deploy/helm-chart/README.md#configure-delayed-termination-for-zero-downtime-upgrades).

---

Feel free to copy this revised section. Let me know if you need any further modifications or if we should proceed to the next section!

%%%%

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

You can gain access to NGINX Gateway Fabric by creating a **NodePort** Service or a **LoadBalancer** Service.
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

Create a Service with type **NodePort**:

```shell
kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/nodeport.yaml
```

A **NodePort** Service will randomly allocate one port on every Node of the cluster. To access NGINX Gateway Fabric,
use an IP address of any Node in the cluster along with the allocated port.

### Create a LoadBalancer Service

Create a Service with type **LoadBalancer** using the appropriate manifest for your cloud provider.

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
