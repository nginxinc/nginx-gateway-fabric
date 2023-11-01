# Zero-Downtime Scaling

This document describes a test plan for testing zero-downtime scaling of NGF.

*Zero-downtime scaling* means that when the number of replicas of NGF is scaled up or down, the clients don't experience
any interruptions to the traffic they send to applications exposed via NGF.

<!-- TOC -->
- [Zero-Downtime Scaling](#zero-downtime-scaling)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
  - [Test Environment](#test-environment)
  - [Tests](#tests)
    - [Setup](#setup)
    - [Scale Gradually](#scale-gradually)
    - [Scale Abruptly](#scale-abruptly)
    - [Analyze](#analyze)
<!-- TOC -->

## Goals

- Ensure that scaling up NGF doesn't lead to any loss of traffic flowing through the data plane.
- Ensure that scaling down NGF doesn't lead to any loss of traffic flowing through the data plane.
- Ensure that after scaling, NGF can process changes to resources.

## Non-Goals

- Testing long-lived connections.

  When scaling the number of NGF replicas down, Kubernetes will shut down existing NGF pods by sending a SIGTERM. If the
  pod doesn't terminate in 30 seconds (the default period), Kubernetes will send a SIGKILL.

  When proxying Websocket or any long-lived connections, NGINX will not terminate until that connection is closed by
  either the client or the backend. This means that unless all those connections are closed by clients/backends before
  or during an upgrade (which is highly unlikely), NGINX will not terminate, which means Kubernetes will kill NGINX. As
  a result, the clients will see the connections abruptly closed and thus experience downtime.

  As a result, we *will not* use any long-live connections in this test, because NGF cannot support zero-downtime
  upgrades in this case.

- Scaling to over 25 replicas.

## Test Environment

The scaling tests will be performed on two different Kubernetes clusters:

- A Kubernetes cluster with 10 nodes on GKE:
  - Node: n2d-standard-4 (4 vCPU, 8GB memory).
  - Enabled GKE logging.
- A Kubernetes cluster with 25 nodes on GKE:
  - Node: n2d-standard-4 (4 vCPU, 8GB memory).
  - Enabled GKE logging.

The 10 Node cluster will cover the case where multiple NGF Pods are running on the same Node, whereas the 25 Node
cluster will test 1 NGF Pod per Node.

The rest of the test environment is the same for both cases:

- Tester VMs:
  - Configuration:
    - Debian
    - Install packages: wrk, curl, gnuplot.
  - Location - same zone as the Kubernetes cluster.
  - First VM - for HTTP traffic.
  - Second VM - for sending HTTPs traffic.
- NGF
  - Deployment with 1 replicas (during the test we will scale this up to a max of 25 replicas).
  - Exposed via a Service with type LoadBalancer, private IP.
  - Configure delayed termination with a 40-second sleep.
  - Gateway, two listeners - HTTP and HTTPs.
  - Two apps:
    - Coffee - 10 replicas.
    - Tea - 10 replicas.
  - Two HTTPRoutes
    - Coffee (HTTP).
    - Tea (HTTPS).

Notes:

- For sending traffic, we will use both wrk and curl.
  - *wrk* will generate a lot of traffic continuously and it will have a high chance of catching any
    (however small) periods of downtime.
  - *curl* will generate 1 request every 0.1s. While it might not catch small periods of downtime, it will give us
    a timeline of failed requests for big periods of downtime, which wrk doesn't do.

## Tests

Run the following tests on both the 10 and 25 node clusters.

### Setup

1. Create a cluster.
2. Deploy a previous latest stable version with 1 replica

   For the 25 node cluster:

    ```console
     helm install my-release oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric  --create-namespace --wait -n nginx-gateway --version 0.0.0-edge --values ./values-25-node.yaml
    ```

   For the 10 node cluster:

    ```console
     helm install my-release oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric  --create-namespace --wait -n nginx-gateway --version 0.0.0-edge --values ./values-10-node.yaml
    ```

3. Deploy backend apps:

    ```console
    kubectl apply -f manifests/cafe.yaml
    ```

4. Configure Gateway:

    ```console
    kubectl apply -f manifests/cafe-secret.yaml
    kubectl apply -f manifests/gateway.yaml
    ```

5. Expose apps via HTTPRoutes

    ```console
    kubectl apply -f manifests/cafe-routes.yaml
    ```

6. Check statuses of the Gateway and HTTPRoutes for errors.
7. In Google Monitoring, check NGF and NGINX error logs for errors.
8. In Tester VMs, update `/etc/hosts` to have an entry with the External IP of the NGF Service (`10.128.0.10` in this
   case):

   ```text
   10.128.0.10 cafe.example.com
   ```

### Scale Gradually

1. Scale up
   1. Start sending traffic using wrk from tester VMs for 5 minutes:
      - Tester VM 1:
        - wrk:

          ```console
          wrk -t2 -c2 -d5m --latency --timeout 2s  http://cafe.example.com/coffee
          ```

        - curl:

          ```console
          for i in `seq 1 3000`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -sS --connect-timeout 2 http://cafe.example.com/coffee 2>&1  && sleep 0.1s; done > results.txt
          ```

      - Tester VM 2:
        - wrk:

          ```console
          wrk -t2 -c2 -d5m --latency --timeout 2s  https://cafe.example.com/tea
          ```

        - curl:

          ```console
          for i in `seq 1 3000`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -k -sS --connect-timeout 2 https://cafe.example.com/tea 2>&1  && sleep 0.1s; done > results.txt
          ```

   2. **Immediately** scale up:

       ```console
       ./scale.sh up
       ```

      This script scales NGF to 25 Pods one at a time. After increasing the number of Pods by one, it waits for all Pods
      to be ready, updates the Gateway resource, and then waits for the Gateway status observed generation to increment
      by one. If after 60 seconds, the Gateway status observed generation has not been incremented, the script will
      exit. It prints the Gateway status after the observed generation is incremented. Check this to make sure the
      Gateway is Accepted and Programmed. You can redirect the script's output to a file if you don't want to watch it
      execute.

   3. Once `wrk` finishes, kill the curl for loops.

   4. [Analyze](#analyze) the results.
2. Scale down
   1. Start sending traffic using wrk from tester VMs for 20 minutes:
      - Tester VM 1:
        - wrk:

          ```console
          wrk -t2 -c2 -d20m --latency --timeout 2s  http://cafe.example.com/coffee
          ```

        - curl:

          ```console
          for i in `seq 1 12000`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -sS --connect-timeout 2 http://cafe.example.com/coffee 2>&1  && sleep 0.1s; done > results.txt
          ```

      - Tester VM 2:
        - wrk:

          ```console
          wrk -t2 -c2 -d20m --latency --timeout 2s  https://cafe.example.com/tea
          ```

        - curl:

          ```console
          for i in `seq 1 12000`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -k -sS --connect-timeout 2 https://cafe.example.com/tea 2>&1  && sleep 0.1s; done > results.txt
          ```

   2. **Immediately** scale down:

       ```console
       ./scale.sh down
       ```

      This script scales NGF from 25 Pods to 1 Pod one at a time. After decreasing the number of Pods by one, it sleeps
      for 40 seconds to wait for the Pod to terminate. Then, it updates the Gateway resource and waits for the Gateway
      status observed generation to increment by one. If after 60 seconds, the Gateway status observed generation has
      not been incremented, the script will exit. It prints the Gateway status after the observed generation is
      incremented. Check this to make sure the Gateway is Accepted and Programmed. You can redirect the script's output
      to a file if you don't want to watch it execute.

   3. Once the scale script finishes, kill wrk and the curl for loops.

   4. [Analyze](#analyze) the results.

### Scale Abruptly

1. Scale up
   1. Start sending traffic using wrk from tester VMs for 2 minutes:
      - Tester VM 1:
        - wrk:

          ```console
          wrk -t2 -c2 -d2m --latency --timeout 2s  http://cafe.example.com/coffee
          ```

        - curl:

          ```console
          for i in `seq 1 600`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -sS --connect-timeout 2 http://cafe.example.com/coffee 2>&1  && sleep 0.1s; done > results.txt
          ```

      - Tester VM 2:
        - wrk:

          ```console
          wrk -t2 -c2 -d2m --latency --timeout 2s  https://cafe.example.com/tea
          ```

        - curl:

          ```console
          for i in `seq 1 600`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -k -sS --connect-timeout 2 https://cafe.example.com/tea 2>&1  && sleep 0.1s; done > results.txt
          ```

   2. Scale NGF up to 25 Pods:

      ```console
      kubectl scale -n nginx-gateway deployment <NGF deployment name> --replicas 25
      ```

   3. Update the Gateway

      ```console
      kubectl apply -f manifests/gateway-2.yaml
      ```

   4. Check the status of the Gateway and make sure it's been updated (there should be three listeners)

      ```console
      kubectl describe gateway
      ```

   5. [Analyze](#analyze) the results.
2. Scale down
   1. Start sending traffic using wrk from tester VMs for 2 minutes:
      - Tester VM 1:
        - wrk:

          ```console
          wrk -t2 -c2 -d2m --latency --timeout 2s  http://cafe.example.com/coffee
          ```

        - curl:

          ```console
          for i in `seq 1 600`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -sS --connect-timeout 2 http://cafe.example.com/coffee 2>&1  && sleep 0.1s; done > results.txt
          ```

      - Tester VM 2:
        - wrk:

          ```console
          wrk -t2 -c2 -d2m --latency --timeout 2s  https://cafe.example.com/tea
          ```

        - curl:

          ```console
          for i in `seq 1 600`; do printf  "\nRequest $i\n" && date --rfc-3339=ns && curl -k -sS --connect-timeout 2 https://cafe.example.com/tea 2>&1  && sleep 0.1s; done > results.txt
          ```

   2. Scale NGF down to 1 Pods:

      ```console
      kubectl scale -n nginx-gateway deployment <NGF deployment name> --replicas 1
      ```

   3. Update the Gateway

      ```console
      kubectl apply -f manifests/gateway-1.yaml
      ```

   4. Check the status of the Gateway and make sure it's been updated (there should be two listeners)

      ```console
      kubectl describe gateway
      ```

   5. [Analyze](#analyze) the results.

### Analyze

After each test analyze and record the following:

- Tester VMs:
  - Analyze the output of wrk commands for errors and latencies.
    - Create graphs from curl output.
      See [graphing instructions](/tests/zero-downtime-upgrades/zero-downtime-upgrades.md#converting-curl-output-to-a-graph).
    - Check logs
      - NGINX Access logs - we expect only 200 responses. Google Monitoring query:

          ```text
          resource.type="k8s_container"
          resource.labels.cluster_name="<your cluster name>"
          resource.labels.container_name="nginx"
          resource.labels.namespace_name="nginx-gateway"
          labels.k8s-pod/app_kubernetes_io/instance="<your instance name>"
          severity=INFO
          SEARCH("200")
          ```

          To check for non-200 responses:

         ```text
         resource.type="k8s_container"
         resource.labels.cluster_name="<your cluster name>"
         resource.labels.container_name="nginx"
         resource.labels.namespace_name="nginx-gateway"
         labels.k8s-pod/app_kubernetes_io/instance="<your instance name>"
         severity=INFO
         NOT SEARCH("200")
         ```

      - NGINX Error logs - we expect no errors or warnings:
        Google Monitoring query:

          ```text
          resource.type="k8s_container"
          resource.labels.cluster_name="<your cluster name>"
          resource.labels.container_name="nginx"
          resource.labels.namespace_name="nginx-gateway"
          labels.k8s-pod/app_kubernetes_io/instance="<your instance name>"
          severity=ERROR
          SEARCH("`[warn]`") OR SEARCH("`[error]`")
          ```

      - NGF logs - we expect no errors

        ```text
        resource.type="k8s_container"
        resource.labels.cluster_name="<your cluster name>"
        resource.labels.container_name="nginx-gateway"
        resource.labels.namespace_name="nginx-gateway"
        labels.k8s-pod/app_kubernetes_io/instance="<your instance name>"
        severity="ERROR"
        SEARCH("error")
        ```
