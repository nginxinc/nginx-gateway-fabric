---
title: "Installation with Manifests"
description: "Learn how to install, upgrade, and uninstall NGINX Gateway Fabric using Kubernetes manifests."
weight: 200
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Prerequisites

To complete this guide, you'll need to install:

- [kubectl](https://kubernetes.io/docs/tasks/tools/), a command-line interface for managing Kubernetes clusters.


## Deploy NGINX Gateway Fabric

Deploying NGINX Gateway Fabric with Kubernetes manifests takes only a few steps. With manifests, you can configure your deployment exactly how you want. Manifests also make it easy to replicate deployments across environments or clusters, ensuring consistency.

{{<note>}}By default, NGINX Gateway Fabric is installed in the **nginx-gateway** namespace. You can deploy in another namespace by modifying the manifest files.{{</note>}}

1. **Install the Gateway API Resources:**
   - Start by installing the Gateway API resources, including the CRDs and the validating webhook:
     ```shell
     kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
     ```

2. **Deploy the NGINX Gateway Fabric CRDs:**
   - Next, deploy the NGINX Gateway Fabric CRDs:
     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
     ```

3. **Deploy NGINX Gateway Fabric:**
   - Then, deploy NGINX Gateway Fabric:
     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
     ```

4. **Verify the Deployment:**
   - To confirm that NGINX Gateway Fabric is running, check the pods in the `nginx-gateway` namespace:
     ```shell
     kubectl get pods -n nginx-gateway
     ```
     The output should look similar to this (note that the pod name will include a unique string):
     ```text
     NAME                             READY   STATUS    RESTARTS   AGE
     nginx-gateway-5d4f4c7db7-xk2kq   2/2     Running   0          112s
     ```


## Upgrade NGINX Gateway Fabric

{{<tip>}}For guidance on zero downtime upgrades, see the [Delay Pod Termination](#configure-delayed-pod-termination-for-zero-downtime-upgrades) section below.{{</tip>}}

To upgrade NGINX Gateway Fabric and get the latest features and improvements, take the following steps:

1. **Upgrade Gateway API Resources:**

   - Verify that your NGINX Gateway Fabric version is compatible with the Gateway API resources. Refer to the [Technical Specifications]({{< relref "reference/technical-specifications.md" >}}) for details.
   - Review the [release notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.8.1) for any important upgrade-specific information.
   - To upgrade the Gateway API resources, run:

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

## Delay Pod Termination for Zero Downtime Upgrades {#configure-delayed-pod-termination-for-zero-downtime-upgrades}

{{< include "installation/delay-pod-termination/delay-pod-termination-overview.md" >}}

Follow these steps to configure delayed pod termination:

1. Open the `nginx-gateway.yaml` for editing.

1. **Add delayed shutdown hooks**:

   - In the `nginx-gateway.yaml` file, add `lifecycle: preStop` hooks to both the `nginx` and `nginx-gateway` container definitions. These hooks instruct the containers to delay their shutdown process, allowing time for connections to close gracefully. Update the `sleep` value to what works for your environment.

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

1. **Set the termination grace period**:

   - {{<include "installation/delay-pod-termination/termination-grace-period.md">}}

1. Save the changes.

{{<see-also>}} 
For additional information on configuring and understanding the behavior of containers and pods during their lifecycle, refer to the following Kubernetes documentation:
- [Container Lifecycle Hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks)
- [Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/Pods/Pod-lifecycle/#Pod-termination)
{{</see-also>}} 


## Uninstall NGINX Gateway Fabric

Follow these steps to uninstall NGINX Gateway Fabric and Gateway API from your Kubernetes cluster:

1. **Uninstall NGINX Gateway Fabric:**

   - To remove NGINX Gateway Fabric and its custom resource definitions (CRDs), run:

     ```shell
     kubectl delete -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
     ```

     ```shell
     kubectl delete -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
     ```

2. **Remove the Gateway API resources:**

   - {{<include "installation/helm/uninstall-gateway-api-resources.md" >}}

## Next Steps

### Expose NGINX Gateway Fabric

{{<include "installation/next-step-expose-fabric.md">}}

