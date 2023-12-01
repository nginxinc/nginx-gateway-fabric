---
title: "Deploying NGINX Gateway Fabric on a kind Cluster"
description: "Learn how to run NGINX Gateway Fabric on a kind (Kubernetes in Docker) cluster."
weight: 300
toc: true
docs: "DOCS-000"
---

{{< custom-styles >}}

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kind](https://kind.sigs.k8s.io/)

## Prepare Cluster

To create a cluster, use the `kind` tool. For detailed instructions, refer to the kind quick start guide [Creating a Cluster](https://kind.sigs.k8s.io/docs/user/quick-start/#creating-a-cluster).

Alternatively, run the following `make` command in the root of your repository:

```makefile
make create-kind-cluster
```

This command creates the cluster using settings from your makefile.


## Deploy NGINX Gateway Fabric

Follow the instructions to deploy NGINX Gateway Fabric on your kind cluster:

- [Installation with Helm]({{< relref "installation/installing-ngf/helm.md" >}})
- [Installation with Kubernetes manifests]({{< relref "installation/installing-ngf/manifests.md" >}})


## Access NGINX Gateway Fabric

Forward your local ports **8080** and **8443** to ports **80** and **443** on the **nginx-gateway** Pod:

```shell
kubectl -n nginx-gateway port-forward <pod-name> 8080:80 8443:443
```

{{< note >}}NGINX will only start listening on these ports after you set up a [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener.{{</note>}}

## Use NGINX Gateway Fabric

See the tutorials in the [examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples) directory to get started.
