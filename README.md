# nginx-gateway-kubernetes

NGINX Gateway is an open source project managed by NGINX, Inc. It implements a collection of resources that model service networking in Kubernetes.

## Run NGINX Gateway 

### Prepare Kubernetes Cluster

If you'd like to use that Makefile to create a cluster, run:
   ```
   make create-kind-cluster
   ```

### Build NGINX Gateway

1. Build the image:
   ```
   make container
   ```
1. Push the image to your registry. If you're using Kind, run:
   ```
   kind load docker-image nginx-gateway:0.0.1
   ```

## Deploy NGINX Gateway

1. Install Gateway APIs CRDs:
   ```
   kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.4.0" \
   | kubectl apply -f -
   ```
1. Install NGINX Gateway CRDs:
   ```
   kubectl apply -f deploy/manifests/crds
   ```
1. Create the GatewayClass resource:
   ```
   kubectl apply -f deploy/manifests/gatewayclass.yaml 
   ```
1. Create the GatewayConfig resource:
   ```
   kubectl apply -f deploy/manifests/gatewayconfig.yaml
   ```
1. Deploy the NGINX Gateway:
   
   If you're not using Kind, before deploying, make sure to edit the Deployment spec in `nginx-gateway.yaml` so that the image for `nginx-gateway` container points to your registry. 
   ```
   kubectl apply -f deploy/manifests/nginx-gateway.yaml
   ```
1. Confirm the NGINX Gateway is running in `nginx-gateway` namespace:
   ```
   kubectl get pods -n nginx-gateway
   NAME                             READY   STATUS    RESTARTS   AGE
   nginx-gateway-5d4f4c7db7-xk2kq   2/2     Running   0          112s
   ```
   
## Use NGINX Gateway

1. Forward local port 8080 to port 80 of the NGINX Gateway pod:
   ```
   kubectl -n nginx-gateway port-forward <pod-name> 8080:80
   ```
1. Curl port 8080:
   ```
   curl localhost:8080
   hello from nginx-gateway-5d4f4c7db7-xk2kq
   ```
