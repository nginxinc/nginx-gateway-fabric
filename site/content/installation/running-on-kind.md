---
title: "Running on kind"
description: "Learn how to run NGINX Gateway Fabric on a kind cluster."
weight: 300
toc: true
docs: "DOCS-000"
---

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kind](https://kind.sigs.k8s.io/)

## Prepare Cluster

Create a cluster with `kind`. You can follow
their [instructions](https://kind.sigs.k8s.io/docs/user/quick-start/#creating-a-cluster), or run the following make
command at the root of the repository:

```makefile
make create-kind-cluster
```

## Deploy NGINX Gateway Fabric

Follow the [installation](./how-to/installation/installation.md) instructions to deploy NGINX Gateway Fabric on your Kind cluster.

> **Note**
>
> For `kind` clusters, NodePort services require [extra configuration](https://kind.sigs.k8s.io/docs/user/configuration/#nodeport-with-port-mappings)
> and LoadBalancer services need [a third-party controller](https://kind.sigs.k8s.io/docs/user/loadbalancer/)
> like MetalLB for external IP assignment.
> However, the Helm chart creates a LoadBalancer service by default. Therefore, the `--wait`
> flag will hang until timeout. To avoid this, you can disable service creation by adding `--set service.create=false`
> to your Helm command and use the port-forwarding command below instead to try out the examples.

## Access NGINX Gateway Fabric

Forward local ports 8080 and 8443 to ports 80 and 443 of the nginx-gateway Pod:

```shell
kubectl -n nginx-gateway port-forward <pod-name> 8080:80 8443:443
```

> Note: NGINX will not listen on any ports until you configure a
> [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener.

## Use NGINX Gateway Fabric

To get started, follow the tutorials in the [examples](../examples) directory.
