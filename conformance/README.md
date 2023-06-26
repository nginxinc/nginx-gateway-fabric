# Running [Gateway Conformance Tests](https://gateway-api.sigs.k8s.io/concepts/conformance/#3-conformance-tests) in kind

## Prerequisites:

* [kind](https://kind.sigs.k8s.io/).
* Docker.
* Golang.
* [yq](https://github.com/mikefarah/yq/#macos--linux-via-homebrew)

**Note**: all commands in steps below are executed from the ```conformance``` directory

List available commands:

```bash
$ make

build-test-runner-image        Build conformance test runner image
create-kind-cluster            Create a kind cluster
delete-kind-cluster            Delete kind cluster
help                           Display this help
install-nkg                    Install NKG with provisioner on configured kind cluster
prepare-nkg                    Build and load NKG container on configured kind cluster
run-conformance-tests          Run conformance tests
uninstall-nkg                  Uninstall NKG on configured kind cluster
update-test-kind-config        Update kind config
```
### Step 1 - Create a kind Cluster

```bash
$ make create-kind-cluster
```

### Step 2 - Build Nginx Kubernetes Gateway container and load it and the NGINX container to configured kind cluster

```bash
$ make NKG_PREFIX=<repo_name> NKG_TAG=<image_tag> prepare-nkg

```
### Step 3 - Build conformance test runner image
```bash
$ make build-test-runner-image
```

### Step 4 - Install Nginx Kubernetes Gateway
```bash
$ make NKG_PREFIX=<repo_name> NKG_TAG=<image_tag> install-nkg
```

### Step 5 - Run Gateway conformance tests
```bash
$ make NKG_PREFIX=<repo_name> NKG_TAG=<image_tag> run-conformance-tests
```

### Step 6 - Uninstall Nginx Kubernetes Gateway
```bash
$ make uninstall-nkg
```

### Step 7 - Revert changes to the NKG deployment manifest
**Warning**: `make undo-image-update` will hard reset changes to the deploy/manifests/deployment.yaml file!
```bash
$ make undo-image-update
```

### Step 8 - Delete kind cluster
```bash
$ make delete-kind-cluster
```
