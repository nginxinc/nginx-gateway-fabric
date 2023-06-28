# Running [Gateway Conformance Tests](https://gateway-api.sigs.k8s.io/concepts/conformance/#3-conformance-tests) in kind

## Prerequisites:

* [kind](https://kind.sigs.k8s.io/).
* Docker.
* Golang.
* [yq](https://github.com/mikefarah/yq/#install)

**Note**: all commands in steps below are executed from the ```conformance``` directory

List available commands:

```bash
$ make

build-and-load-images          Build NKG container and load it and NGINX container on configured kind cluster
build-test-runner-image        Build conformance test runner image
cleanup-conformance-tests      Clean up conformance tests fixtures
create-kind-cluster            Create a kind cluster
delete-kind-cluster            Delete kind cluster
help                           Display this help
install-nkg-edge               Install NKG with provisioner from edge on configured kind cluster
install-nkg-local-build        Install NKG from local build with provisioner on configured kind cluster
preload-nginx-container        Preload NGINX container on configured kind cluster
prepare-nkg-dependencies       Install NKG dependencies on configured kind cluster
run-conformance-tests          Run conformance tests
undo-image-update              Undo the NKG image name and tag in deployment manifest
uninstall-nkg                  Uninstall NKG on configured kind cluster
```

**Note:** The following variables are configurable when running the below `make` commands:

| Variable  | Default | Description |
| ------------- | ------------- | ------------- |
| TAG | latest  | The tag for the conformance test image |
| PREFIX | conformance-test-runner | The prefix for the conformance test image |
| BUILD_NKG | true | Flag to indicate if the local NKG image needs to be built |
| NKG_TAG  | edge  | The tag for the locally built NKG image |
| NKG_PREFIX | nginx-kubernetes-gateway  | The prefix for the locally built NKG image |
| KIND_KUBE_CONFIG_FOLDER | ~/.kube/kind  | The location of the kubeconfig folder |
| GATEWAY_CLASS | nginx | The gateway class that should be used for the tests |
| SUPPORTED_FEATURES | HTTPRoute,HTTPRouteQueryParamMatching, HTTPRouteMethodMatching,HTTPRoutePortRedirect, HTTPRouteSchemeRedirect | The supported features that should be tested by the conformance tests. Ensure the list is comma separated with no spaces. |
| EXEMPT_FEATURES | ReferenceGrant | The features that should not be tested by the conformance tests |
| NGINX_IMAGE | as defined in the ../deploy/manifests/deployment.yaml file  | The NGINX image for the NKG deployments |
| NKG_DEPLOYMENT_MANIFEST | ../deploy/manifests/deployment.yaml | The location of the NKG deployment manifest |

### Step 1 - Create a kind Cluster

```bash
$ make create-kind-cluster
```
### Step 2 - Install Nginx Kubernetes Gateway to configured kind cluster

#### *Option 1* Build and install Nginx Kubernetes Gateway from local to configured kind cluster
```bash
$ make install-nkg-local-build
```

**Note:** You can optionally skip the actual *build* step by setting the BUILD_NKG flag to "false". However, if choosing 
this option, the following step *must* be completed manually *before* the build step:
 * Set NKG_PREFIX=<nkg_repo_name> NKG_TAG=<nkg_image_tag> to preferred values.
 * Navigate to `deploy/manifests` and update values in `deployment.yaml` as specified in below code-block.
 * Save the changes.
 ```
 .
 ..
 containers:
 - image: <nkg_repo_name>:<nkg_image_tag>
   imagePullPolicy: Never
 ..
 .
 ```

#### *Option 2* Install Nginx Kubernetes Gateway from edge to configured kind cluster
Instead of the above command, you can skip the build NKG image step and prepare the environment to instead
use the `edge` image

```bash
$ make install-nkg-edge
```

### Step 3 - Build conformance test runner image
```bash
$ make build-test-runner-image
```

### Step 4 - Run Gateway conformance tests
```bash
$ make run-conformance-tests
```

### Step 5 - Cleanup the conformance test fixtures and uninstall Nginx Kubernetes Gateway
```bash
$ make cleanup-conformance-tests
$ make uninstall-nkg
```

### Step 6 - Revert changes to the NKG deployment manifest
**Optional** Not required if using `edge` image
**Warning**: `make undo-image-update` will hard reset changes to the deploy/manifests/deployment.yaml file!
```bash
$ make undo-image-update
```

### Step 7 - Delete kind cluster
```bash
$ make delete-kind-cluster
```
