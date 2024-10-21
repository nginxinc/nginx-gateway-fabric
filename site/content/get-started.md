---
title: Get started
toc: true
weight: 200
docs: DOCS-000
---

This is a guide for getting started with NGINX Gateway Fabric. It explains how to:

- Set up a [kind (Kubernetes in Docker)](https://kind.sigs.k8s.io/) cluster
- Install [NGINX Gateway Fabric](https://blog.nginx.org/blog/5-things-to-know-about-nginx-gateway-fabric) with [NGINX](https://nginx.org/)
- Test NGINX Gateway Fabric with an example application

By following the steps in order, you will finish with a functional NGINX Gateway Fabric cluster.

---

## Before you begin

To complete this guide, you need the following prerequisites installed:

- [Go 1.16](https://go.dev/dl/) or newer, which is used by kind
- [Docker](https://docs.docker.com/get-started/get-docker/), for creating and managing containers
- [kind](https://kind.sigs.k8s.io/#installation-and-usage), which allows for running a local Kubernetes cluster using Docker
- [kubectl](https://kubernetes.io/docs/tasks/tools/), which provides a command line interface (CLI) for interacting with Kubernetes clusters
- [Helm 3.0](https://helm.sh/docs/intro/install/) or newer to install NGINX Gateway Fabric
- [curl](https://curl.se/), to test the example application

## Set up a kind cluster

Create the file _cluster-config.yaml_ with the following contents, noting the highlighted lines:

```yaml {linenos=true, hl_lines=[6, 9]}
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 31437
    hostPort: 8080
    protocol: TCP
  - containerPort: 31438
    hostPort: 8443
    protocol: TCP
```

{{< warning >}}
Take note of the two _containerPort_ values. They are necessary for later configuring a _NodePort_.
{{< /warning >}}

Run the following command:

```shell
kind create cluster --config cluster-config.yaml
```
```text
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.31.0) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂
```

{{< note >}}
If you have cloned [the NGINX Gateway Fabric repository](https://github.com/nginxinc/nginx-gateway-fabric/tree/main), you can also create a kind cluster from the root folder with the following *make* command:

```shell
make create-kind-cluster
```
{{< /note >}}

---

## Install NGINX Gateway Fabric

### Add Gateway API resources

Use `kubectl` to add the API resources for NGINX Gateway Fabric with the following command:

```shell
kubectl kustomize "https://github.com/nginxinc/nginx-gateway-fabric/config/crd/gateway-api/standard?ref=v1.4.0" | kubectl apply -f -
```
```text
customresourcedefinition.apiextensions.k8s.io/gatewayclasses.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/gateways.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/grpcroutes.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/httproutes.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/referencegrants.gateway.networking.k8s.io created
```

{{< note >}}
To use experimental features, you'll need to install the API resources from the experimental channel instead.

```shell
kubectl kustomize "https://github.com/nginxinc/nginx-gateway-fabric/config/crd/gateway-api/experimental?ref=v1.4.0" | kubectl apply -f -
```
{{< /note >}}

---

### Install the Helm chart

Use `helm` to install NGINX Gateway Fabric with the following command:

```shell
helm install ngf oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric --create-namespace -n nginx-gateway --set service.create=false
```
```text
Pulled: ghcr.io/nginxinc/charts/nginx-gateway-fabric:1.4.0
Digest: sha256:9bbd1a2fcbfd5407ad6be39f796f582e6263512f1f3a8969b427d39063cc6fee
NAME: ngf
LAST DEPLOYED: Fri Oct 11 16:57:20 2024
NAMESPACE: nginx-gateway
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

{{< note >}}
If you installed the API resources from the experimental channel during the last step, you will need to enable the _nginxGateway.gwAPIExperimentalFeatures_ option:

```shell
helm install ngf oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric --create-namespace -n nginx-gateway --set service.create=false --set nginxGateway.gwAPIExperimentalFeatures.enable=true
```
{{< /note >}}

---

### Set up a NodePort

Create the file _nodeport-config.yaml_ with the following contents:

{{< note >}}
The highlighted _nodePort_ values should equal the _containerPort_ values from _cluster-config.yaml_ [when you created the kind cluster](#set-up-a-kind-cluster).
{{< /note >}}

```yaml {linenos=true, hl_lines=[20, 25]}
apiVersion: v1
kind: Service
metadata:
  name: nginx-gateway
  namespace: nginx-gateway
  labels:
    app.kubernetes.io/name: nginx-gateway-fabric
    app.kubernetes.io/instance: ngf
    app.kubernetes.io/version: "1.4.0"
spec:
  type: NodePort
  selector:
    app.kubernetes.io/name: nginx-gateway-fabric
    app.kubernetes.io/instance: ngf
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
    nodePort: 31437
  - name: https
    port: 443
    protocol: TCP
    targetPort: 443
    nodePort: 31438
```

Apply it using `kubectl`:

```shell
kubectl apply -f nodeport-config.yaml
```
```text
service/nginx-gateway created
```

{{< warning >}}
The NodePort resource must be deployed in the same namespace as NGINX Gateway Fabric.

If you are making customizations, ensure your `labels:` and `selectors:` also match the labels on the NGINX Gateway Fabric Deployment.
{{< /warning >}}

---

## Create an example application

In the previous section, you deployed NGINX Gateway Fabric to a local cluster. This section shows you how to deploy a simple web application to test that NGINX Gateway Fabric works.

{{< note >}}
The YAML code in the following sections can be found in the [cafe-example folder](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples/cafe-example) of the GitHub repository.
{{< /note >}}

---

### Create the application resources

Create the file _cafe.yaml_ with the following contents:

{{< ghcode "https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/refs/heads/main/examples/cafe-example/cafe.yaml">}}

Apply it:

```shell
kubectl apply -f cafe.yaml
```
```text
deployment.apps/coffee created
service/coffee created
deployment.apps/tea created
service/tea created
```

Verify that the new pods are in the `default` namespace:

```shell
kubectl -n default get pods
```
```text
NAME                      READY   STATUS    RESTARTS   AGE
coffee-6db967495b-dvg5w   1/1     Running   0          80s
tea-7b7d6c947d-8xmhm      1/1     Running   0          80s
```

---

### Create Gateway and HTTPRoute resources

Create the file _gateway.yaml_ with the following contents:

{{< ghcode "https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/refs/heads/main/examples/cafe-example/gateway.yaml">}}

Apply it using `kubectl`:

```shell
kubectl apply -f gateway.yaml
```
```text
gateway.gateway.networking.k8s.io/gateway created
```

Create the file _cafe-routes.yaml_ with the following contents:

{{< ghcode "https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/refs/heads/main/examples/cafe-example/cafe-routes.yaml">}}

Apply it using `kubectl`:

```shell
kubectl apply -f cafe-routes.yaml
```
```text
httproute.gateway.networking.k8s.io/coffee created
httproute.gateway.networking.k8s.io/tea created
```

You can check that all of the expected services are available using `kubectl get`:

```shell
kubectl get service --all-namespaces
```
```text
NAMESPACE       NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
default         coffee          ClusterIP   10.96.18.163    <none>        80/TCP                       2m51s
default         kubernetes      ClusterIP   10.96.0.1       <none>        443/TCP                      4m41s
default         tea             ClusterIP   10.96.169.132   <none>        80/TCP                       2m51s
kube-system     kube-dns        ClusterIP   10.96.0.10      <none>        53/UDP,53/TCP,9153/TCP       4m40s
nginx-gateway   nginx-gateway   NodePort    10.96.186.45    <none>        80:31437/TCP,443:31438/TCP   3m6s
```

---

## Test NGINX Gateway Fabric

The cluster was configured with port `8080` as the `containerPort` value, alongside the `nodePort` value of the NodePort service.

Since the NodePort `targetPort` values match the _tea_ and _coffee_ service `port` values, no port forwarding is required.

You can use `curl` to test the new services by targeting the hostname (_cafe.example.com_) with the _/coffee_ and _/tea_ paths:

```shell
curl --resolve cafe.example.com:8080:127.0.0.1 http://cafe.example.com:8080/coffee
```
```text
Server address: 10.244.0.6:8080
Server name: coffee-6db967495b-984mx
Date: 17/Oct/2024:15:50:22 +0000
URI: /coffee
Request ID: 8ad83b06ea42b996ad6bd28032b38e28
```

```shell
curl --resolve cafe.example.com:8080:127.0.0.1 http://cafe.example.com:8080/coffee
```
```text
Server address: 10.244.0.6:8080
Server name: coffee-6db967495b-984mx
Date: 17/Oct/2024:15:50:29 +0000
URI: /coffee
Request ID: 2dfc85564dd5b5dad7e62b980bb60ee1
```

---

## Additional reading

- [Install NGINX Gateway Fabric](http://localhost:1313/nginx-gateway-fabric/installation/installing-ngf/), for additional ways to install NGINX Gateway Fabric
- [How-to guides](http://localhost:1313/nginx-gateway-fabric/how-to/), for configuring your cluster
