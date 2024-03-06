# System Testing

The tests in this directory are meant to be run on a live Kubernetes environment to verify a real system. These
are similar to the existing [conformance tests](../conformance/README.md), but will verify things such as:

- NGF-specific functionality
- Non-Functional requirements testing (such as performance, scale, etc.)

When running locally, the tests create a port-forward from your NGF Pod to localhost using a port chosen by the
test framework. Traffic is sent over this port. If running on a GCP VM targeting a GKE cluster, the tests will create an
internal LoadBalancer service which will receive the test traffic.

Directory structure is as follows:

- `framework`: contains utility functions for running the tests
- `suite`: contains the test files
- `results`: contains the results files

**Note**: Existing NFR tests will be migrated into this testing `suite` and results stored in the `results` directory.

## Prerequisites

- Kubernetes cluster.
- Docker.
- Golang.

If running the tests on a VM (`make create-vm-and-run-tests` or `make run-tests-on-vm`):

- The [gcloud CLI](https://cloud.google.com/sdk/docs/install)
- A GKE cluster (if `master-authorized-networks` is enabled, please set `ADD_VM_IP_AUTH_NETWORKS=true` in your vars.env file)
- Access to GCP Service Account with Kubernetes admin permissions

**Note**: all commands in steps below are executed from the `tests` directory

```shell
make
```

```text
add-local-ip-to-cluster        Add local IP to the GKE cluster master-authorized-networks
build-images-with-plus         Build NGF and NGINX Plus images
build-images                   Build NGF and NGINX images
cleanup-gcp                    Cleanup all GCP resources
cleanup-router                 Delete the GKE router
cleanup-vm                     Delete the test GCP VM and delete the firewall rule
create-and-setup-vm            Create and setup a GCP VM for tests
create-gke-cluster             Create a GKE cluster
create-gke-router              Create a GKE router to allow egress traffic from private nodes (allows for external image pulls)
create-kind-cluster            Create a kind cluster
delete-gke-cluster             Delete the GKE cluster
delete-kind-cluster            Delete kind cluster
help                           Display this help
load-images-with-plus          Load NGF and NGINX Plus images on configured kind cluster
load-images                    Load NGF and NGINX images on configured kind cluster
run-tests-on-vm                Run the tests on a GCP VM
setup-gcp-and-run-tests        Create and setup a GKE router and GCP VM for tests and run the tests
test                           Run the system tests against your default k8s cluster
```

**Note:** The following variables are configurable when running the below `make` commands:

| Variable            | Default                         | Description                                                    |
| ------------------- | ------------------------------- | -------------------------------------------------------------- |
| TAG                 | edge                            | tag for the locally built NGF images                           |
| PREFIX              | nginx-gateway-fabric            | prefix for the locally built NGF image                         |
| NGINX_PREFIX        | nginx-gateway-fabric/nginx      | prefix for the locally built NGINX image                       |
| NGINX_PLUS_PREFIX   | nginx-gateway-fabric/nginx-plus | prefix for the locally built NGINX Plus image                  |
| PLUS_ENABLED        | false                           | Flag to indicate if NGINX Plus should be enabled               |
| PULL_POLICY         | Never                           | NGF image pull policy                                          |
| GW_API_VERSION      | 1.0.0                           | version of Gateway API resources to install                    |
| K8S_VERSION         | latest                          | version of k8s that the tests are run on                       |
| GW_SERVICE_TYPE     | NodePort                        | type of Service that should be created                         |
| GW_SVC_GKE_INTERNAL | false                           | specifies if the LoadBalancer should be a GKE internal service |
| GINKGO_LABEL        | ""                              | name of the ginkgo label that will filter the tests to run     |
| GINKGO_FLAGS        | ""                              | other ginkgo flags to pass to the go test command              |

## Step 1 - Create a Kubernetes cluster

This can be done in a cloud provider of choice, or locally using `kind`.

To create a local `kind` cluster:

```makefile
make create-kind-cluster
```

> Note: The default kind cluster deployed is the latest available version. You can specify a different version by
> defining the kind image to use through the KIND_IMAGE variable, e.g.

```makefile
make create-kind-cluster KIND_IMAGE=kindest/node:v1.27.3
```

To create a GKE cluster:

Before running the below `make` command, copy the `scripts/vars.env-example` file to `scripts/vars.env` and populate the
required env vars. `GKE_SVC_ACCOUNT` needs to be the name of a service account that has Kubernetes admin permissions,
and `GKE_NODES_SERVICE_ACCOUNT` needs to be the name of a service account that has `Artifact Registry Reader`,
`Kubernetes Engine Node Service Account` and `Monitoring Viewer` permissions.

```makefile
make create-gke-cluster
```

> Note: The GKE cluster is created with `master-authorized-networks`, meaning only IPs from explicitly allowed CIDR ranges
> will be able to access the cluster. The script will automatically add your current IP to the authorized list, but if
> your IP changes, you can add your new local IP to the `master-authorized-networks` of the cluster by running the
> following:

```makefile
make add-local-ip-to-cluster
```

## Step 2 - Build and Load Images

Loading the images only applies to a `kind` cluster. If using a cloud provider, you will need to tag and push
your images to a registry that is accessible from that cloud provider.

```makefile
make build-images load-images TAG=$(whoami)
```

Or, to build NGF with NGINX Plus enabled (NGINX Plus cert and key must exist in the root of the repo):

```makefile
make build-images-with-plus load-images-with-plus TAG=$(whoami)
```

## Step 3 - Run the tests

### 3a - Run the tests locally

```makefile
make test TAG=$(whoami)
```

Or, to run the tests with NGINX Plus enabled:

```makefile
make test TAG=$(whoami) PLUS_ENABLED=true
```

### 3b - Run the tests on a GKE cluster from a GCP VM

This step only applies if you would like to run the tests on a GKE cluster from a GCP based VM.

Before running the below `make` command, copy the `scripts/vars.env-example` file to `scripts/vars.env` and populate the
required env vars. `GKE_SVC_ACCOUNT` needs to be the name of a service account that has Kubernetes admin permissions.

In order to run the tests in GCP, you need a few things:

- GKE router to allow egress traffic (used by upgrade tests for pulling images from Github)
  - this assumes that your GKE cluster is using private nodes. If using public nodes, you don't need this.
- GCP VM and firewall rule to send ingress traffic to GKE

To set up the GCP environment with the router and VM and then run the tests, run the following command:

```makefile
make setup-gcp-and-run-tests
```

If you just need a VM and no router (this will not run the tests):

```makefile
make create-and-setup-vm
```

To use an existing VM to run the tests, run the following

```makefile
make run-tests-on-vm
```

### Common test amendments

To run all tests with the label "performance", use the GINKGO_LABEL variable:

```makefile
make test TAG=$(whoami) GINKGO_LABEL=performance
```

or to pass a specific flag, e.g. run a specific test, use the GINKGO_FLAGS variable:

```makefile
make test TAG=$(whoami) GINKGO_FLAGS='-ginkgo.focus "writes the system info to a results file"'
```

If you are running the tests in GCP, add your required label/ flags to `scripts/var.env`.

You can also modify the tests code for a similar outcome. To run a specific test, you can "focus" it by adding the `F`
prefix to the name. For example:

```go
It("runs some test", func(){
    ...
})
```

becomes:

```go
FIt("runs some test", func(){
    ...
})
```

This can also be done at higher levels like `Context`.

To disable a specific test, add the `X` prefix to it, similar to the previous example:

```go
It("runs some test", func(){
    ...
})
```

becomes:

```go
XIt("runs some test", func(){
    ...
})
```

For more information of filtering specs, see [the docs here](https://onsi.github.io/ginkgo/#filtering-specs).

## Step 4 - Cleanup

1. Delete kind cluster, if required

    ```makefile
    make delete-kind-cluster
    ```

2. Delete the GCP components (GKE cluster, GKE router, VM, and firewall rule), if required

    ```makefile
    make cleanup-gcp
    ```

    or

    ```makefile
    make cleanup-router
    ```

    ```makefile
    make cleanup-vm
    ```

    ```makefile
    make delete-gke-cluster
    ```
