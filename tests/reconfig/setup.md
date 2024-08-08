# Reconfig tests

<!-- TOC -->
- [Reconfig tests](#reconfig-tests)
  - [Goals](#goals)
  - [Test Environment](#test-environment)
  - [Setup](#setup)
  - [Tests](#tests)
    - [Test 1: Resources exist before start-up](#test-1-resources-exist-before-start-up)
    - [Test 2: Start NGF, deploy Gateway, create many resources attached to GW](#test-2-start-ngf-deploy-gateway-create-many-resources-attached-to-gw)
    - [Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway](#test-3-start-ngf-create-many-resources-attached-to-a-gateway-deploy-the-gateway)
<!-- TOC -->

## Goals

- Measure how long it takes NGF to reconfigure NGINX and update statuses when a number of Gateway API and
  referenced core Kubernetes resources are created at once.
- Two runs of each test should be ran with differing numbers of resources. Each run will deploy:
  - a single Gateway, Secret, and ReferenceGrant resources
  - `x+1` number of namespaces
  - `2x` number of backend apps and services
  - `3x` number of HTTPRoutes.
- Where x=30 OR x=150.

## Test Environment

The following cluster will be sufficient:

- A Kubernetes cluster with 4 nodes on GKE
  - Node: e2-medium (2 vCPU, 4GB memory)

## Setup

1. Create cloud cluster
2. Install Gateway API Resources:

   ```bash
   kubectl kustomize config/crd/gateway-api/standard | kubectl apply -f -
   ```

3. Deploy NGF from edge using Helm install and wait for LoadBalancer Service to be ready
   (NOTE: For Test 1, deploy AFTER resources):

   ```console
   helm install my-release oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric  --version 0.0.0-edge \
      --create-namespace --wait -n nginx-gateway --set nginxGateway.productTelemetry.enable=false
   ```

4. Run tests:
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
5. After each individual test:

   - Describe the Gateway resource and make sure the status is correct.
   - Check the logs of both NGF containers for errors.
   - Parse the logs for TimeToReady numbers (see steps 6-7 below).
   - Grab metrics.
     Note: You can expose metrics by running the below snippet and then navigating to `127.0.0.1:9113/metrics`:

     ```console
     GW_POD=$(kubectl get pods -n nginx-gateway | sed -n '2s/^\([^[:space:]]*\).*$/\1/p')
     kubectl port-forward $GW_POD -n nginx-gateway 9113:9113 &
     ```

6. Measure NGINX Reloads and Time to Ready Results
   1. TimeToReadyTotal as described in each test - NGF logs.
   2. TimeToReadyAvgSingle which is the average time between updating any resource and the
      NGINX configuration being reloaded - NGF logs.
   3. NGINX Reload count - metrics.
   4. Average NGINX reload duration - metrics.
      1. The average reload duration can be computed by taking the `nginx_gateway_fabric_nginx_reloads_milliseconds_sum`
         metric value and dividing it by the `nginx_gateway_fabric_nginx_reloads_milliseconds_count` metric value.
7. Measure Event Batch Processing Results
   1. Event Batch Total - `nginx_gateway_fabric_event_batch_processing_milliseconds_count` metric.
   2. Average Event Batch Processing duration - metrics.
      1. The average event batch processing duration can be computed by taking the `nginx_gateway_fabric_event_batch_processing_milliseconds_sum`
         metric value and dividing it by the `nginx_gateway_fabric_event_batch_processing_milliseconds_count` metric value.
8. For accuracy, repeat the test suite once or twice, take the averages, and look for any anomalies or outliers.

## Tests

### Test 1: Resources exist before start-up

1. Deploy Gateway resources before start-up:
   1. Use either of the provided scripts with the required number of resources,
      e.g. `cd scripts && bash create-resources-gw-last.sh 30`. The script will deploy backend apps and services, wait
      60 seconds for them to be ready, and deploy 1 Gateway, 1 RefGrant, 1 Secret, and HTTPRoutes.
   2. Deploy NGF
   3. Measure TimeToReadyTotal as the time it takes from start-up -> final config written and
      NGINX reloaded. Measure the other results as described in steps 6-7 of the [Setup](#setup) section.

### Test 2: Start NGF, deploy Gateway, create many resources attached to GW

1. Deploy all Gateway resources, NGF running:
   1. Deploy NGF
   2. Run the provided script with the required number of resources,
      e.g. `cd scripts && bash create-resources-routes-last.sh 30`. The script will deploy backend apps and services,
      wait 60 seconds for them to be ready, and deploy 1 Gateway, 1 Secret, 1 RefGrant, and HTTPRoutes at the same time.
   3. Measure TimeToReadyTotal as the time it takes from NGF receiving the first HTTPRoute resource update (logs will say "reconciling") -> final
      config written and NGINX reloaded. Measure the other results as described in steps 6-7 of the [Setup](#setup) section.

### Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway

1. Deploy HTTPRoute resources, NGF running, Gateway last:
   1. Deploy NGF
   2. Run the provided script with the required number of resources,
      e.g. `cd scripts && bash create-resources-gw-last.sh 30`.
      The script will deploy the namespaces, backend apps and services, 1 Secret, 1 ReferenceGrant, and the HTTPRoutes;
      wait 60 seconds for the backend apps to be ready, and then deploy 1 Gateway for all HTTPRoutes.
   3. Measure TimeToReadyTotal as the time it takes from NGF receiving gateway resource -> config written and NGINX reloaded.
      Measure the other results as described in steps 6-7 of the [Setup](#setup) section.
