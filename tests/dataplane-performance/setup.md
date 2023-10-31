# Dataplane Performance Testing

## Goals

- To capture the average latency of requests being proxied through NGINX using a variety of routing rules, so that we
  can see if different routing rules provide different results, and so that we can know if any future work has an impact
  on data plane performance.
- A route is created and tested for each routing method below:
  - path based routing
  - header based routing
  - query param based routing
  - method based routing

## Test Environment

- A Kubernetes cluster with 3 nodes on GKE
  - Node: e2-medium (2 vCPU, 4GB memory)
- Tester VM on Google Cloud:
  - Instance Type: e2-medium (2 vCPU, 4GB memory)
  - Configuration:
    - Debian
    - Install packages: wrk
  - Location - same zone as the Kubernetes cluster.

## Setup

1. Create cloud cluster
2. Deploy CRDs:

   ```bash
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

3. Deploy NGF from edge using Helm install, and expose using an internal LoadBalancer service:

   ```console
   helm install my-release oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric  --version 0.0.0-edge \
      --set service.annotations.'networking\.gke\.io\/load-balancer-type'="Internal" --create-namespace \
      --wait -n nginx-gateway
   ```

## Tests

1. First create the test resources. The manifests provided will deploy a single application and service, and two
   HTTPRoutes with matches for path based, header based, query based, and method based routing.

   ```console
   kubectl apply -f manifests/gateway.yaml
   kubectl apply -f manifests/coffee.yaml
   kubectl apply -f manifests/cafe-routes.yaml
   ```

2. Get the external IP of the nginx-gateway service and add an entry in /etc/hosts for `<GW_IP> cafe.example.com`.

   ```console
   kubectl get service my-release-nginx-gateway-fabric \
   --namespace nginx-gateway \
   --output jsonpath='{.status.loadBalancer.ingress[0].ip}'
   ```

3. Copy the scripts in the `scripts/` directory to the test VM. Run tests using `wrk` (change the file output name to
   version under test). The tests use `wrk` to send requests to coffee application using each of the configured routing
   rules, capturing the number of requests and the average latency.

   ```console
   bash wrk-latency.sh > 1.0.0.md
   ```

4. Analyse the results and check for anomolies. Copy the results file from the test VM into the `results/` directory.
   Append any findings or observations to the generated results document. Add test environment information to the
   generated document.

5. Cleanup the deployed resources

   ```console
   kubectl delete -f manifests/cafe-routes.yaml
   kubectl delete -f manifests/coffee.yaml
   kubectl delete -f manifests/gateway.yaml
   helm uninstall my-release -n nginx-gateway
   ```
