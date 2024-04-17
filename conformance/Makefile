TAG = edge
VERSION = 1.2.0 # NGF version
PREFIX = nginx-gateway-fabric
NGINX_PREFIX = $(PREFIX)/nginx
NGINX_PLUS_PREFIX ?= $(PREFIX)/nginx-plus
GW_API_VERSION ?= 1.0.0
GATEWAY_CLASS = nginx
SUPPORTED_FEATURES = HTTPRouteQueryParamMatching,HTTPRouteMethodMatching,HTTPRoutePortRedirect,HTTPRouteSchemeRedirect,HTTPRouteHostRewrite,HTTPRoutePathRewrite,GatewayPort8080
KIND_IMAGE ?= $(shell grep -m1 'FROM kindest/node' <tests/Dockerfile | awk -F'[ ]' '{print $$2}')
KIND_KUBE_CONFIG=$${HOME}/.kube/kind/config
CONFORMANCE_TAG = latest
CONFORMANCE_PREFIX = conformance-test-runner
NGF_MANIFEST=../deploy/manifests/nginx-gateway.yaml
CRDS=../deploy/manifests/crds/
STATIC_MANIFEST=provisioner/static-deployment.yaml
PROVISIONER_MANIFEST=provisioner/provisioner.yaml
ENABLE_EXPERIMENTAL ?= false
.DEFAULT_GOAL := help

ifeq ($(ENABLE_EXPERIMENTAL),true)
    SUPPORTED_FEATURES +=,GRPCExactMethodMatching,GRPCRouteListenerHostnameMatching
endif

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: update-go-modules
update-go-modules: ## Update the gateway-api go modules to latest main version
	go get -u sigs.k8s.io/gateway-api@main
	go mod tidy

.PHONY: build-test-runner-image
build-test-runner-image: ## Build conformance test runner image
	docker build -t $(CONFORMANCE_PREFIX):$(CONFORMANCE_TAG) -f tests/Dockerfile ..

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	kind create cluster --image $(KIND_IMAGE)
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG)

.PHONY: update-ngf-manifest
update-ngf-manifest: ## Update the NGF deployment manifest image names and imagePullPolicies
	cd .. && make generate-manifests HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE="--set nginxGateway.kind=skip" HELM_TEMPLATE_COMMON_ARGS="--set nginxGateway.image.repository=$(PREFIX) --set nginxGateway.image.tag=$(TAG) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=$(NGINX_PREFIX) --set nginx.image.tag=$(TAG) --set nginx.image.pullPolicy=Never --set nginxGateway.gwAPIExperimentalFeatures.enable=$(ENABLE_EXPERIMENTAL)" && cd -

.PHONY: update-ngf-manifest-with-plus
update-ngf-manifest-with-plus: ## Update the NGF deployment manifest image names and imagePullPolicies including nginx-plus
	cd .. && make generate-manifests HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE="--set nginxGateway.kind=skip" HELM_TEMPLATE_COMMON_ARGS="--set nginxGateway.image.repository=$(PREFIX) --set nginxGateway.image.tag=$(TAG) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=$(NGINX_PLUS_PREFIX) --set nginx.image.tag=$(TAG) --set nginx.image.pullPolicy=Never --set nginx.plus=true" && cd -

.PHONY: build-images
build-images: ## Build NGF and NGINX images
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) build-images

.PHONY: load-images
load-images: ## Load NGF and NGINX images on configured kind cluster
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) load-images

.PHONY: build-images-with-plus
build-images-with-plus: ## Build NGF and NGINX Plus images
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) build-images-with-plus

.PHONY: load-images-with-plus
load-images-with-plus: ## Load NGF and NGINX Plus images on configured kind cluster
	cd .. && make PREFIX=$(PREFIX) TAG=$(TAG) load-images-with-plus

.PHONY: prepare-ngf-dependencies
prepare-ngf-dependencies: update-ngf-manifest ## Install NGF dependencies on configured kind cluster
	./scripts/install-gateway.sh $(GW_API_VERSION) $(ENABLE_EXPERIMENTAL)
	kubectl apply -f $(CRDS)
	kubectl apply -f $(NGF_MANIFEST)

.PHONY: prepare-ngf-dependencies-with-plus
prepare-ngf-dependencies-with-plus: update-ngf-manifest-with-plus ## Install NGF dependencies with Plus on configured kind cluster
	./scripts/install-gateway.sh $(GW_API_VERSION)
	kubectl apply -f $(CRDS)
	kubectl apply -f $(NGF_MANIFEST)

.PHONY: deploy-updated-provisioner
deploy-updated-provisioner: ## Update provisioner manifest and deploy to the configured kind cluster
	yq '(select(di != 3))' $(PROVISIONER_MANIFEST) | kubectl apply -f -
	yq '(select(.spec.template.spec.containers[].image) | .spec.template.spec.containers[].image="$(PREFIX):$(TAG)" | .spec.template.spec.containers[].imagePullPolicy = "Never")' $(PROVISIONER_MANIFEST) | kubectl apply -f -

.PHONY: install-ngf-local-build
install-ngf-local-build: prepare-ngf-dependencies build-images load-images deploy-updated-provisioner ## Install NGF from local build with provisioner on configured kind cluster

.PHONY: install-ngf-local-no-build
install-ngf-local-no-build: prepare-ngf-dependencies load-images deploy-updated-provisioner ## Install NGF from local build with provisioner on configured kind cluster but do not build the NGF image

.PHONY: install-ngf-local-build-with-plus
install-ngf-local-build-with-plus: prepare-ngf-dependencies-with-plus build-images-with-plus load-images-with-plus deploy-updated-provisioner ## Install NGF with Plus from local build with provisioner on configured kind cluster

.PHONY: install-ngf-local-no-build-with-plus
install-ngf-local-no-build-with-plus: prepare-ngf-dependencies-with-plus load-images-with-plus deploy-updated-provisioner ## Install NGF with Plus from local build with provisioner on configured kind cluster but do not build the NGF image

.PHONY: install-ngf-edge
install-ngf-edge: prepare-ngf-dependencies ## Install NGF with provisioner from edge on configured kind cluster
	kubectl apply -f $(PROVISIONER_MANIFEST)

.PHONY: run-conformance-tests
run-conformance-tests: ## Run conformance tests
	kind load docker-image $(CONFORMANCE_PREFIX):$(CONFORMANCE_TAG)
	kubectl apply -f tests/conformance-rbac.yaml
	kubectl run -i conformance \
		--image=$(CONFORMANCE_PREFIX):$(CONFORMANCE_TAG) --image-pull-policy=Never \
		--overrides='{ "spec": { "serviceAccountName": "conformance" }  }' \
		--restart=Never -- sh -c "go test -v . -tags conformance,experimental -args --gateway-class=$(GATEWAY_CLASS) \
						        --supported-features=$(SUPPORTED_FEATURES) --version=$(VERSION) \
								--report-output=output.txt; cat output.txt" | tee output.txt
	bash scripts/check-pod-exit-code.sh
	sed -e '1,/CONFORMANCE PROFILE/d' output.txt > conformance-profile.yaml
	rm output.txt
	[ $(shell cat conformance-profile.yaml | yq '.profiles[0].core.result') != "failure" ] \
		&& [ $(shell cat conformance-profile.yaml | yq '.profiles[0].extended.result') != "failure" ] || exit 2

.PHONY: cleanup-conformance-tests
cleanup-conformance-tests: ## Clean up conformance tests fixtures
	kubectl delete pod conformance
	kubectl delete -f tests/conformance-rbac.yaml

.PHONY: uninstall-ngf
uninstall-ngf: uninstall-k8s-components undo-manifests-update ## Uninstall NGF on configured kind cluster and undo manifest changes

.PHONY: uninstall-k8s-components
uninstall-k8s-components: ## Uninstall installed components on configured kind cluster
	-kubectl delete -f $(NGF_MANIFEST)
	./scripts/uninstall-gateway.sh $(GW_API_VERSION) $(ENABLE_EXPERIMENTAL)
	kubectl delete clusterrole nginx-gateway-provisioner
	kubectl delete clusterrolebinding nginx-gateway-provisioner

.PHONY: undo-manifests-update
undo-manifests-update: ## Undo the changes in the manifest files
	cd .. && make generate-manifests && cd -

.PHONY: reset-go-modules
reset-go-modules: ## Reset the go modules changes
	git checkout -- ../go.mod ../go.sum

.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster
