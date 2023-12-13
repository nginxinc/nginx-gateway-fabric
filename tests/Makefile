TAG = edge
PREFIX = nginx-gateway-fabric
NGINX_PREFIX = $(PREFIX)/nginx
PULL_POLICY=Never
GW_API_VERSION ?= 1.0.0
K8S_VERSION ?= latest ## Expected format: 1.24 (major.minor) or latest
GW_SERVICE_TYPE=NodePort
GW_SVC_GKE_INTERNAL=false
GINKGO_FLAGS=

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	cd .. && make create-kind-cluster

.PHONY: build-images
build-images: ## Build NGF and NGINX images
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) build-images

.PHONY: load-images
load-images: ## Load NGF and NGINX images on configured kind cluster
	kind load docker-image $(PREFIX):$(TAG) $(NGINX_PREFIX):$(TAG)

test: ## Run the system tests against your default k8s cluster
	go test -v ./suite $(GINKGO_FLAGS) -args --gateway-api-version=$(GW_API_VERSION) --image-tag=$(TAG) \
		--ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PREFIX) --pull-policy=$(PULL_POLICY) \
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

.PHONY: create-vm-and-run-tests
create-vm-and-run-tests: create-and-setup-vm run-tests-on-vm ## Create and setup a GCP VM for tests and run the tests

.PHONY: cleanup-vm
cleanup-vm: ## Delete the test GCP VM
	source utils/vars.env && gcloud compute instances delete ${VM_NAME} --project=${GKE_PROJECT} --zone=${GKE_CLUSTER_ZONE}
