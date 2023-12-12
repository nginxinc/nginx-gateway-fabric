# System Testing

The tests in this directory are meant to be run on a live Kubernetes environment to verify a real system. These
are similar to the existing [conformance tests](../conformance/README.md), but will verify things such as:

- NGF-specific functionality
- Non-Functional requirements testing (such as performance, scale, etc.)

When running, the tests create a port-forward from your NGF Pod to localhost using a port chosen by the
test framework. Traffic is sent over this port.

Directory structure is as follows:

- `framework`: contains utility functions for running the tests
- `suite`: contains the test files

**Note**: Existing NFR tests will be migrated into this testing `suite` and results stored in a `results` directory.

## Prerequisites

- Kubernetes cluster.
- Docker.
- Golang.

If running the tests on a VM (`make run-tests-on-vm`):

- A GKE cluster allowing network access from VMs in the same region
- Access to GCP Service Account with Kubernetes admin permissions
- SSH access to GCP VMs in the same region as your cluster

**Note**: all commands in steps below are executed from the `tests` directory

```shell
make
```

```text
build-images                   Build NGF and NGINX images
cleanup-vm                     Delete the test GCP VM
create-kind-cluster            Create a kind cluster
delete-kind-cluster            Delete kind cluster
help                           Display this help
load-images                    Load NGF and NGINX images on configured kind cluster
run-tests-on-vm                Run the tests on a GCP VM
test                           Run the system tests against your default k8s cluster
```

**Note:** The following variables are configurable when running the below `make` commands:

| Variable            | Default                    | Description                                                    |
| ------------------- | -------------------------- | -------------------------------------------------------------- |
| TAG                 | edge                       | tag for the locally built NGF images                           |
| PREFIX              | nginx-gateway-fabric       | prefix for the locally built NGF image                         |
| NGINX_PREFIX        | nginx-gateway-fabric/nginx | prefix for the locally built NGINX image                       |
| PULL_POLICY         | Never                      | NGF image pull policy                                          |
| GW_API_VERSION      | 1.0.0                      | version of Gateway API resources to install                    |
| K8S_VERSION         | latest                     | version of k8s that the tests are run on                       |
| GW_SERVICE_TYPE     | NodePort                   | Type of Service that should be created                         |
| GW_SVC_GKE_INTERNAL | false                      | Specifies if the LoadBalancer should be a GKE internal service |

## Step 1 - Create a Kubernetes cluster

This can be done in a cloud provider of choice, or locally using `kind`:

```makefile
make create-kind-cluster
```

> Note: The default kind cluster deployed is the latest available version. You can specify a different version by
> defining the kind image to use through the KIND_IMAGE variable, e.g.

```makefile
make create-kind-cluster KIND_IMAGE=kindest/node:v1.27.3
```

## Step 2 - Build and Load Images

Loading the images only applies to a `kind` cluster. If using a cloud provider, you will need to tag and push
your images to a registry that is accessible from that cloud provider.

```makefile
make build-images load-images TAG=$(whoami)
```

## Step 3 - Run the tests

### 3a - Run the tests locally

```makefile
make test TAG=$(whoami)
```

### 3b - Run the tests on a GKE cluster from a GCP VM

This step only applies if you would like to run the tests from a GCP based VM. The VM should be created in the same
zone as your GKE cluster, and requires a service account that has Kubernetes admin permissions. Additionally, you need
ssh access to the VM and the VM needs to have network access to the Kubernetes control node.

Before running the below `make` command, populate the required env vars in `scripts/vars.env`.

```makefile
make run-tests-on-vm
```

### Common test amendments

To run a specific test, you can "focus" it by adding the `F` prefix to the name. For example:

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

## Step 4 - Cleanup

1. Delete kind cluster, if required

    ```makefile
    make delete-kind-cluster
    ```

2. Delete the cloud VM, if required

    ```makefile
    make cleanup-vm
    ```
