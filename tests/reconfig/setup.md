# Reconfig tests

## Setup

1. Create cloud cluster
2. Deploy CRDs:

   ```bash
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

3. Deploy NGF from edge using Helm install (NOTE: For Test 1, deploy AFTER resources):

   ```bash
   helm install my-release oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric  --version 0.0.0-edge \
      --create-namespace --wait -n nginx-gateway
   ```

4. Optional: Add pod scrape if running in GKE (see [GKE Pod scrape config](#gke-pod-scrape-config)).
5. Run tests:
   1. There are 3 versions of the reconfiguration tests that need to be ran, with a low and high number of resources.
      Therefore, a full test suite includes 6 test runs.
   2. There are scripts to generate the required resources and config changes.
   3. Run each test using the provided script (`scripts/create-resources-gw-last.sh` or
      `scripts/create-resources-routes-last.sh` depending on the test).
   4. The scripts accept a number parameter to indicate how many resources should be created. Currently, we are running
      with 30 or 150. The scripts will create a single Gateway, Secret and ReferenceGrant resources, `x+1` number of
      namespaces, `2x` number of backend apps and services, and `3x` number of HTTPRoutes.
      - Note: Clean up after each test run for isolated results. There's a script provided for removing all the test
        fixtures `scripts/delete-multiple.sh` which takes a number (needs to be the same number as what was used in the
        create script.)
6. After each individual test run, grab logs of both NGF containers and grab metrics.
   Note: You can expose metrics by running the below snippet and then navigating to `127.0.0.1:9113/metrics`:

   ```bash
   GW_POD=$(k get pods -n nginx-gateway | sed -n '2s/^\([^[:space:]]*\).*$/\1/p')
   kubectl port-forward $GW_POD -n nginx-gateway 9113:9113 &
   ```

7. Measure Time To Ready by as described in each test, get the reload count, and get the average NGINX reload duration.
8. For accuracy, repeat the test suite once or twice, take the averages, and look for any anomolies/ outliers.

## Tests

### Test 1: Resources exist before start-up

1. Deploy Gateway resources before start-up:
   1. Use either of the provided scripts with the required number of resources,
      e.g. `cd scripts && bash create-resources-gw-last.sh 30`. The script will deploy backend apps and services, wait
      60 seconds for them to be ready, and deploy 1 Gateway, 1 RefGrant, 1 Secret, and HTTPRoutes.
   2. Deploy NFG
   3. Check logs for time takes from start-up -> config written and NGINX reloaded. Get reload count and average reload
      duration from metrics and logs.

### Test 2: Start NGF, deploy Gateway, create many resources attached to GW

1. Deploy all Gateway resources, NFG running:
   1. Deploy NFG
   2. Run the provided script with the required number of resources,
      e.g. `cd scripts && bash create-resources-routes-last.sh 30`. The script will deploy backend apps and services,
      wait 60 seconds for them to be ready, and deploy 1 Gateway, 1 Secret, 1 RefGrant, and HTTPRoutes at the same time.
   3. Check logs for time takes from NFG receiving first resource update -> final config written, and NGINX's final
      reload. Check logs for average individual HTTPRoute TTR also. Get reload count and average reload duration from
      metrics and logs.

### Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway

1. Deploy HTTPRoute resources, NFG running, Gateway last:
   1. Deploy NFG
   2. Run the provided script with the required number of resources,
      e.g. `cd scripts && bash create-resources-gw-last.sh 30`.
      The script will deploy the namespaces, backend apps and services, 1 Secret, 1 ReferenceGrant, and the HTTPRoutes;
      wait 60 seconds for the backend apps to be ready, and then deploy 1 Gateway for all HTTPRoutes.
   3. Check logs for time takes from NFG receiving gateway resource -> config written and NGINX reloaded. Get reload
      count and average reload duration from metrics and logs.

## GKE Pod scrape config

To create a Pod scrape config, you can run the following:

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: monitoring.googleapis.com/v1
  kind: PodMonitoring
  metadata:
    name: prom-example
    namespace: nginx-gateway
  spec:
    selector:
      matchLabels:
        app.kubernetes.io/name: nginx-gateway
    endpoints:
    - port: metrics
      interval: 30s
  EOF
  ```
