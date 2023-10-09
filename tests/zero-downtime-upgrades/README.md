# Zero-Downtime Upgrades

This document describes a test plan for testing zero-downtime upgrades of NGF.

*Zero-downtime upgrades* means that during an NGF upgrade clients don't experience any
interruptions to the traffic they send to applications exposed via NGF.

## Goals

- Ensure that upgrading NFG doesn't lead to any loss of traffic flowing through the data plane.
- Ensure that after an upgrade, NGF can process changes to resources.

## Non-Goals

During an upgrade, Kubernetes will shut down existing NGF pods by sending a SIGTERM. If the pod doesn't terminate in 30
seconds (the default period) , Kubernetes will send a SIGKILL.

When proxying Websocket or any long-lived connections, NGINX will not terminate until
that connection is closed by either the client or the backend. This means that unless all those connections are closed
by clients/backends before or during an upgrade (which is highly unlikely), NGINX will not terminate, which means
Kubernetes will kill NGINX. As a result, the clients will see the connections abruptly closed and thus experience
downtime.

As a result, we *will not* use any long-live connections in this test, because NGF cannot support zero-downtime upgrades
in this case.

## Test Environment

- A Kubernetes cluster with 3 nodes on GKE
    - Node: e2-medium (2 vCPU, 4GB memory)
    - Enabled GKE logging.
- Tester VMs:
    - Configuration:
        - Debian
        - Install packages: wrk
    - Location - same zone as the Kubernetes cluster.
    - First VM - for HTTP traffic
    - Second VM - for sending HTTPs traffic
- NGF
    - Deployment with 2 replicas
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

- Create a cluster
- Deploy NGF, previous latest stable version with 2 replicas.
- Expose NGF via Service Load Balancer, internal (only accessible within the Google Cloud region)
- Deploy backend apps
- Configure Gateway
- Expose apps via HTTPRoutes
- Check statuses of the Gateway and HTTPRoutes
- Start sending traffic using wrk from tester VMs: HTTP and HTTPs.
- Check that NGINX access logs look good (all responses 200)
- Check that there are no errors in NGINX errors logs and NGF error logs.

### Upgrade

- Upgrade to the new version build from main branch. For example, `kubectl apply -f` the latest manifests.
- Check that the new pods are running and the old one are removed.

## After Upgrade

- Update the Gateway resource - add one listener -- and make sure NGF processed the update (it updates the status of
  the Gateway resource accordingly).

### Analyze

- Stop wrk and save the output
- Check old pods logs (in Google Monitoring)
    - NGINX Access logs - no errors
    - NGINX Error logs - no errors
    - NGF logs - no errors
- Check New pods (in Google Monitoring)
    - NGINX Access logs - no errors
    - NGINX Error logs - no errors
    - NGF logs - no errors
