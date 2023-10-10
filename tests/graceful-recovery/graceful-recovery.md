# Graceful recovery from restarts

This document describes how we test graceful recovery from restarts on NGF.

<!-- TOC -->
- [Graceful recovery from restarts](#graceful-recovery-from-restarts)
  - [Goal](#goal)
  - [Test Environment](#test-environment)
  - [Steps](#steps)
    - [Setup](#setup)
    - [Run the tests](#run-the-tests)
      - [Restart nginx-gateway container](#restart-nginx-gateway-container)
      - [Restart NGINX container](#restart-nginx-container)
      - [Restart Node with draining](#restart-node-with-draining)
      - [Restart Node without draining](#restart-node-without-draining)
<!-- TOC -->

## Goal

Ensure that NGF can recover gracefully from container failures without any user intervention.

## Test Environment

- A Kubernetes cluster with 3 nodes on GKE
  - Node: e2-medium (2 vCPU, 4GB memory)
- A Kind cluster

## Steps

### Setup

1. Setup GKE Cluster.
2. Clone the repo and change into the nginx-gateway-fabric directory.
3. Check out the latest tag (unless you are installing the edge version from the main branch).
4. Go into `deploy/manifests/nginx-gateway.yaml` and change `runAsNonRoot` from `true` to `false`.
This allows us to insert our ephemeral container as root which enables us to restart the nginx-gateway container.
5. Follow the [installation instructions](https://github.com/nginxinc/nginx-gateway-fabric/blob/main/docs/installation.md)
to deploy NGINX Gateway Fabric using manifests and expose it through a LoadBalancer Service.
6. In a separate terminal track NGF logs by running

    ```console
    kubectl -n nginx-gateway logs -f deploy/nginx-gateway
    ```

7. In a separate terminal track NGINX container logs by running

    ```console
    kubectl -n nginx-gateway logs -f <NGF_POD> -c nginx
    ```

8. Exec into the NGINX container inside of the NGF pod by running

    ```console
    kubectl exec -it -n nginx-gateway <NGF_POD> --container nginx -- sh
    ```

9. In a different terminal, deploy the
[https-termination example](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples/https-termination).
10. Inside the NGINX container, navigate to `/etc/nginx/conf.d` and check `http.conf` and `config-version.config` to see
if the configuration and version were correctly updated.
11. Send traffic through the example application and ensure it is working correctly.

### Run the tests

#### Restart nginx-gateway container

1. Ensure NGF and NGINX container logs are set up and traffic flows through the example application correctly.
2. Insert ephemeral container in NGF Pod.

    ```console
    kubectl debug -it -n nginx-gateway <NGF_POD> --image=busybox:1.28 --target=nginx-gateway
    ```

3. Kill nginx-gateway process through a SIGKILL signal (Process command should start with `/usr/bin/gateway`).

    ```console
    kill -9 <nginx-gateway_PID>
    ```

4. Check for errors in the NGF and NGINX container logs.
5. When the nginx-gateway container is back up, ensure traffic flows through the example application correctly.
6. Open up the NGF and NGINX container logs and check for errors.
7. Inside the NGINX container, check that `http.conf` was not changed and `config-version.conf` had its version set to `2`.
8. Send traffic through the example application and ensure it is working correctly.
9. Check that NGF can still process changes of resources.
   1. Delete the HTTPRoute resources.

       ```console
        kubectl delete -f ../../examples/https-termination/cafe-routes.yaml
       ```

   2. Inside the NGINX container, check that `http.conf` and `config-version.conf` were correctly updated.
   3. Send traffic through the example application using the updated resources and ensure traffic does not flow.
   4. Apply the HTTPRoute resources.

       ```console
       kubectl apply -f ../../examples/https-termination/cafe-routes.yaml
       ```

   5. Inside the NGINX container, check that `http.conf` and `config-version.conf` were correctly updated.
   6. Send traffic through the example application using the updated resources and ensure traffic flows correctly.

#### Restart NGINX container

1. Ensure NGF and NGINX container logs are set up and traffic flows through the example application correctly.
2. If the terminal inside the NGINX container is no longer running, Exec back into the NGINX container.
3. Inside the NGINX container, kill the nginx-master process through a SIGKILL signal
(Process command should start with `nginx: master process`).

    ```console
    kill -9 <nginx-master_PID>
    ```

4. When NGINX container is back up, ensure traffic flows through the example application correctly.
5. Open up the NGINX container logs and check for errors.
6. Exec back into the NGINX container and check that `http.conf` and `config-version.conf` were not changed.
7. Check that NGF can still process changes of resources.
    1. Delete the HTTPRoute resources.

        ```console
         kubectl delete -f ../../examples/https-termination/cafe-routes.yaml
        ```

    2. Inside the NGINX container, check that `http.conf` and `config-version.conf` were correctly updated.
    3. Send traffic through the example application using the updated resources and ensure traffic does not flow.
    4. Apply the HTTPRoute resources.

        ```console
        kubectl apply -f ../../examples/https-termination/cafe-routes.yaml
        ```

    5. Inside the NGINX container, check that `http.conf` and `config-version.conf` were correctly updated.
    6. Send traffic through the example application using the updated resources and ensure traffic flows correctly.

#### Restart Node with draining

1. Switch over to a one-Node Kind cluster. Can run `make create-kind-cluster` from main directory.
2. Run steps 4-11 of the [Setup](#setup) section above using
[this guide](https://github.com/nginxinc/nginx-gateway-fabric/blob/main/docs/running-on-kind.md) for running on Kind.
3. Ensure NGF and NGINX container logs are set up and traffic flows through the example application correctly.
4. Drain the Node of its resources.

    ```console
    kubectl drain kind-control-plane --ignore-daemonsets --delete-local-data
    ```

5. Delete the Node.

    ```console
    kubectl delete node kind-control-plane
    ```

6. Restart the Docker container.

    ```console
    docker restart kind-control-plane
    ```

7. Open up both NGF and NGINX container logs and check for errors.
8. Exec back into the NGINX container and check that `http.conf` and `config-version.conf` were not changed.
9. Send traffic through the example application and ensure it is working correctly.
10. Check that NGF can still process changes of resources.
    1. Delete the HTTPRoute resources.

        ```console
         kubectl delete -f ../../examples/https-termination/cafe-routes.yaml
        ```

    2. Inside the NGINX container, check that `http.conf` and `config-version.conf` were correctly updated.
    3. Send traffic through the example application using the updated resources and ensure traffic does not flow.
    4. Apply the HTTPRoute resources.

        ```console
        kubectl apply -f ../../examples/https-termination/cafe-routes.yaml
        ```

    5. Inside the NGINX container, check that `http.conf` and `config-version.conf` were correctly updated.
    6. Send traffic through the example application using the updated resources and ensure traffic flows correctly.

#### Restart Node without draining

1. Repeat the above test but remove steps 4-5 which include draining and deleting the Node.
