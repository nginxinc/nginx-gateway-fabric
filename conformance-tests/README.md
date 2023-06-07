# Running [Gateway Conformance Tests](https://gateway-api.sigs.k8s.io/concepts/conformance/#3-conformance-tests) in Kind

## Prerequisites:

* [Kind](https://kind.sigs.k8s.io/).
* Docker.
* Golang.

**Note**: all commands in steps below are executed from the ```conformance-tests``` directory

List available commands:

```bash
$ make

build-test-image               Build conformance test image
create-kind-cluster            Create a KinD cluster
delete-kind-cluster            Delete kind cluster
help                           Display this help
install-nkg                    Install NKG on configured KinD cluster
run-conformance-tests          Run conformance tests
uninstall-nkg                  Uninstall NKG from configured KinD cluster
update-test-kind-config        Update KinD config
```
### Step 1 - Create a Kind Cluster

```bash
$ make create-kind-cluster
```

### Step 2 - Build conformance test image
```bash
$ make build-test-image
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

### Step 6 - Delete KinD cluster
```bash
$ make delete-kind-cluster
```