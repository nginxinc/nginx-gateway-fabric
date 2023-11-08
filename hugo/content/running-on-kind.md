---
title: "Running on kind"
description: "Learn how to run NGINX Gateway Fabric on a kind cluster."
weight: 900
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

Follow the [installation](./installation.md) instructions to deploy NGINX Gateway Fabric on your Kind cluster.

## Access NGINX Gateway Fabric

Forward local ports 8080 and 8443 to ports 80 and 443 of the nginx-gateway Pod:

```shell
kubectl -n nginx-gateway port-forward <pod-name> 8080:80 8443:443
```

> Note: NGINX will not listen on any ports until you configure a
> [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener.

## Use NGINX Gateway Fabric

To get started, follow the tutorials in the [examples](../examples) directory.
