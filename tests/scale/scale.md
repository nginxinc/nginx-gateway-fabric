# Scale Tests

This document describes how we scale test NGF.

<!-- TOC -->
- [Scale Tests](#scale-tests)
  - [Goals](#goals)
  - [Test Environment](#test-environment)
  - [Steps](#steps)
    - [Setup](#setup)
    - [Run the tests](#run-the-tests)
      - [Scale Listeners to Max of 64](#scale-listeners-to-max-of-64)
      - [Scale HTTPS Listeners to Max of 64](#scale-https-listeners-to-max-of-64)
      - [Scale HTTPRoutes](#scale-httproutes)
      - [Scale Upstream Servers](#scale-upstream-servers)
      - [Scale HTTP Matches](#scale-http-matches)
    - [Analyze](#analyze)
    - [Results](#results)
<!-- TOC -->

## Goals

- Measure how NGF performs when the number of Gateway API and referenced core Kubernetes resources are scaled.
- Test the following number of resources:
  - Max number of HTTP and HTTPS Listeners (64)
  - Max number of Upstream Servers (648)
  - Max number of HTTPMatches
  - 1000 HTTPRoutes

## Test Environment

For most of the tests, the following cluster will be sufficient:

- A Kubernetes cluster with 4 nodes on GKE
  - Node: n2d-standard-8 (8 vCPU, 32GB memory)
  - Enabled GKE logging

The Upstream Server scale test requires a bigger cluster to accommodate the large number of Pods. Those cluster details
are listed in the [Scale Upstream Servers](#scale-upstream-servers) test steps.

## Steps

### Setup

- Install Gateway API Resources:

  ```console
  kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
  ```

- Install edge NGF and save the Pod Name and LoadBalancer IP for tests:

  ```console
  helm install scale-test oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric  --create-namespace --wait -n nginx-gateway --version=0.0.0-edge
  ```

  ```console
  export NGF_IP=$(kubectl get svc -n nginx-gateway scale-test-nginx-gateway-fabric --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
  export NGF_POD=$(kubectl get pods -n nginx-gateway -l "app.kubernetes.io/name=nginx-gateway-fabric,app.kubernetes.io/instance=scale-test" -o jsonpath="{.items[0].metadata.name}")
  ```

- Install Prometheus:

  ```console
  kubectl apply -f manifets/prom-clusterrole.yaml
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
  helm install prom prometheus-community/prometheus --set useExistingClusterRoleName=prometheus -n prom
  ```

- Create a directory under [results](/tests/scale/results) and name it after the version of NGF you are testing. Then
  create a file for the result summary, also named after the NGF version. For
  example: [1.0.0.md](/tests/scale/results/1.0.0/1.0.0.md).

### Run the tests

#### Scale Listeners to Max of 64

Test Goal: Measure how NGF performs as the number of Listeners increases to the max of 64.

Test Plan:

- Scale up to 64 HTTP Listeners
- All Listeners are on a single Gateway
- Each Listener has 1 HTTPRoute attached
- Each HTTPRoute references 1 unique Service
- Services and Deployments are created before scaling Listeners.
- After each Listener + HTTPRoute is created, measure the time it takes to get a successful response from the new
  route (time to ready).
- Record the time to ready in seconds in a csv file for each iteration.

Total Resources Created:

- 1 Gateway with 64 Listeners
- 64 HTTPRoutes
- 64 Services, Deployments, Pods

Follow the steps below to run the test:

- Run the test:

  ```console
  go test -v -tags scale -run TestScale_Listeners -i 64
  ```

- [Analyze](#analyze) the results.

- Clean up::

  Delete resources from cluster:

  ```console
  kubectl delete -Rf TestScale_Listeners
  ```

  Delete generated manifests:

  ```console
  rm -rf TestScale_Listeners
  ```

- Check for any errors or restarts after cleanup.
- Check NGINX conf to make sure it looks correct.

#### Scale HTTPS Listeners to Max of 64

Test Goal: Measure how NGF performs as the number of HTTPS Listeners increases to the max of 64.

Test Plan:

- Scale up to 64 HTTPS Listeners
- All Listeners are on a single Gateway
- Each Listener has 1 HTTPRoute attached
- Each Listener references a unique Secret
- Each HTTPRoute references 1 unique Service
- Services, Deployments, and Secrets are created before scaling Listeners
- After each Listener + HTTPRoute is created, measure the time it takes to get a successful response from the new
  route (time to ready).
- Record the time to ready in seconds in a csv file for each iteration.

Total Resources Created:

- 1 Gateway with 64 HTTPS Listeners
- 64 Secrets
- 64 HTTPRoutes
- 64 Services, Deployments, Pods

Follow the steps below to run the test:

- Run the test:

  ```console
  go test -v -tags scale -run TestScale_HTTPSListeners -i 64
  ```

- [Analyze](#analyze) the results.

- Clean up:

  Delete resources from cluster:

  ```console
  kubectl delete -Rf TestScale_HTTPSListeners
  ```

  Delete generated manifests:

  ```console
  rm -rf TestScale_HTTPSListeners
  ```

- Check for any errors or restarts after cleanup.
- Check NGINX conf to make sure it looks correct.

#### Scale HTTPRoutes

Test Goal: Measure how NGF performs as the number of HTTPRoutes increases to 1000.

Test Plan:

- Scale up to 1000 HTTPRoutes
- All HTTPRoutes attach to a single Gateway with one Listener
- Each HTTPRoute references the same Service
- Service and Deployment are created before scaling HTTPRoutes
- After each HTTPRoute is created, measure the time it takes to get a successful response from the new route (time to
  ready).
- Record the time to ready in seconds in a csv file for each iteration.

Total Resources Created:

- 1 Gateway with 1 Listener
- 1000 HTTPRoutes
- 1 Service, Deployment, Pod

This test takes around 7 hours to run, so I recommend running it on a VM, or overnight with the aid of
[caffeinate](https://www.theapplegeek.co.uk/blog/caffeinate) for MAC users.

Follow the steps below to run the test:

- Run the test:

  ```console
  go test -v -tags scale -timeout 600m -run TestScale_HTTPRoutes -i 1000 -delay 2s
  ```

- [Analyze](#analyze) the results.

- Clean up:

  Delete resources from cluster:

  ```console
  kubectl delete -Rf TestScale_HTTPRoutes
  ```

  Delete generated manifests:

  ```console
  rm -rf TestScale_HTTPRoutes
  ```

- Check for any errors or restarts after cleanup.
- Check NGINX conf to make sure it looks correct.

#### Scale Upstream Servers

Test Goal: Measure how NGF performs as the number of Upstream Servers increases to the max of 648.

Test Plan:

- Deploy a single Gateway with 1 Listener and attach one HTTPRoute that references a single Service
- Scale the deployment for that Service to 648 Pods (this is the limit that the upstream zone size allows)
- Gateway, HTTPRoute, Service, and Deployment with 1 replica are created before scaling up to 648 replicas.

Total Resources Created:

- 1 Gateway with 1 Listener
- 1 HTTPRoutes
- 1 Service, 1 Deployment, 648 Pods

Test Environment:

For this test you must use a much bigger cluster in order to create 648 Pods.

- A Kubernetes cluster with 12 nodes on GKE
  - Node: n2d-standard-16 (16 vCPU, 64GB memory)
  - Enabled GKE logging

Follow the steps below to run the test:

- Apply manifest

  ```console
  kubectl apply -f manifests/scale-upstreams.yaml
  ```

- Check the status of the Gateway and HTTPRoute to make sure everything is OK before scaling.

  ```console
  kubectl describe gateway gateway
  kubectl describe httproute route
  ```

- Get the start time as a UNIX timestamp and record it in the results.

  ```console
  date +%s
  ```

  This will be used in the metrics query.

- Open a new terminal window and start the following loop:

  ```console
  for i in $(seq 1 150); do curl --resolve cafe.example.com:80:$NGF_IP http://cafe.example.com:80/; sleep 1; done >> requests.log
  ```

- Back in your original terminal, scale the backend app:

  ```console
  kubectl scale deploy backend --replicas 648
  ```

- Wait for all Pods to become available:

  ```console
  watch kubectl get deploy backend
  ```

- Check the NGINX config for 648 upstream servers:

  ```console
  kubectl exec -it -n nginx-gateway $NGF_POD -c nginx -- nginx -T | grep -E "server (?:[0-9]{1,3}\.){3}[0-9]{1,3}:8080" | wc -l
  ```

- Get the end time as a UNIX timestamp and make a note of it:

  ```console
  date +%s
  ```

- In the terminal you started the request loop, kill the loop if it's still running and check the request.log to see if
  any of the requests failed. Record any failures in the results file.

- [Analyze](#analyze) the results. Use the start time and end time you made note of earlier for the
  queries. You can calculate the test duration in seconds by subtracting the start time from the end time.

- Clean up:

  ```console
  kubectl delete -f manifests/scale-upstreams.yaml
  ```

- Check for any errors or restarts after cleanup.
- Check NGINX conf to make sure it looks correct.

#### Scale HTTP Matches

Test Goal: Find the difference in latency between the first match and last match for the max length of
the `http_matches` variable.

Test Plan:

- Deploy a single Gateway with 1 Listener and attach one HTTPRoute that references a single Service
- Within the HTTPRoute configure the max number of matches (max is determined by the length of the
  generated `http_matches` variable (4096 characters))
- Use `wrk` to send requests to the _first_ match in `http_matches` list and measure the latency
- Use `wrk` to send requests to the _last_ match in `http_matches` list and measure the latency

Total Resources Created:

- 1 Gateway with 1 Listener
- 1 HTTPRoute with 7 rules and 50 matches
- 1 Service, 1 Deployment, 1 Pod

Follow these steps to run the test:

- Download [wrk](https://github.com/wg/wrk)

- Apply manifest:

  ```console
  kubectl apply -f manifests/scale-matches.yaml
  ```

- Check the status of the Gateway and HTTPRoute to make sure everything is OK before scaling.

  ```console
  kubectl describe gateway gateway
  kubectl describe httproute route
  ```

- Test the first match:

  ```console
  ./wrk -t2 -c10 -d30 http://cafe.example.com -H "header-1: header-1-val"
  ```

- Test the last match:

  ```console
   ./wrk -t2 -c10 -d30 http://cafe.example.com -H "header-50: header-50-val"
  ```

- Copy and paste the results into the results file.

- Clean up::

  ```console
  kubectl delete -f manifests/scale-matches.yaml
  ```

### Analyze

- Query Prometheus for reload metrics. To access the Prometheus Server, run:

  ```console
  export POD_NAME=$(kubectl get pods --namespace prom -l "app.kubernetes.io/name=prometheus,app.kubernetes.io/instance=prom" -o jsonpath="{.items[0].metadata.name}")
  kubectl --namespace prom port-forward $POD_NAME 9090
  ```

  To query Prometheus, you can either browse to localhost:9090 or use curl. The following instructions assume you are
  using the prom GUI.

  > Note:
  > For the tests that write to a csv file, the `Test Start`, `Test End + 10s`, and `Duration` are at the
  > end of the results.csv file in the `results/<NGF_VERSION>/<TEST_NAME>` directory.
  > We are using `Test End + 10s` in the Prometheus query to account for the 10s scraping interval.

  Total number of reloads:

    ```console
    nginx_gateway_fabric_nginx_reloads_total - nginx_gateway_fabric_nginx_reloads_total @ <Test Start>
    ```

  Total number of reload errors:

    ```console
    nginx_gateway_fabric_nginx_reload_errors_total - nginx_gateway_fabric_nginx_reload_errors_total @ <Test Start>
    ```

  Average reload time (ms):

    ```console
    rate(nginx_gateway_fabric_nginx_reloads_milliseconds_sum[<Duration>] @ <Test End + 10s>) /
    rate(nginx_gateway_fabric_nginx_reloads_milliseconds_count[<Duration>] @ <Test End + 10s>)
    ```

  Record these numbers in a table in the results file.

- Take screenshots of memory and CPU usage in GKE Dashboard

  To Monitor memory and CPU usage, navigate to the Kubernetes Engine > Workloads > Filter by `nginx-gateway` namespace >
  click on NGF Pod name. You should see graphs for CPU, Memory, and Disk.

  - Convert the `Start Time` and `End Time` UNIX timestamps to your local date time:

  ```console
  date -r <UNIX Timestamp>
  ```

  - Create a custom time frame for the graphs in GKE.
  - Take a screenshot of the CPU and Memory graphs individually. Store them in the `results/<NGF_VERSION>/<TEST_NAME>`
    directory.

- If the test writes time to ready numbers to a csv, create a time to ready graph.
  - Use https://chart-studio.plotly.com/create/#/ to plot the time to ready numbers on a graph.
    - Remove the `"Test Start", "Test End", "Test End + 10s", "Duration"` rows from the bottom of the csv.
    - Upload the csv file to plotly.
    - Create a new `Trace`, select `Line` as the type.
    - Set the Y axis to the Time to Ready column.
    - Set the X axis to the number of resources column.
    - Label the graph and take a screenshot.
    - Store the graph in the `results/<NGF_VERSION>/<TEST_NAME>` directory.

- Check for errors or restarts and record in the results file. File a bug if there's unexpected errors or restarts.
- Check NGINX conf and make sure it looks correct. File a bug if there is an issue.

### Results

- [1.0.0](/tests/scale/results/1.0.0/1.0.0.md)
