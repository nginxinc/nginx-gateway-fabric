---
docs: "DOCS-000"
---

Installing Gateway API resources from the experimental channel includes everything in the standard release channel plus additional experimental resources and fields.
NGINX Gateway Fabric currently supports a subset of the additional features provided by the experimental channel. In order
to use these experimental resources, the [Gateway API resources](https://github.com/kubernetes-sigs/gateway-api) from the experimental channel must be installed before deploying NGINX Gateway Fabric.
Additionally, NGINX Gateway Fabric must have experimental features enabled.

{{<caution>}}As noted in the [Gateway API documentation](https://gateway-api.sigs.k8s.io/guides/#install-experimental-channel), future releases of the Gateway API can include breaking changes to experimental resources and fields.{{</caution>}}

To install the Gateway API resources from the experimental channel, run the following:

```shell
kubectl kustomize "https://github.com/nginxinc/nginx-gateway-fabric/config/crd/gateway-api/experimental?ref=v1.4.0" | kubectl apply -f -
```

{{<note>}}If you plan to use the `edge` version of NGINX Gateway Fabric, you can replace the version in `ref` with `main`, for example `ref=main`.{{</note>}}


To enable experimental features on NGINX Gateway Fabric:
