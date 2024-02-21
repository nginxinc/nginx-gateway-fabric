---
title: "Installation with Kubernetes manifests"
description: "Learn how to install, upgrade, and uninstall NGINX Gateway Fabric using Kubernetes manifests."
weight: 200
toc: true
docs: "DOCS-1429"
---

{{<custom-styles>}}

## Prerequisites

To complete this guide, you'll need to install:

- [kubectl](https://kubernetes.io/docs/tasks/tools/), a command-line interface for managing Kubernetes clusters.


## Deploy NGINX Gateway Fabric

Deploying NGINX Gateway Fabric with Kubernetes manifests takes only a few steps. With manifests, you can configure your deployment exactly how you want. Manifests also make it easy to replicate deployments across environments or clusters, ensuring consistency.

- If you’d like to use NGINX Plus:
  1. To pull from the F5 Container registry, configure a docker registry secret using your JWT token from the MyF5 portal by following the instructions from [here](https://docs.nginx.com/nginx-gateway-fabric/installation/ngf-images/jwt-token-docker-secret). Make sure to specify the secret in the `imagePullSecrets` field of the `nginx-gateway` ServiceAccount.
  1. Alternatively, pull an NGINX Gateway Fabric image with NGINX Plus and push it to your private registry by following the instructions from [here]({{<relref "installation/ngf-images/pulling-ngf-image.md">}}).
  1. Update the nginx container's `image` field of the `nginx-gateway` Deployment accordingly.

### 1. Install the Gateway API resources

{{<include "installation/install-gateway-api-resources.md" >}}

### 2. Deploy the NGINX Gateway Fabric CRDs

#### Stable release

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
   ```

#### Edge version

   ```shell
   git clone https://github.com/nginxinc/nginx-gateway-fabric.git
   cd nginx-gateway-fabric
   ```

   ```shell
   kubectl apply -f deploy/manifests/crds
   ```

### 3. Deploy NGINX Gateway Fabric

   {{<note>}}By default, NGINX Gateway Fabric is installed in the **nginx-gateway** namespace. You can deploy in another namespace by modifying the manifest files.{{</note>}}

#### Stable release

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
   ```

#### Edge version

- For NGINX:

   ```shell
   kubectl apply -f deploy/manifests/nginx-gateway.yaml
   ```

- For NGINX Plus

   ```shell
   kubectl apply -f deploy/manifests/nginx-plus-gateway.yaml
   ```

   Update the nginx-plus-gateway.yaml file to include your chosen image from the F5 Container registry or your custom container image.

#### Enable experimental features

We support a subset of the additional features provided by the Gateway API experimental channel. To enable the experimental features of Gateway API which are supported by NGINX Gateway Fabric:

- For NGINX:

   ```shell
   kubectl apply -f deploy/manifests/nginx-gateway-experimental.yaml
   ```

- For NGINX Plus

   ```shell
   kubectl apply -f deploy/manifests/nginx-plus-gateway-experimental.yaml
   ```

   Update the nginx-plus-gateway-experimental.yaml file to include your chosen image from the F5 Container registry or your custom container image.

{{<note>}}Requires the Gateway APIs installed from the experimental channel.{{</note>}}

### 4. Verify the Deployment

To confirm that NGINX Gateway Fabric is running, check the pods in the `nginx-gateway` namespace:

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

1. **Upgrade Gateway API resources:**

   - Verify that your NGINX Gateway Fabric version is compatible with the Gateway API resources. Refer to the [Technical Specifications]({{< relref "reference/technical-specifications.md" >}}) for details.
   - Review the [release notes](https://github.com/kubernetes-sigs/gateway-api/releases) for any important upgrade-specific information.
   - To upgrade the Gateway API resources, run:

      ```shell
      kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml
      ```

      or, if you installed the from the experimental channel:

      ```shell
      kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/experimental-install.yaml
      ```

   - If you are running on Kubernetes 1.23 or 1.24, you also need to update the validating webhook:

      ```shell
      kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/webhook-install.yaml
      ```

   - If you are running on Kubernetes 1.25 or newer and have the validating webhook installed, you should remove the
   webhook:

      ```shell
      kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/webhook-install.yaml
      ```

1. **Upgrade NGINX Gateway Fabric CRDs:**
   - To upgrade the Custom Resource Definitions (CRDs), run:

      ```shell
      kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
      ```

1. **Upgrade NGINX Gateway Fabric deployment:**
   - To upgrade the deployment, run:

      ```shell
      kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
      ```


## Delay pod termination for zero downtime upgrades {#configure-delayed-pod-termination-for-zero-downtime-upgrades}

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

1. **Remove the Gateway API resources:**

   - {{<include "installation/uninstall-gateway-api-resources.md" >}}

## Next steps

### Expose NGINX Gateway Fabric

{{<include "installation/next-step-expose-fabric.md">}}
