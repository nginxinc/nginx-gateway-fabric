# Development Quickstart

This guide will assist you in setting up your development environment for NGINX Kubernetes Gateway, covering the steps
to build, install, and execute tasks necessary for submitting PRs. By following this guide, you'll have a fully prepared
development environment that allows you to contribute to the project effectively.

## Setup Your Development Environment

Follow these steps to set up your development environment.

1. Install:
    - [Go](https://golang.org/doc/install)
    - [Docker](https://docs.docker.com/get-docker/) v18.09+
    - [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
    - [git](https://git-scm.com/)
    - [GNU Make](https://www.gnu.org/software/software.html)
    - [fieldalignment](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/fieldalignment):

      ```shell
      go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
      ```

2. [Fork the project repository](https://github.com/nginxinc/nginx-kubernetes-gateway/fork)
3. Clone your repository, and install the project dependencies:

   ```shell
   git clone https://github.com/<YOUR-USERNAME>/nginx-kubernetes-gateway.git
   cd nginx-kubernetes-gateway
   make deps
   ```

## Build the Binary and Image

### Build the Binary

To build the binary, run the make build command from the project's root directory:

```shell
make build
```

This command will build the binary and output it to the `/build/.out` directory.

### Build the Image

To build an NGINX Kubernetes Gateway container image from source run the following make command:

```shell
make TAG=$(whoami) container
```

This will build the docker image and tag it with your user ID, e.g. `docker.io/library/nginx-kubernetes-gateway:user`

## Deploy on Kind

1. Create a `kind` cluster:

   ```shell
   make create-kind-cluster
   ```

2. Build the NKG image and load it onto your `kind` cluster:

   ```shell
   make TAG=$(whoami) container
   kind load docker-image docker.io/library/nginx-kubernetes-gateway:$(whoami)
   ```

3. Modify the image name and image pull policy for the `nginx-gateway` container in the
   NKG [deployment manifest](../../deploy/manifests/nginx-gateway.yaml). Set the image name to the image you built in
   the previous step and the image pull policy to `Never`. Once the changes are made, follow
   the [installation instructions](../installation.md) to install NKG on your `kind` cluster.

   Alternatively, you can update the image name and pull policy by using the following command when applying
   `nginx-gateway.yaml`:

   ```shell 
   cat deploy/manifests/nginx-gateway.yaml | sed "s|image: ghcr.io/nginxinc/nginx-kubernetes-gateway.*|image: docker.io/library/nginx-kubernetes-gateway:<YOUR-TAG>|" | sed "s|imagePullPolicy: Always|imagePullPolicy: Never|" | kubectl apply -f -
   ```

### Run Examples

To make sure NKG is running properly, try out the [examples](../../examples).

## Run the Unit Tests

To run all the unit tests, run the make unit-test command from the project's root directory:

```shell
make unit-test
```

For more details on testing, see the [testing](testing.md) documentation.

## Run the Linter

To lint the code, run the following make command from the project's root directory:

```shell
make lint
```

> **Note**
> fieldalignment errors can be fixed by running: `fieldalignment -fix <path-to-package>`
