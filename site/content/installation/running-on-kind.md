---
title: "Deploy NGINX Gateway Fabric on a kind Cluster"
weight: 400
toc: true
docs: "DOCS-1428"
---

## Overview

Learn how to run NGINX Gateway Fabric on a kind (Kubernetes in Docker) cluster.

## Before you begin

To complete the steps in this guide, you first need to install the following tools for Kubernetes management and development:

- [kubectl](https://kubernetes.io/docs/tasks/tools/): A command-line interface for Kubernetes that allows you to manage and inspect cluster resources, and control containerized applications.
- [kind](https://kind.sigs.k8s.io/): Short for _Kubernetes in Docker_, this tool lets you run Kubernetes clusters locally using Docker containers, ideal for testing and development purposes.


## Create a kind Cluster

To create a kind cluster, choose from the following options:

- **Option 1**: Use the `kind` tool. For detailed instructions, refer to the kind quick start guide [Creating a Cluster](https://kind.sigs.k8s.io/docs/user/quick-start/#creating-a-cluster).

- **Option 2**: Run the following `make` command in the root of your repository:

   ```makefile
   make create-kind-cluster
   ```

   This command creates a kind cluster using the settings from your Makefile.


## Deploy NGINX Gateway Fabric

Now that you've created a kind cluster, the next step is to install NGINX Gateway Fabric.

To install NGINX Gateway Fabric, choose the appropriate installation guide that suits your setup:

- [Installation with Helm]({{< relref "installation/installing-ngf/helm.md" >}})
- [Installation with Kubernetes manifests]({{< relref "installation/installing-ngf/manifests.md" >}})

{{<note>}}
When using kind clusters, be aware that NodePort services require [additional setup](https://kind.sigs.k8s.io/docs/user/configuration/#nodeport-with-port-mappings). Also, for LoadBalancer services, you’ll need a [third-party controller](https://kind.sigs.k8s.io/docs/user/loadbalancer/) like MetalLB to assign external IPs. The default Helm chart creates a LoadBalancer service; however, you can disable this by adding `--set service.create=false` to your Helm command. Afterward, you can [configure port forwarding](#configure-port-forwarding) as described below to access the examples.
{{</note>}}

## Configure Port Forwarding {#configure-port-forwarding}

Once NGINX Gateway Fabric has been installed, you need to configure port forwarding from local ports **8080** and **8443** to ports **80** and **443** on the **nginx-gateway** Pod.

To configure port forwarding, run the following command:

```shell
kubectl -n nginx-gateway port-forward <pod-name> 8080:80 8443:443
```

{{< note >}}NGINX will only start listening on these ports after you set up a [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener.{{</note>}}

## Get Started with NGINX Gateway Fabric

Learn how to use NGINX Gateway Fabric by exploring the tutorials in the [examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.3.0/examples) directory. The guides provide practical instructions and scenarios to help you use NGINX Gateway Fabric effectively.
