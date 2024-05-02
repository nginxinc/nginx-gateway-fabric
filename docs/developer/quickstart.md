# Development Quickstart

This guide will assist you in setting up your development environment for NGINX Gateway Fabric, covering the steps
to build, install, and execute tasks necessary for submitting pull requests. By following this guide, you'll have a
fully prepared development environment that allows you to contribute to the project effectively.

## Setup Your Development Environment

Follow these steps to set up your development environment.

1. Install:

   - [Go](https://golang.org/doc/install) v1.21.0+
   - [Docker](https://docs.docker.com/get-docker/) v18.09+
   - [Kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
   - [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
   - [Helm](https://helm.sh/docs/intro/quickstart/#install-helm)
   - [git](https://git-scm.com/)
   - [GNU Make](https://www.gnu.org/software/software.html)
   - [yq](https://github.com/mikefarah/yq/#install)
   - [fieldalignment](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/fieldalignment):

   ```shell
   go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
   ```

   - [pre-commit](https://pre-commit.com/#install):

   ```shell
   brew install pre-commit
   ```

2. [Fork the project repository](https://github.com/nginxinc/nginx-gateway-fabric/fork)
3. Clone your repository with ssh, and install the project dependencies:

   ```shell
   git clone git@github.com:<YOUR-USERNAME>/nginx-gateway-fabric.git
   cd nginx-gateway-fabric
   ```

   and then run

   ```shell
   pre-commit install
   ```

   in the project root directory to install the git hooks.

   ```makefile
   make deps
   ```

4. Finally, add the original project repository as the remote upstream:

   ```shell
   git remote add upstream git@github.com:nginxinc/nginx-gateway-fabric.git
   ```


## Build the Binary and Images

### Setting GOARCH

The [Makefile](/Makefile) uses the GOARCH variable to build the binary and container images. The default value of GOARCH is `amd64`.

If you are deploying NGINX Gateway Fabric on a kind cluster, and the architecture of your machine is not `amd64`, you will want to set the GOARCH variable to the architecture of your local machine. You can find the value of GOARCH by running `go env`. Export the GOARCH variable in your `~/.zshrc` or `~/.bashrc`.

```shell
echo "export GOARCH=< Your architecture (e.g. arm64 or amd64) >" >> ~/.bashrc
source ~/.bashrc
```

or for zsh:

```shell
echo "export GOARCH=< Your architecture (e.g. arm64 or amd64) >" >> ~/.zshrc
source ~/.zshrc
```

### Build the Binary

To build the binary, run the make build command from the project's root directory:

```makefile
make GOARCH=$GOARCH build
```

This command will build the binary and output it to the `/build/.out` directory.

### Build the Images

To build the NGINX Gateway Fabric and NGINX container images from source run the following make command:

```makefile
make GOARCH=$GOARCH TAG=$(whoami) build-images
```

This will build the docker images `nginx-gateway-fabric:<your-user>` and `nginx-gateway-fabric/nginx:<your-user>`.

### Build the Images with NGINX Plus

> Note: You will need a valid NGINX Plus license certificate and key named `nginx-repo.crt` and `nginx-repo.key` in the
> root of this repo to build the NGINX Plus image.

To build the NGINX Gateway Fabric and NGINX Plus container images from source run the following make command:

```makefile
make TAG=$(whoami) build-images-with-plus
```

This will build the docker images `nginx-gateway-fabric:<your-user>` and `nginx-gateway-fabric/nginx-plus:<your-user>`.

## Deploy on Kind

1. Create a `kind` cluster:

   ```makefile
   make create-kind-cluster
   ```

2. Load the previously built images onto your `kind` cluster:

   ```shell
   kind load docker-image nginx-gateway-fabric:$(whoami) nginx-gateway-fabric/nginx:$(whoami)
   ```

   or

   ```shell
   kind load docker-image nginx-gateway-fabric:$(whoami) nginx-gateway-fabric/nginx-plus:$(whoami)
   ```

3. Install Gateway API CRDs:

   ```shell
   kubectl kustomize config/crd/gateway-api/standard | kubectl apply -f -
   ```

   If you're implementing experimental Gateway API features, install Gateway API CRDs from the experimental channel:

   ```shell
   kubectl kustomize config/crd/gateway-api/experimental | kubectl apply -f -
   ```

4. Install NGF using your custom image and expose NGF with a NodePort Service:

   - To install with Helm (where your release name is `my-release`):

     ```shell
     helm install my-release ./charts/nginx-gateway-fabric --create-namespace --wait --set service.type=NodePort --set nginxGateway.image.repository=nginx-gateway-fabric --set nginxGateway.image.tag=$(whoami) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=nginx-gateway-fabric/nginx --set nginx.image.tag=$(whoami) --set nginx.image.pullPolicy=Never -n nginx-gateway
     ```

   - To install NGINX Plus with Helm (where your release name is `my-release`):

     ```shell
     helm install my-release ./charts/nginx-gateway-fabric --create-namespace --wait --set service.type=NodePort --set nginxGateway.image.repository=nginx-gateway-fabric --set nginxGateway.image.tag=$(whoami) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=nginx-gateway-fabric/nginx-plus --set nginx.image.tag=$(whoami) --set nginx.image.pullPolicy=Never --set nginx.plus=true -n nginx-gateway
     ```

   > For more information on Helm configuration options see the Helm [README](../../charts/nginx-gateway-fabric/README.md).

   - To install with manifests:

     ```shell
     make generate-manifests HELM_TEMPLATE_COMMON_ARGS="--set nginxGateway.image.repository=nginx-gateway-fabric --set nginxGateway.image.tag=$(whoami) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=nginx-gateway-fabric/nginx --set nginx.image.tag=$(whoami) --set nginx.image.pullPolicy=Never"
     kubectl apply -f deploy/crds.yaml
     kubectl apply -f deploy/manifests/nginx-gateway.yaml
     kubectl apply -f deploy/manifests/service/nodeport.yaml
     ```

   - To install NGINX Plus with manifests:

     ```shell
     make generate-manifests HELM_TEMPLATE_COMMON_ARGS="--set nginxGateway.image.repository=nginx-gateway-fabric --set nginxGateway.image.tag=$(whoami) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=nginx-gateway-fabric/nginx-plus --set nginx.image.tag=$(whoami) --set nginx.image.pullPolicy=Never --set nginx.plus=true"
     kubectl apply -f deploy/crds.yaml
     kubectl apply -f deploy/manifests/nginx-gateway.yaml
     kubectl apply -f deploy/manifests/service/nodeport.yaml
     ```

   - To install with experimental manifests:

     ```shell
     make generate-manifests HELM_TEMPLATE_COMMON_ARGS="--set nginxGateway.image.repository=nginx-gateway-fabric --set nginxGateway.image.tag=$(whoami) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=nginx-gateway-fabric/nginx --set nginx.image.tag=$(whoami) --set nginx.image.pullPolicy=Never"
     kubectl apply -f deploy/crds.yaml
     kubectl apply -f deploy/manifests/nginx-gateway-experimental.yaml
     kubectl apply -f deploy/manifests/service/nodeport.yaml
     ```

### Run Examples

To make sure NGF is running properly, try out the [examples](/examples).

## Run the Unit Tests

To run all the unit tests, run the make unit-test command from the project's root directory:

```makefile
make unit-test
```

For more details on testing, see the [testing](/docs/developer/testing.md) documentation.

## Gateway API Conformance Testing

To run Gateway API conformance tests, please follow the instructions on [this](/tests/README.md#conformance-testing) page.

## Run the Linter

To lint the code, run the following make command from the project's root directory:

```makefile
make lint
```

> **Note**
> fieldalignment errors can be fixed by running: `fieldalignment -fix <path-to-package>`

## Run the Helm Linter

Run the following make command from the project's root directory to lint the Helm Chart code:

```shell
make lint-helm
```

## Run Code Generation

To ensure all the generated code is up to date, run the following make command from the project's root directory:

```shell
make generate
```

That command also will generate the avro scheme (`.avdl`) for product telemetry data points.

## Update Generated Manifests

To update the generated manifests, run the following make command from the project's root directory:

```shell
make generate-manifests
```
