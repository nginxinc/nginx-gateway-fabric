# Zero-Downtime Upgrades

This document describes a test plan for testing zero-downtime upgrades of NGF.

*Zero-downtime upgrades* means that during an NGF upgrade clients don't experience any
interruptions to the traffic they send to applications exposed via NGF.

<!-- TOC -->

- [Zero-Downtime Upgrades](#zero-downtime-upgrades)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
  - [Test Environment](#test-environment)
  - [Steps](#steps)
    - [Start](#start)
    - [Upgrade](#upgrade)
    - [After Upgrade](#after-upgrade)
    - [Analyze](#analyze)
  - [Results](#results)
  - [Appendix](#appendix)
    - [Pod Affinity](#pod-affinity)
    - [Converting Curl Output to a Graph](#converting-curl-output-to-a-graph)

<!-- TOC -->

## Goals

- Ensure that upgrading NGF doesn't lead to any loss of traffic flowing through the data plane.
- Ensure that after an upgrade, NGF can process changes to resources.
- Detect if any special instructions will be required to provide to users to perform
  an upgrade.

## Non-Goals

During an upgrade, Kubernetes will shut down existing NGF Pods by sending a SIGTERM. If the Pod doesn't terminate in 30
seconds (the default period) , Kubernetes will send a SIGKILL.

When proxying Websocket or any long-lived connections, NGINX will not terminate until
that connection is closed by either the client or the backend. This means that unless all those connections are closed
by clients/backends before or during an upgrade (which is highly unlikely), NGINX will not terminate, which means
Kubernetes will kill NGINX. As a result, the clients will see the connections abruptly closed and thus experience
downtime.

As a result, we *will not* use any long-live connections in this test, because NGF cannot support zero-downtime upgrades
in this case.

## Test Environment

- A Kubernetes cluster with 10 nodes on GKE
  - Node: e2-medium (2 vCPU, 4GB memory)
  - Enabled GKE logging.
- Tester VMs on Google Cloud Platform:
  - Configuration:
    - Debian
    - Install packages: wrk, curl, gnuplot
  - Location - same zone as the Kubernetes cluster.
  - First VM for HTTP traffic
  - Second VM - for sending HTTPs traffic
- NGF
  - Deployment with 2 replicas scheduled on different nodes.
  - Exposed via a Service with type LoadBalancer, private IP
  - Gateway, two listeners - HTTP and HTTPs
  - Two backends:
    - Coffee - 3 replicas
    - Tea - 3 replicas
  - Two HTTPRoutes
    - Coffee (HTTP)
    - Tea (HTTPS)

Notes:

- For sending traffic, we will use both wrk and curl.
  - *wrk* will generate a lot of traffic continuously, and it will have a high chance of catching of any
      (however small) periods of downtime.
  - *curl* will generate 1 request every 0.1s. While it might not catch small periods of downtime, it will
      give us timeline of failed request for big periods of downtime, which wrk doesn't do.
- We use Pod anti-affinity to tell Kubernetes to schedule NGF Pods on different nodes. We also use a 10 node cluster so
  that the chance of Kubernetes scheduling new Pods on the same
  nodes is minimal. Scheduling new Pods on different nodes will help better catch
  any interdependencies with an external load balancer (typically the node of a new Pod will be added
  to the pool in the load balancer, and the node of an old one will be removed).

## Steps

### Start

1. Create a cluster.
2. Deploy a previous latest stable version with 2 replicas with added [anti-affinity](#pod-affinity).
3. Expose NGF via a Service Load Balancer, internal (only accessible within the Google Cloud region) by adding
   `networking.gke.io/load-balancer-type: "Internal"` annotation to the Service.
4. Deploy backend apps:

    ```console
    kubectl apply -f manifests/cafe.yaml
    ```

5. Configure Gateway:

    ```console
    kubectl apply -f manifests/cafe-secret.yaml
    kubectl apply -f manifests/gateway.yaml
    ```

6. Expose apps via HTTPRoutes

    ```console
    kubectl apply -f manifests/cafe-routes.yaml
    ```

7. Check statuses of the Gateway and HTTPRoutes for errors.
8. In Google Monitoring, check NGF and NGINX error logs for errors.
9. In Tester VMs, update `/etc/hosts` to have an entry with the External IP of the NGF Service (`10.128.0.10` in this
   case):

   ```text
   10.128.0.10 cafe.example.com
   ```

### Upgrade

1. Follow the [upgrade instructions](/docs/installation.md#upgrade-nginx-gateway-fabric-from-manifests) to:
    1. Upgrade Gateway API version to the one that matches the supported version of new release.
    2. Upgrade NGF CRDs.
2. Start sending traffic using wrk from tester VMs for 1 minute:
    - Tester VM 1:
        - wrk:

          ```console
          wrk -t2 -c100 -d60s --latency --timeout 2s  http://cafe.example.com/coffee
          ```

        - curl:

          ```console
          for i in `seq 1 600`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -sS --connect-timeout 2 http://cafe.example.com/coffee 2>&1  && sleep 0.1s; done > results.txt
          ```

    - Tester VM 2:
        - wrk:

          ```console
          wrk -t2 -c100 -d60s --latency --timeout 2s  https://cafe.example.com/tea
          ```

        - curl:

          ```console
          for i in `seq 1 600`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -k -sS --connect-timeout 2 https://cafe.example.com/tea 2>&1  && sleep 0.1s; done > results.txt
          ```

3. **Immediately** upgrade NGF manifests by
   following [upgrade instructions](/docs/installation.md#upgrade-nginx-gateway-fabric-from-manifests).
   > Don't forget to modify the manifests to have 2 replicas and Pod affinity.
4. Ensure the new Pods are running and the old ones terminate.

### After Upgrade

1. Update the Gateway resource by adding one new listener `http-new`:

    ```console
    kubectl apply -f manifests/gateway-updated.yaml
    ```

2. Check that at NGF has a leader elected among the new Pods:

    ```console
    kubectl -n nginx-gateway logs <nkg-pod> | grep leader
    ```

3. Ensure the status of the Gateway resource includes the new listener.

### Analyze

- Tester VMs:
  - Analyze the output of wrk commands for errors and latencies.
  - Create graphs from curl output (see [instructions](#converting-curl-output-to-a-graph) in Appendix) and check for
      any failures on them.
- Check the old Pods logs in Google Monitoring
  - NGINX Access logs - we expect only 200 responses.
      Google Monitoring query:

      ```text
      severity=INFO
      "GET" "HTTP/1.1" -"200"
      ```

  - NGINX Error logs - we expect no errors or warnings
      Google Monitoring query:

      ```text
      severity=ERROR
      SEARCH("`[warn]`") OR SEARCH("`[error]`")
      ```

  - NGF logs - we expect no errors
  - Specifically look at the NGF logs before it exited, to make sure all components shutdown correctly.
- Check the new Pods (in Google Monitoring)
  - NGINX Access logs - only 200 responses.
  - NGINX Error logs - no errors or warnings.
  - NGF logs - no errors

## Results

- [1.0.0](results/1.0.0/1.0.0.md)

## Appendix

### Pod Affinity

- To ensure Kubernetes doesn't schedule NGF Pods on the same nodes, use an anti-affinity rule:

    ```yaml
        spec:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
              - topologyKey: kubernetes.io/hostname
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/name: nginx-gateway
    ```

### Converting Curl Output to a Graph

The output of a curl command is saved in `results.txt`. To convert it into a graph,
go through the following steps:

1. Convert the output into a csv file:

    ```console
    awk '
    /Request [0-9]+/ {
        getline
        datetime = $0
        getline
        if ($1 == "curl:") {
            print datetime ",0"  # Failed
        } else {
            print datetime ",1"  # Success
        }
    }' results.txt > results.csv
    ```

2. Plot a graph using the csv file:

    ```console
    gnuplot requests-plot.gp
    ```

   As a result, gnuplot will create `graph.png` with a graph.
3. Download the resulting `graph.png` to you local machine.
4. Also download `results.csv`.
