# Running on `kind`

This guide walks you through how to run NGINX Kubernetes Gateway on a [kind](https://kind.sigs.k8s.io/) cluster.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kind](https://kind.sigs.k8s.io/)

## Prepare Cluster

Create a cluster with `kind`. You can follow their [instructions](https://kind.sigs.k8s.io/docs/user/quick-start/#creating-a-cluster), or run the following make command at the root of the repository:

```
make create-kind-cluster
```
    
## Deploy NGINX Kubernetes Gateway

Follow the [installation](./installation.md) instructions to deploy NGINX Kubernetes Gateway on your Kind cluster. 

## Access NGINX Kubernetes Gateway

Forward local ports 8080 and 8443 to ports 80 and 443 of the nginx-gateway Pod:

```
kubectl -n nginx-gateway port-forward <pod-name> 8080:80 8443:443
```

> Note: NGINX will not listen on any ports until you configure a [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener. 

## Use NGINX Kubernetes Gateway
To get started, follow the tutorials in the [examples](../examples) directory.
