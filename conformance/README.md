# Running [Gateway Conformance Tests](https://gateway-api.sigs.k8s.io/concepts/conformance/#3-conformance-tests) in kind

## Prerequisites:

* [kind](https://kind.sigs.k8s.io/).
* Docker.
* Golang.

**Note**: all commands in steps below are executed from the ```conformance``` directory

List available commands:

```bash
$ make

build-test-image               Build conformance test image
create-kind-cluster            Create a kind cluster
delete-kind-cluster            Delete kind cluster
help                           Display this help
install-nkg                    Install NKG on configured kind cluster
run-conformance-tests          Run conformance tests
uninstall-nkg                  Uninstall NKG from configured kind cluster
update-test-kind-config        Update kind config
```
### Step 1 - Create a kind Cluster

```bash
$ make create-kind-cluster
```

### Step 2 - Build and load Nginx Kubernetes Gateway container to configured kind cluster
**Note**: this step is only required when user wants to run conformance tests using locally built image of Nginx Kubernetes Gateway

```bash
$ make NKG_PREFIX=<repo_name> NKG_TAG=<image_tag> prepare-nkg
```

### Step 3 - Update NKG deploymeny and provisioner manifests
**Note**: this step is only required when user wants to run conformance tests using locally built image of Nginx Kubernetes Gateway
* Navigate to `deploy/manifests` and update values in `deployment.yaml` as specified in below code-block.
* Navigate to `conformance/provisioner` and update values in `provisioner.yaml` as specified in below code-block.
```
.
..
containers:
- image: <repo_name>:<image_tag>
  imagePullPolicy: Never
..
.
```

### Step 2 - Build conformance test runner image
```bash
$ make build-test-runner-image
```

### Step 3 - Install Nginx Kubernetes Gateway
```bash
$ make install-nkg
```

### Step 4 - Run Gateway conformance tests
```bash
$ make run-conformance-tests
```

### Step 5 - Uninstall Nginx Kubernetes Gateway
```bash
$ make uninstall-nkg
```

### Step 7 - Delete kind cluster
```bash
$ make delete-kind-cluster
```
