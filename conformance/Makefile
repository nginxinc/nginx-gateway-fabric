NGF_TAG = edge
NGF_PREFIX = nginx-gateway-fabric
NGINX_IMAGE_NAME = $(NGF_PREFIX)/nginx
GW_API_VERSION ?= 0.8.1
GATEWAY_CLASS = nginx
SUPPORTED_FEATURES = HTTPRoute,HTTPRouteQueryParamMatching,HTTPRouteMethodMatching,HTTPRoutePortRedirect,HTTPRouteSchemeRedirect,GatewayClassObservedGenerationBump
KIND_IMAGE ?= $(shell grep -m1 'FROM kindest/node' <tests/Dockerfile | awk -F'[ ]' '{print $$2}')
KIND_KUBE_CONFIG=$${HOME}/.kube/kind/config
TAG = latest
PREFIX = conformance-test-runner
NGF_MANIFEST=../deploy/manifests/nginx-gateway.yaml
CRDS=../deploy/manifests/crds/
STATIC_MANIFEST=provisioner/static-deployment.yaml
PROVISIONER_MANIFEST=provisioner/provisioner.yaml
NGINX_IMAGE=$(shell yq '.spec.template.spec.containers[1].image as $$nginx_ver | $$nginx_ver' $(STATIC_MANIFEST))
.DEFAULT_GOAL := help

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: update-go-modules
update-go-modules: ## Update the gateway-api go modules to latest main version
	go get -u sigs.k8s.io/gateway-api@main
	go mod tidy

.PHONY: build-test-runner-image
build-test-runner-image: ## Build conformance test runner image
	docker build -t $(PREFIX):$(TAG) -f tests/Dockerfile ..

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	kind create cluster --image $(KIND_IMAGE)
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG)

.PHONY: update-ngf-manifest
update-ngf-manifest: ## Update the NGF deployment manifest image names and imagePullPolicies
	cd .. && make generate-manifests HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE="--set nginxGateway.kind=skip" HELM_TEMPLATE_COMMON_ARGS="--set nginxGateway.image.repository=$(NGF_PREFIX) --set nginxGateway.image.tag=$(NGF_TAG) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=$(NGINX_IMAGE_NAME) --set nginx.image.tag=$(NGF_TAG) --set nginx.image.pullPolicy=Never" && cd -

.PHONY: build-images
build-images: ## Build NGF and nginx images
	cd .. && make PREFIX=$(NGF_PREFIX) TAG=$(NGF_TAG) build-images

.PHONY: load-images
load-images: ## Load NGF and NGINX images on configured kind cluster
	kind load docker-image $(NGF_PREFIX):$(NGF_TAG) $(NGINX_IMAGE_NAME):$(NGF_TAG)

.PHONY: prepare-ngf-dependencies
prepare-ngf-dependencies: update-ngf-manifest ## Install NGF dependencies on configured kind cluster
	./scripts/install-gateway.sh $(GW_API_VERSION)
	kubectl wait --for=condition=available --timeout=60s deployment gateway-api-admission-server -n gateway-system
	kubectl apply -f $(CRDS)
	kubectl apply -f $(NGF_MANIFEST)

.PHONY: deploy-updated-provisioner
deploy-updated-provisioner: ## Update provisioner manifest and deploy to the configured kind cluster
	yq '(select(di != 3))' $(PROVISIONER_MANIFEST) | kubectl apply -f -
	yq '(select(.spec.template.spec.containers[].image) | .spec.template.spec.containers[].image="$(NGF_PREFIX):$(NGF_TAG)" | .spec.template.spec.containers[].imagePullPolicy = "Never")' $(PROVISIONER_MANIFEST) | kubectl apply -f -

.PHONY: install-ngf-local-build
install-ngf-local-build: prepare-ngf-dependencies build-images load-images deploy-updated-provisioner ## Install NGF from local build with provisioner on configured kind cluster

.PHONY: install-ngf-local-no-build
install-ngf-local-no-build: prepare-ngf-dependencies load-images deploy-updated-provisioner ## Install NGF from local build with provisioner on configured kind cluster but do not build the NGF image

.PHONY: install-ngf-edge
install-ngf-edge: prepare-ngf-dependencies ## Install NGF with provisioner from edge on configured kind cluster
	kubectl apply -f $(PROVISIONER_MANIFEST)

.PHONY: run-conformance-tests
run-conformance-tests: ## Run conformance tests
	kind load docker-image $(PREFIX):$(TAG)
	kubectl apply -f tests/conformance-rbac.yaml
	kubectl run -i conformance \
		--image=$(PREFIX):$(TAG) --image-pull-policy=Never \
		--overrides='{ "spec": { "serviceAccountName": "conformance" }  }' \
		--restart=Never -- sh -c "go test -v . -tags conformance,experimental -args --gateway-class=$(GATEWAY_CLASS) \
						        --supported-features=$(SUPPORTED_FEATURES) --version=$(VERSION) \
								--report-output=output.txt; cat output.txt" | tee output.txt
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
	./scripts/uninstall-gateway.sh $(GW_API_VERSION)
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
