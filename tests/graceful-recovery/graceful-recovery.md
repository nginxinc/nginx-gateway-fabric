# Graceful recovery from restarts

## Goal
Ensure that NGF can recover gracefully from container failures without any user intervention.

## Cluster Details

- GKE 1.27.3-gke.100
- us-central1-c
- Machine type of node is e2-medium
- 3 nodes

## Setup

1. Setup GKE Cluster.
2. Clone the repo and change into the nginx-gateway-fabric directory.
3. Check out the latest tag (unless you are installing the edge version from the main branch).
4. Go into `deploy/manifests/nginx-gateway.yaml` and change `runAsNonRoot` from `true` to `false`.
5. Follow the [installation instructions](https://github.com/nginxinc/nginx-gateway-fabric/blob/main/docs/installation.md)
to deploy NGINX Gateway Fabric.
6. In a separate terminal track NGF logs by running `kubectl -n nginx-gateway logs -f deploy/nginx-gateway`
7. In a separate terminal track NGINX container logs by running
`kubectl -n nginx-gateway logs -f <NGF_POD> -c nginx`
8. Exec into the NGINX container inside of the NGF pod by running
`kubectl exec -it -n nginx-gateway <NGF_POD> --container nginx -- bin/sh`
9. Inside the NGINX container, navigate to `/etc/nginx/conf.d` and ensure that
`http.conf` and `config-version.conf` look correct.
10. In a different terminal, deploy the
[https-termination example](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples/https-termination).
11. Inside the NGINX container, check `http.conf` and `config-version.config` to see
if the configuration and version were correctly updated.
12. Send traffic through the example application and ensure it is working correctly.

## Tests

### Restart nginx-gateway container

1. Ensure NGF and NGINX container logs are set up and traffic flows through the example application correctly.
2. Insert ephemeral container in NGF Pod and kill the nginx-gateway process.
    1. `kubectl debug -it -n nginx-gateway <NGF_POD> --image=busybox:1.28 --target=nginx-gateway`
    2. run `ps -A`
    3. run `kill <nginx-gateway_PID>` (Command should start with `/usr/bin/gateway`)
3. Check for errors in the NGF and NGINX container logs.
4. When the nginx-gateway container is back up, ensure traffic flows through the example application correctly.
5. Open up the NGF and NGINX container logs and check for errors.
6. Inside the NGINX container, check that `http.conf` was not changed and `config-version.conf` had its version set to `2`.
7. Send traffic through the example application and ensure it is working correctly.
8. Check that NGF can still update statuses of resources.
   1. Delete the HTTPRoute resources by running `kubectl delete -f cafe-routes.yaml` in `/examples/https-termination`
   2. Inside the terminal which is inside the NGINX container, check that `http.conf` and
   `config-version.conf` were correctly updated.
   3. Send traffic through the example application using the updated resources and ensure traffic does not flow.
   4. Apply the HTTPRoute resources by running `kubectl apply -f cafe-routes.yaml` in `/examples/https-termination`
   5. Inside the terminal which is inside the NGINX container, check that `http.conf` and
   `config-version.conf` were correctly updated.
   6. Send traffic through the example application using the updated resources and ensure traffic flows correctly.

### Restart NGINX container

1. Ensure NGF and NGINX container logs are set up and traffic flows through the example application correctly.
2. Insert ephemeral container in NGF Pod and kill the nginx-master process.
   1. If there isn't already an ephemeral container inserted, run:
   `kubectl debug -it -n nginx-gateway <NGF_POD> --image=busybox:1.28 --target=nginx-gateway`
   2. run `ps -A`
   3. run `kill <nginx-master_PID>` (Command should start with `nginx: master process`)
3. When NGINX container is back up, ensure traffic flows through the example application correctly.
4. Open up the NGINX container logs and check for errors.
5. Exec back into the NGINX container and check that `http.conf` and `config-version.conf` were not changed.

### Restart Node with draining

1. Switch over to a one-node Kind cluster. Can run `make create-kind-cluster` from main directory.
2. Run steps 4-12 of the Setup section above using [this guide]
(https://github.com/nginxinc/nginx-gateway-fabric/blob/main/docs/running-on-kind.md) for running on Kind.
3. Ensure NGF and NGINX container logs are set up and traffic flows through the example application correctly.
4. Drain the node of its resources by running `kubectl drain kind-control-plane --ignore-daemonsets --delete-local-data`
5. Delete the node by running `kubectl delete node kind-control-plane`
6. Restart the docker container by running `docker restart kind-control-plane`
7. Open up both NGF and NGINX container logs and check for errors.
8. Exec back into the NGINX container and check that `http.conf` and `config-version.conf` were not changed.
9. Send traffic through the example application and ensure it is working correctly.
10. Check that NGF can still update statuses of resources.
    1. Delete the HTTPRoute resources by running `kubectl delete -f cafe-routes.yaml` in `/examples/https-termination`
    2. Inside the terminal which is inside the NGINX container, check that `http.conf` and
    `config-version.conf` were correctly updated.
    3. Send traffic through the example application using the updated resources and ensure traffic does not flow.
    4. Apply the HTTPRoute resources by running `kubectl apply -f cafe-routes.yaml` in `/examples/https-termination`
    5. Inside the terminal which is inside the NGINX container, check that `http.conf` and
    `config-version.conf` were correctly updated.
    6. Send traffic through the example application using the updated resources and ensure traffic flows correctly.

### Restart Node without draining

1. Repeat the above test but remove steps 4-5 which include draining and deleting the node.
