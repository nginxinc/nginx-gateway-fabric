# Running [Gateway Conformance Tests](https://gateway-api.sigs.k8s.io/concepts/conformance/#3-conformance-tests) in kind

## Prerequisites

- [kind](https://kind.sigs.k8s.io/).
- Docker.
- Golang.
- [yq](https://github.com/mikefarah/yq/#install)

**Note**: all commands in steps below are executed from the ```conformance``` directory

List available commands:

```shell
make
```

```text
build-images                   Build NGF and nginx images
build-test-runner-image        Build conformance test runner image
cleanup-conformance-tests      Clean up conformance tests fixtures
create-kind-cluster            Create a kind cluster
delete-kind-cluster            Delete kind cluster
deploy-updated-provisioner     Update provisioner manifest and deploy to the configured kind cluster
help                           Display this help
install-ngf-edge               Install NGF with provisioner from edge on configured kind cluster
install-ngf-local-build        Install NGF from local build with provisioner on configured kind cluster
install-ngf-local-no-build     Install NGF from local build with provisioner on configured kind cluster but do not build the NGF image
load-images                    Load NGF and NGINX images on configured kind cluster
prepare-ngf-dependencies       Install NGF dependencies on configured kind cluster
reset-go-modules               Reset the go modules changes
run-conformance-tests          Run conformance tests
undo-manifests-update          Undo the changes in the manifest files
uninstall-ngf                  Uninstall NGF on configured kind cluster and undo manifest changes
update-go-modules              Update the gateway-api go modules to latest main version
update-ngf-manifest            Update the NGF deployment manifest image names and imagePullPolicies
```

**Note:** The following variables are configurable when running the below `make` commands:

| Variable             | Default                                                                                                       | Description                                                                                                               |
|----------------------|---------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------|
| TAG                  | latest                                                                                                        | The tag for the conformance test image                                                                                    |
| PREFIX               | conformance-test-runner                                                                                       | The prefix for the conformance test image                                                                                 |
| NGF_TAG              | edge                                                                                                          | The tag for the locally built NGF image                                                                                   |
| NGF_PREFIX           | nginx-gateway-fabric                                                                                          | The prefix for the locally built NGF image                                                                                |
| GW_API_VERSION       | 1.0.0                                                                                                         | Tag for the Gateway API version to check out. Set to `main` to get the latest version                                     |
| KIND_IMAGE           | Latest kind image, as defined in the tests/Dockerfile                                                         | The kind image to use                                                                                                     |
| KIND_KUBE_CONFIG     | ~/.kube/kind/config                                                                                           | The location of the kubeconfig                                                                                            |
| GATEWAY_CLASS        | nginx                                                                                                         | The gateway class that should be used for the tests                                                                       |
| SUPPORTED_FEATURES   | HTTPRoute,HTTPRouteQueryParamMatching, HTTPRouteMethodMatching,HTTPRoutePortRedirect, HTTPRouteSchemeRedirect | The supported features that should be tested by the conformance tests. Ensure the list is comma separated with no spaces. |
| EXEMPT_FEATURES      | ReferenceGrant                                                                                                | The features that should not be tested by the conformance tests                                                           |
| NGINX_IMAGE          | as defined in the provisioner/static-deployment.yaml file                                                     | The NGINX image for the NGF deployments                                                                                   |
| NGF_MANIFEST         | ../deploy/manifests/nginx-gateway.yaml                                                                        | The location of the NGF manifest                                                                                          |
| SERVICE_MANIFEST     | ../deploy/manifests/service/nodeport.yaml                                                                     | The location of the NGF Service manifest                                                                                  |
| STATIC_MANIFEST      | provisioner/static-deployment.yaml                                                                            | The location of the NGF static deployment manifest                                                                        |
| PROVISIONER_MANIFEST | provisioner/provisioner.yaml                                                                                  | The location of the NGF provisioner manifest                                                                              |
| INSTALL_WEBHOOK      | false                                                                                                         | Install the Gateway API Validating Webhook. Necessary for Kubernetes versions < 1.25.                                     |

### Step 1 - Create a kind Cluster

```makefile
make create-kind-cluster
```

> Note: The default kind cluster deployed is the latest available version. You can specify a different version by
> defining the kind image to use through the KIND_IMAGE variable, e.g.

```makefile
make create-kind-cluster KIND_IMAGE=kindest/node:v1.27.3
```

### Step 2 - Install NGINX Gateway Fabric to configured kind cluster

> Note: If you want to run the latest conformance tests from the Gateway API `main` branch, set the following
> environment variable before deploying NGF:

```bash
 export GW_API_VERSION=main
```

> Otherwise, the latest stable version will be used by default.

#### *Option 1* Build and install NGINX Gateway Fabric from local to configured kind cluster

```makefile
make install-ngf-local-build
```

#### *Option 2* Install NGINX Gateway Fabric from local already built image to configured kind cluster
You can optionally skip the actual *build* step.

```makefile
make install-ngf-local-no-build
```

> Note:  If choosing this option, the following step *must* be completed manually *before* you build the image:

```makefile
make update-ngf-manifest NGF_PREFIX=<ngf_repo_name> NGF_TAG=<ngf_image_tag>
```

#### *Option 3* Install NGINX Gateway Fabric from edge to configured kind cluster
You can also skip the build NGF image step and prepare the environment to instead use the `edge` image

```makefile
make install-ngf-edge
```

### Step 3 - Build conformance test runner image

> Note: If you want to run the latest conformance tests from the Gateway API `main` branch, run the following
> make command to update the Go modules to `main`:

 ```makefile
 make update-go-modules
 ```

> You can also point to a specific fork/branch by running:

 ```bash
 go mod edit -replace=sigs.k8s.io/gateway-api=<your-fork>@<your-branch>
 go mod download
 go mod verify
 go mod tidy
 ```

> Otherwise, the latest stable version will be used by default.

```makefile
make build-test-runner-image
```

### Step 4 - Run Gateway conformance tests

```makefile
make run-conformance-tests
```

### Step 5 - Cleanup the conformance test fixtures and uninstall NGINX Gateway Fabric

```makefile
make cleanup-conformance-tests
```

```makefile
make uninstall-ngf
```

### Step 6 - Revert changes to Go modules
**Optional** Not required if you aren't running the `main` Gateway API tests.

```makefile
make reset-go-modules
```

### Step 7 - Delete kind cluster

```makefile
make delete-kind-cluster
```
