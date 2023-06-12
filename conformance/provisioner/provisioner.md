# Provisioner

Provisioner implements data plane provisioning for NGINX Kubernetes Gateway (NKG): it creates an NKG static mode
Deployment for each Gateway that belongs to the provisioner GatewayClass.

```
Usage:
  gateway provisioner-mode [flags]

Flags:
  -h, --help   help for provisioner-mode

Global Flags:
      --gateway-ctlr-name string   The name of the Gateway controller. The controller name must be of the form: DOMAIN/PATH. The controller's domain is 'k8s-gateway.nginx.org' (default "")
      --gatewayclass string        The name of the GatewayClass resource. Every NGINX Gateway must have a unique corresponding GatewayClass resource. (default "")
```

> Note: Provisioner is not ready for production yet (see this issue for more details
https://github.com/nginxinc/nginx-kubernetes-gateway/issues/634). However, it can be used in the Gateway API conformance
tests, which expect a Gateway API implementation to provision an independent data plane per Gateway.

> Note: Provisioner uses [this manifest](/deploy/manifests/deployment.yaml) to create an NKG static mode Deployment.
This manifest gets included into the NKG binary during the NKG build. To customize the Deployment, modify the manifest 
and **re-build** NKG.

How to deploy:

1. Follow the [installation](/docs/installation.md) instructions up until the Deploy the NGINX Kubernetes Gateway Step
   to deploy prerequisites for both the static mode Deployments and the provisioner.
1. Deploy provisioner:
   ```
   kubectl apply -f conformance/provisioner/provisioner.yaml
   ```
1. Confirm the provisioner is running in nginx-gateway namespace:
   ```
   kubectl get pods -n nginx-gateway 
   NAME                                         READY   STATUS    RESTARTS   AGE
   nginx-gateway-provisioner-6c9d9fdcb8-b2pf8   1/1     Running   0          11m
   ```
