# Graceful recovery from restarts

This document describes how we test graceful recovery from restarts on NGF.

<!-- TOC -->
- [Graceful recovery from restarts](#graceful-recovery-from-restarts)
  - [Goal](#goal)
  - [Test Environment](#test-environment)
  - [Steps](#steps)
    - [Setup](#setup)
    - [Run the tests](#run-the-tests)
      - [Restart Node with draining](#restart-node-with-draining)
      - [Restart Node without draining](#restart-node-without-draining)
<!-- TOC -->

## Goal

Ensure that NGF can recover gracefully from container failures without any user intervention.

## Test Environment

- A Kind cluster

## Steps

### Setup

1. Deploy a one-Node Kind cluster. Can run `make create-kind-cluster` from main directory.

2. Go into `deploy/manifests/nginx-gateway.yaml` and change the following:

   - `runAsNonRoot` from `true` to `false`: this allows us to insert our ephemeral container as root which enables us to restart the nginx-gateway container.
   - Add the `--product-telemetry-disable` argument to the nginx-gateway container args.

3. Follow [this guide](https://docs.nginx.com/nginx-gateway-fabric/installation/running-on-kind/) to deploy NGINX Gateway Fabric using manifests and expose it through a NodePort Service.

4. In a separate terminal track NGF logs.

    ```console
    kubectl -n nginx-gateway logs -f deploy/nginx-gateway -c nginx-gateway
    ```

5. In a separate terminal track NGINX container logs.

    ```console
    kubectl -n nginx-gateway logs -f deploy/nginx-gateway -c nginx
    ```

6. In a separate terminal Exec into the NGINX container inside the NGF pod.

    ```console
    kubectl exec -it -n nginx-gateway $(kubectl get pods -n nginx-gateway | sed -n '2s/^\([^[:space:]]*\).*$/\1/p') --container nginx -- sh
    ```

7. In a different terminal, deploy the
[https-termination example](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples/https-termination).
8. Send traffic through the example application and ensure it is working correctly.

### Run the tests

#### Restart Node with draining

1. Drain the Node of its resources.

    ```console
    kubectl drain kind-control-plane --ignore-daemonsets --delete-local-data
    ```

2. Delete the Node.

    ```console
    kubectl delete node kind-control-plane
    ```

3. Restart the Docker container.

    ```console
    docker restart kind-control-plane
    ```

4. Check the logs of the old and new NGF and NGINX containers for errors.
5. Send traffic through the example application and ensure it is working correctly.
6. Check that NGF can still process changes of resources.
    1. Delete the HTTPRoute resources.

        ```console
         kubectl delete -f ../../examples/https-termination/cafe-routes.yaml
        ```

    2. Send traffic through the example application using the updated resources and ensure traffic does not flow.
    3. Apply the HTTPRoute resources.

        ```console
        kubectl apply -f ../../examples/https-termination/cafe-routes.yaml
        ```

    4. Send traffic through the example application using the updated resources and ensure traffic flows correctly.

#### Restart Node without draining

1. Repeat the above test but remove steps 1-2 which include draining and deleting the Node.
