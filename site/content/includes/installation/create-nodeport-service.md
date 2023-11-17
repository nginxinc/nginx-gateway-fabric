---
docs:
---

To create a **NodePort** service:

```shell
kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/nodeport.yaml
```

A **NodePort** service allocates a port on every cluster node. Access NGINX Gateway Fabric using any node's IP address and the allocated port.
