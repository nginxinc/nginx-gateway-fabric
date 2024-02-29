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
GINKGO_LABEL=functional
GINKGO_FLAGS=

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

.PHONY: build-images
build-images: ## Build NGF and NGINX images
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) build-images

.PHONY: build-images-with-plus
build-images-with-plus: ## Build NGF and NGINX Plus images
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) build-images-with-plus

.PHONY: load-images
load-images: ## Load NGF and NGINX images on configured kind cluster
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) load-images

.PHONY: load-images-with-plus
load-images-with-plus: ## Load NGF and NGINX Plus images on configured kind cluster
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) load-images-with-plus

.PHONY: install-ngf-local
install-ngf-local: ## Install NGF from local build on configured kind cluster.
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) install-ngf-local-build

.PHONY: install-ngf-local-with-plus
install-ngf-local-with-plus: ## Install NGF with NGINX Plus from local build on configured kind cluster.
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) install-ngf-local-build-with-plus

test: ## Run the system tests against your default k8s cluster
	go test -v ./suite -args --gateway-api-version=$(GW_API_VERSION) \
		--gateway-api-prev-version=$(GW_API_PREV_VERSION) --image-tag=$(TAG) --plus-enabled=$(PLUS_ENABLED) \
		--ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PREFIX) --pull-policy=$(PULL_POLICY) \
		--k8s-version=$(K8S_VERSION) --service-type=$(GW_SERVICE_TYPE) --is-gke-internal-lb=$(GW_SVC_GKE_INTERNAL)

functional-test: ## Run the functional tests against your default k8s cluster
	go test -v ./suite $(GINKGO_FLAGS) -args --gateway-api-version=$(GW_API_VERSION) \
		--gateway-api-prev-version=$(GW_API_PREV_VERSION) --image-tag=$(TAG) --plus-enabled=$(PLUS_ENABLED) \
		--ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PREFIX) --pull-policy=$(PULL_POLICY) \
		--k8s-version=$(K8S_VERSION) --service-type=$(GW_SERVICE_TYPE) --is-gke-internal-lb=$(GW_SVC_GKE_INTERNAL)

functional-test-nginx-plus: ## Run the functional tests against your default k8s cluster
	go test -v ./suite $(GINKGO_FLAGS) -args --gateway-api-version=$(GW_API_VERSION) \
		--gateway-api-prev-version=$(GW_API_PREV_VERSION) --image-tag=$(TAG) --plus-enabled=true \
		--ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PLUS_PREFIX) --pull-policy=$(PULL_POLICY) \
		--k8s-version=$(K8S_VERSION) --service-type=$(GW_SERVICE_TYPE) --is-gke-internal-lb=$(GW_SVC_GKE_INTERNAL)

.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster

.PHONY: run-tests-on-vm
run-tests-on-vm: ## Run the tests on a GCP VM
	bash scripts/run-tests-gcp-vm.sh

.PHONY: create-and-setup-vm
create-and-setup-vm: ## Create and setup a GCP VM for tests
	bash scripts/create-and-setup-gcp-vm.sh

.PHONY: cleanup-vm
cleanup-vm: ## Delete the test GCP VM and delete the firewall rule
	bash scripts/cleanup-vm.sh

.PHONY: create-gke-router
create-gke-router: ## Create a GKE router to allow egress traffic from private nodes (allows for external image pulls)
	bash scripts/create-gke-router.sh

.PHONY: cleanup-router
cleanup-router: ## Delete the GKE router
	bash scripts/cleanup-router.sh

.PHONY: setup-gcp-and-run-tests
setup-gcp-and-run-tests: create-gke-router create-and-setup-vm run-tests-on-vm ## Create and setup a GKE router and GCP VM for tests and run the tests

.PHONY: cleanup-gcp
cleanup-gcp: cleanup-router cleanup-vm ## Cleanup all GCP resources
