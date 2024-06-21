CI=false
CLUSTER_NAME ?= kind
CONFORMANCE_PREFIX = conformance-test-runner## Prefix for the conformance test runner image
CONFORMANCE_TAG = latest## Tag for the conformance test runner image
GATEWAY_CLASS = nginx## Gateway class to use
GINKGO_FLAGS =
GINKGO_LABEL =
GW_API_VERSION ?= $(shell sed -n 's/.*ref=v\(.*\)/\1/p' ../config/crd/gateway-api/standard/kustomization.yaml)## Supported Gateway API version from current NGF
GW_API_PREV_VERSION ?= 1.1.0## Supported Gateway API version from previous NGF release
GW_SERVICE_TYPE = NodePort## Service type to use for the gateway
GW_SVC_GKE_INTERNAL = false
NGF_VERSION ?= $(shell git describe --tags $(shell git rev-list --tags --max-count=1))## NGF version to be tested (defaults to latest tag)
PULL_POLICY = Never## Pull policy for the images
PROVISIONER_MANIFEST = conformance/provisioner/provisioner.yaml
SUPPORTED_FEATURES = HTTPRouteQueryParamMatching,HTTPRouteMethodMatching,HTTPRoutePortRedirect,HTTPRouteSchemeRedirect,HTTPRouteHostRewrite,HTTPRoutePathRewrite,GatewayPort8080,HTTPRouteResponseHeaderModification,GRPCExactMethodMatching,GRPCRouteListenerHostnameMatching,GRPCRouteHeaderMatching

ifneq ($(GINKGO_LABEL),)
    override GINKGO_FLAGS += --label-filter "$(GINKGO_LABEL)"
endif

.PHONY: update-go-modules
update-go-modules: ## Update the gateway-api go modules to latest main version
	go get -u sigs.k8s.io/gateway-api@main
	go mod tidy

.PHONY: build-test-runner-image
build-test-runner-image: ## Build conformance test runner image
	docker build -t $(CONFORMANCE_PREFIX):$(CONFORMANCE_TAG) -f conformance/Dockerfile .

.PHONY: run-conformance-tests
run-conformance-tests: ## Run conformance tests
	kind load docker-image $(CONFORMANCE_PREFIX):$(CONFORMANCE_TAG) --name $(CLUSTER_NAME)
	kubectl apply -f conformance/conformance-rbac.yaml
	kubectl run -i conformance \
		--image=$(CONFORMANCE_PREFIX):$(CONFORMANCE_TAG) --image-pull-policy=Never \
		--overrides='{ "spec": { "serviceAccountName": "conformance" }  }' \
		--restart=Never -- sh -c "go test -v . -tags conformance,experimental -args --gateway-class=$(GATEWAY_CLASS) \
						        --supported-features=$(SUPPORTED_FEATURES) --version=$(NGF_VERSION) \
								--report-output=output.txt; cat output.txt" | tee output.txt
	bash scripts/check-pod-exit-code.sh
	sed -e '1,/CONFORMANCE PROFILE/d' output.txt > conformance-profile.yaml
	rm output.txt
	grpc_core_result=`yq '.profiles[0].core.result' conformance-profile.yaml`; \
	http_core_result=`yq '.profiles[1].core.result' conformance-profile.yaml`; \
	http_extended_result=`yq '.profiles[1].extended.result' conformance-profile.yaml`; \
	if [ "$$grpc_core_result" != "failure" ] && [ "$$http_core_result" != "failure" ] && [ "$$http_extended_result" != "failure" ] ; then \
		exit 0; \
	else \
		exit 2; \
	fi

.PHONY: cleanup-conformance-tests
cleanup-conformance-tests: ## Clean up conformance tests fixtures
	kubectl delete pod conformance
	kubectl delete -f tests/conformance-rbac.yaml

.PHONY: build
build: generate-static-deployment

.PHONY: reset-go-modules
reset-go-modules: ## Reset the go modules changes
	git checkout -- ../go.mod ../go.sum

-include ../Makefile

# Check if PLUS_ENABLED is true
ifeq ($(PLUS_ENABLED),true)
    # If true, set NGINX_PREFIX to $NGINX_PLUS_PREFIX
    NGINX_PREFIX := $(NGINX_PLUS_PREFIX)
endif

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
start-longevity-test: export START_LONGEVITY=true
start-longevity-test: nfr-test ## Start the longevity test to run for 4 days in GKE

.PHONY: stop-longevity-test
stop-longevity-test: export STOP_LONGEVITY=true
stop-longevity-test: nfr-test ## Stop the longevity test and collects results

.PHONY: .vm-nfr-test
.vm-nfr-test: ## Runs the NFR tests on the GCP VM (called by `nfr-test`)
	go run github.com/onsi/ginkgo/v2/ginkgo --randomize-all --randomize-suites --keep-going --fail-on-pending --trace -r -v \
	    --label-filter "nfr" $(GINKGO_FLAGS) ./suite --  --gateway-api-version=$(GW_API_VERSION) \
		--gateway-api-prev-version=$(GW_API_PREV_VERSION) --image-tag=$(TAG) --version-under-test=$(NGF_VERSION) \
		--plus-enabled=$(PLUS_ENABLED) --ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PREFIX) --nginx-plus-image-repo=$(NGINX_PLUS_PREFIX) \
		--pull-policy=$(PULL_POLICY) --service-type=$(GW_SERVICE_TYPE) \
		--is-gke-internal-lb=$(GW_SVC_GKE_INTERNAL)

.PHONY: test
test: ## Runs the functional tests on your default k8s cluster
	go run github.com/onsi/ginkgo/v2/ginkgo --randomize-all --randomize-suites --keep-going --fail-on-pending --trace -r -v \
		--label-filter "functional" $(GINKGO_FLAGS) ./suite -- \
		--gateway-api-version=$(GW_API_VERSION) --gateway-api-prev-version=$(GW_API_PREV_VERSION) \
		--image-tag=$(TAG) --version-under-test=$(NGF_VERSION) \
		--plus-enabled=$(PLUS_ENABLED) --ngf-image-repo=$(PREFIX) --nginx-image-repo=$(NGINX_PREFIX) --nginx-plus-image-repo=$(NGINX_PLUS_PREFIX) \
		--pull-policy=$(PULL_POLICY) --service-type=$(GW_SERVICE_TYPE) \
		--is-gke-internal-lb=$(GW_SVC_GKE_INTERNAL) --cluster_name=$(CLUSTER_NAME)

.PHONY: test-with-plus
test-with-plus: PLUS_ENABLED=true
test-with-plus: test ## Runs the functional tests for NGF with NGINX Plus on your default k8s cluster

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

HELM_PARAMETERS += --set nameOverride=nginx-gateway --set nginxGateway.kind=skip --set service.create=false

.PHONY: deploy-updated-provisioner
deploy-updated-provisioner: ## Update provisioner manifest and deploy to the configured kind cluster
	yq '(select(di != 3))' $(PROVISIONER_MANIFEST) | kubectl apply -f -
	yq '(select(.spec.template.spec.containers[].image) | .spec.template.spec.containers[].image="$(PREFIX):$(TAG)" | .spec.template.spec.containers[].imagePullPolicy = "Never")' $(PROVISIONER_MANIFEST) | kubectl apply -f -

.PHONY: generate-static-deployment
generate-static-deployment:
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) --set metrics.enable=false --set nginxGateway.productTelemetry.enable=false -n nginx-gateway -s templates/deployment.yaml --set nginxGateway.image.repository=$(PREFIX) --set nginxGateway.image.tag=$(TAG) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=$(NGINX_PREFIX) --set nginx.image.tag=$(TAG) --set nginx.image.pullPolicy=Never --set nginxGateway.gwAPIExperimentalFeatures.enable=$(ENABLE_EXPERIMENTAL) --set nginx.plus=$(PLUS_ENABLED) > $(SELF_DIR)config/tests/static-deployment.yaml

# this target is used to install the gateway-api CRDs from the main branch (only used in the nightly CI job)
# it overrides the target in the main Makefile when the GW_API_VERSION is set to main
ifeq ($(GW_API_VERSION),main)
install-gateway-crds:
	kubectl kustomize "https://github.com/kubernetes-sigs/gateway-api/config/crd/$(if $(filter true,$(ENABLE_EXPERIMENTAL)),experimental,)?timeout=120&ref=main" | kubectl apply -f -
endif

.PHONY: install-ngf-local-build
install-ngf-local-build: deploy-updated-provisioner

.PHONY: install-ngf-local-no-build
install-ngf-local-no-build: load-images helm-install-local deploy-updated-provisioner ## Install NGF from local build with provisioner on configured kind cluster but do not build the NGF image

.PHONY: install-ngf-local-build-with-plus
install-ngf-local-build-with-plus: deploy-updated-provisioner

.PHONY: install-ngf-local-no-build-with-plus
install-ngf-local-no-build-with-plus: load-images-with-plus helm-install-local-with-plus deploy-updated-provisioner ## Install NGF with Plus from local build with provisioner on configured kind cluster but do not build the NGF image

.PHONY: install-ngf-edge
install-ngf-edge: load-images helm-install-local ## Install NGF with provisioner from edge on configured kind cluster
	kubectl apply -f $(PROVISIONER_MANIFEST)

.PHONY: uninstall-ngf
uninstall-ngf: ## Uninstall NGF on configured kind cluster and undo manifest changes
	-helm uninstall nginx-gateway -n nginx-gateway
	-make uninstall-gateway-crds
	-kubectl delete clusterrole nginx-gateway-provisioner
	-kubectl delete clusterrolebinding nginx-gateway-provisioner
	-kubectl delete namespace nginx-gateway
	-kubectl kustomize ../config/crd | kubectl delete -f -
