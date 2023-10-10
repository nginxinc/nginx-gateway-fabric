# Longevity Test

This document describes how we test NGF for longevity.

<!-- TOC -->

- [Longevity Test](#longevity-test)
  - [Goals](#goals)
  - [Test Environment](#test-environment)
  - [Steps](#steps)
    - [Start](#start)
    - [Check the Test is Running Correctly](#check-the-test-is-running-correctly)
    - [End](#end)
  - [Analyze](#analyze)
  - [Results](#results)

<!-- TOC -->

## Goals

- Ensure that NGF successfully processes both control plane and data plane transactions over a period of time much
  greater than in our other tests.
- Catch bugs that could only appear over a period of time (like resource leaks).

## Test Environment

- A Kubernetes cluster with 3 nodes on GKE
  - Node: e2-medium (2 vCPU, 4GB memory)
  - Enabled GKE logging.
  - Enabled GKE Cloud monitoring with managed Prometheus service, with enabled:
    - system.
    - kube state - pods, deployments.
- Tester VMs:
  - Configuration:
    - Debian
    - Install packages: tmux, wrk
  - Location - same zone as the Kubernetes cluster.
  - First VM - for HTTP traffic
  - Second VM - for sending HTTPs traffic
- NGF
  - Deployment with 1 replica
  - Exposed via a Service with type LoadBalancer, private IP
  - Gateway, two listeners - HTTP and HTTPs
  - Two apps:
    - Coffee - 3 replicas
    - Tea - 3 replicas
  - Two HTTPRoutes
    - Coffee (HTTP)
    - Tea (HTTPS)

## Steps

### Start

Test duration - 4 days.

1. Create a Kubernetes cluster on GKE.
2. Deploy NGF.
3. Expose NGF via a LoadBalancer Service with `"networking.gke.io/load-balancer-type":"Internal"` annotation to
   allocate an internal load balancer.
4. Apply the manifests which will:
    1. Deploy the coffee and tea backends.
    2. Configure HTTP and HTTPS listeners on the Gateway.
    3. Expose coffee via HTTP listener and tea via HTTPS listener.
    4. Create two CronJobs to re-rollout backends:
        1. Coffee - every minute for an hour every 6 hours
        2. Tea - every minute for an hour every 6 hours, 3 hours apart from coffee.
    5. Configure Prometheus on GKE to pick up NGF metrics.

    ```shell
    kubectl apply -f files
    ```

5. In Tester VMs, update `/etc/hosts` to have an entry with the External IP of the NGF Service (`10.128.0.10` in this
   case):

   ```text
   10.128.0.10 cafe.example.com
   ```

6. In Tester VMs, start a tmux session (this is needed so that even if you disconnect from the VM, any launched command
   will keep running):

   ```shell
   tmux
   ```

7. In First VM, start wrk for 4 days for coffee via HTTP:

   ```shell
   wrk -t2 -c100 -d96h http://cafe.example.com/coffee
   ```

8. In Second VM, start wrk for 4 days for tea via HTTPS:

   ```shell
   wrk -t2 -c100 -d96h https://cafe.example.com/tea
   ```

Notes:

- The updated coffee and tea backends in cafe.yaml include extra configuration for zero time upgrades, so that
  wrk in Tester VMs doesn't get 502 from NGF. Based on https://learnk8s.io/graceful-shutdown

### Check the Test is Running Correctly

Check that you don't see any errors:

1. Traffic is flowing - look at the access logs of NGINX.
2. Check that CronJob can run.

   ```shell
   kubectl create job --from=cronjob/coffee-rollout-mgr coffee-test
   kubectl create job --from=cronjob/tea-rollout-mgr tea-test
   ```

3. Check that GKE exports logs and Prometheus metrics.

In case of errors, double check if you prepared the environment and launched the test correctly.

### End

- Remove CronJobs.

## Analyze

- Traffic
  - Tester VMs (clients)
    - As wrk stop, they will print output upon termination. To connect to the tmux session with wrk,
          run `tmux attach -t 0`
    - Check for errors, latency, RPS
- Logs
  - Check the logs for errors in Google Cloud Operations Logging.
    - NGF
    - NGINX
- Check metrics in Google Cloud Monitoring.
  - NGF
    - CPU usage
      - NGINX
      - NGF
    - Memory usage
      - NGINX
      - NGF
    - NGINX metrics
    - Reloads

## Results

- [1.0.0](results/1.0.0/1.0.0.md)
