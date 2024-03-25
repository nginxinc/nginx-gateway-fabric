TAG = edge
PREFIX = nginx-gateway-fabric
NGINX_PREFIX = $(PREFIX)/nginx
NGINX_PLUS_PREFIX = $(PREFIX)/nginx-plus
PLUS_ENABLED=false
PULL_POLICY=Never
GW_API_PREV_VERSION ?= 1.0.0 ## Supported Gateway API version from previous NGF release
GW_API_VERSION ?= 1.0.0 ## Supported Gateway API version for NGF under test
K8S_VERSION ?= latest ## Expected format: 1.24 (major.minor) or latest
GW_SERVICE_TYPE=NodePort
GW_SVC_GKE_INTERNAL=false
GINKGO_LABEL=
GINKGO_FLAGS=
NGF_VERSION=
CI=false
TELEMETRY_ENDPOINT=
TELEMETRY_ENDPOINT_INSECURE=

ifneq ($(GINKGO_LABEL),)
    override GINKGO_FLAGS += -ginkgo.label-filter "$(GINKGO_LABEL)"
endif

# Check if PLUS_ENABLED is true
ifeq ($(PLUS_ENABLED),true)
    # If true, set NGINX_PREFIX to $NGINX_PLUS_PREFIX
    NGINX_PREFIX := $(NGINX_PLUS_PREFIX)
endif

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	cd .. && make create-kind-cluster

.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster

.PHONY: build-images
build-images: ## Build NGF and NGINX images
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) TELEMETRY_ENDPOINT=$(TELEMETRY_ENDPOINT) TELEMETRY_ENDPOINT_INSECURE=$(TELEMETRY_ENDPOINT_INSECURE) build-images

.PHONY: build-images-with-plus
build-images-with-plus: ## Build NGF and NGINX Plus images
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) TELEMETRY_ENDPOINT=$(TELEMETRY_ENDPOINT) TELEMETRY_ENDPOINT_INSECURE=$(TELEMETRY_ENDPOINT_INSECURE) build-images-with-plus

.PHONY: load-images
load-images: ## Load NGF and NGINX images on configured kind cluster
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) load-images

.PHONY: load-images-with-plus
load-images-with-plus: ## Load NGF and NGINX Plus images on configured kind cluster
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) load-images-with-plus

.PHONY: setup-gcp-and-run-tests
setup-gcp-and-run-tests: create-gke-router create-and-setup-vm run-tests-on-vm ## Create and setup a GKE router and GCP VM for tests and run the functional tests

.PHONY: setup-gcp-and-run-nfr-tests
setup-gcp-and-run-nfr-tests: create-gke-router create-and-setup-vm nfr-test ## Create and setup a GKE router and GCP VM for tests and run the NFR tests

.PHONY: create-gke-cluster
create-gke-cluster: ## Create a GKE cluster
	bash scripts/create-gke-cluster.sh $(CI)

.PHONY: create-and-setup-vm
create-and-setup-vm: ## Create and setup a GCP VM for tests
	bash scripts/create-and-setup-gcp-vm.sh

.PHONY: create-gke-router
create-gke-router: ## Create a GKE router to allow egress traffic from private nodes (allows for external image pulls)
	bash scripts/create-gke-router.sh

.PHONY: sync-files-to-vm
sync-files-to-vm: ## Syncs your local NGF files with the NGF repo on the VM
	bash scripts/sync-files-to-vm.sh

.PHONY: run-tests-on-vm
run-tests-on-vm: ## Run the functional tests on a GCP VM
	bash scripts/run-tests-gcp-vm.sh

.PHONY: nfr-test
nfr-test: ## Run the NFR tests on a GCP VM
	NFR=true bash scripts/run-tests-gcp-vm.sh

.PHONY: start-longevity-test
start-longevity-test: ## Start the longevity test to run for 4 days in GKE
	START_LONGEVITY=true $(MAKE) nfr-test

.PHONY: stop-longevity-test
stop-longevity-test: ## Stops the longevity test and collects results
	STOP_LONGEVITY=true $(MAKE) nfr-test

.PHONY: .vm-nfr-test
.vm-nfr-test: ## Runs the NFR tests on the GCP VM (called by `nfr-test`)
	go test -v -timeout 1h ./suite -ginkgo.label-filter "nfr" $(GINKGO_FLAGS) -ginkgo.v -args --gateway-api-version=$(GW_API_VERSION) \
		--gateway-api-prev-version=$(GW_API_PREV_VERSION) --image-tag=$(TAG) --version-under-test=$(NGF_VERSION) \
		--plus-enabled=$(PLUS_ENABLED) --ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PREFIX) --nginx-plus-image-repo=$(NGINX_PLUS_PREFIX) \
		--pull-policy=$(PULL_POLICY) --k8s-version=$(K8S_VERSION) --service-type=$(GW_SERVICE_TYPE) \
		--is-gke-internal-lb=$(GW_SVC_GKE_INTERNAL)

.PHONY: test
test: ## Runs the functional tests on your default k8s cluster
	go test -v ./suite -ginkgo.label-filter "functional" $(GINKGO_FLAGS) -args --gateway-api-version=$(GW_API_VERSION) \
		--gateway-api-prev-version=$(GW_API_PREV_VERSION) --image-tag=$(TAG) --version-under-test=$(NGF_VERSION) \
		--plus-enabled=$(PLUS_ENABLED) --ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PREFIX) \
		--pull-policy=$(PULL_POLICY) --k8s-version=$(K8S_VERSION) --service-type=$(GW_SERVICE_TYPE) \
		--is-gke-internal-lb=$(GW_SVC_GKE_INTERNAL)

.PHONY: test-with-plus
test-with-plus: ## Runs the functional tests for NGF with NGINX Plus on your default k8s cluster
	make test PLUS_ENABLED=true

.PHONY: cleanup-gcp
cleanup-gcp: cleanup-router cleanup-vm delete-gke-cluster ## Cleanup all GCP resources

.PHONY: cleanup-router
cleanup-router: ## Delete the GKE router
	bash scripts/cleanup-router.sh

.PHONY: cleanup-vm
cleanup-vm: ## Delete the test GCP VM and delete the firewall rule
	bash scripts/cleanup-vm.sh

.PHONY: delete-gke-cluster
delete-gke-cluster: ## Delete the GKE cluster
	bash scripts/delete-gke-cluster.sh

.PHONY: add-local-ip-to-cluster
add-local-ip-to-cluster: ## Add local IP to the GKE cluster master-authorized-networks
	bash scripts/add-local-ip-auth-networks.sh
