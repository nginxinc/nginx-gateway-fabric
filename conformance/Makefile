NKG_TAG = edge
NKG_PREFIX = nginx-kubernetes-gateway
GW_API_VERSION ?= 0.7.1
GATEWAY_CLASS = nginx
SUPPORTED_FEATURES = HTTPRoute,HTTPRouteQueryParamMatching,HTTPRouteMethodMatching,HTTPRoutePortRedirect,HTTPRouteSchemeRedirect,GatewayClassObservedGenerationBump
KIND_KUBE_CONFIG=$${HOME}/.kube/kind/config
TAG = latest
PREFIX = conformance-test-runner
NKG_MANIFEST=../deploy/manifests/nginx-gateway.yaml
SERVICE_MANIFEST=../deploy/manifests/service/nodeport.yaml
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
	$(eval KIND_IMAGE=$(shell grep -m1 'FROM kindest/node' <tests/Dockerfile | awk -F'[ ]' '{print $$2}'))
	kind create cluster --image $(KIND_IMAGE)
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG)

.PHONY: preload-nginx-container
preload-nginx-container: ## Preload NGINX container on configured kind cluster
	docker pull $(NGINX_IMAGE)
	kind load docker-image $(NGINX_IMAGE)

.PHONY: update-nkg-manifest
update-nkg-manifest: ## Update the NKG deployment manifest image name and imagePullPolicy
	cd .. && make generate-manifests HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE="--set nginxGateway.kind=skip" HELM_TEMPLATE_COMMON_ARGS="--set nginxGateway.image.repository=$(NKG_PREFIX) --set nginxGateway.image.tag=$(NKG_TAG) --set nginxGateway.image.pullPolicy=Never" && cd -

.PHONY: build-nkg-image
build-nkg-image: ## Build NKG container and load it and NGINX container on configured kind cluster
	cd .. && make PREFIX=$(NKG_PREFIX) TAG=$(NKG_TAG) container

.PHONY: load-images
load-images: preload-nginx-container ## Load NKG and NGINX containers on configured kind cluster
	kind load docker-image $(NKG_PREFIX):$(NKG_TAG)

.PHONY: prepare-nkg-dependencies
prepare-nkg-dependencies: update-nkg-manifest ## Install NKG dependencies on configured kind cluster
	./scripts/install-gateway.sh $(GW_API_VERSION)
	kubectl wait --for=condition=available --timeout=60s deployment gateway-api-admission-server -n gateway-system
	kubectl apply -f $(NKG_MANIFEST)
	kubectl apply -f $(SERVICE_MANIFEST)

.PHONY: deploy-updated-provisioner
deploy-updated-provisioner: ## Update provisioner manifest and deploy to the configured kind cluster
	yq '(select(di != 3))' $(PROVISIONER_MANIFEST) | kubectl apply -f -
	yq '(select(.spec.template.spec.containers[].image) | .spec.template.spec.containers[].image="$(NKG_PREFIX):$(NKG_TAG)" | .spec.template.spec.containers[].imagePullPolicy = "Never")' $(PROVISIONER_MANIFEST) | kubectl apply -f -

.PHONY: install-nkg-local-build
install-nkg-local-build: prepare-nkg-dependencies build-nkg-image load-images deploy-updated-provisioner ## Install NKG from local build with provisioner on configured kind cluster

.PHONY: install-nkg-local-no-build
install-nkg-local-no-build: prepare-nkg-dependencies load-images deploy-updated-provisioner ## Install NKG from local build with provisioner on configured kind cluster but do not build the NKG image

.PHONY: install-nkg-edge
install-nkg-edge: preload-nginx-container prepare-nkg-dependencies ## Install NKG with provisioner from edge on configured kind cluster
	kubectl apply -f $(PROVISIONER_MANIFEST)

.PHONY: run-conformance-tests
run-conformance-tests: ## Run conformance tests
	kind load docker-image $(PREFIX):$(TAG)
	kubectl apply -f tests/conformance-rbac.yaml
	kubectl run -i conformance \
		--image=$(PREFIX):$(TAG) --image-pull-policy=Never \
		--overrides='{ "spec": { "serviceAccountName": "conformance" }  }' \
		--restart=Never -- go test -v . -tags conformance -args --gateway-class=$(GATEWAY_CLASS) --debug \
						        --supported-features=$(SUPPORTED_FEATURES)

.PHONY: cleanup-conformance-tests
cleanup-conformance-tests: ## Clean up conformance tests fixtures
	kubectl delete pod conformance
	kubectl delete -f tests/conformance-rbac.yaml

.PHONY: uninstall-nkg
uninstall-nkg: uninstall-k8s-components undo-manifests-update ## Uninstall NKG on configured kind cluster and undo manifest changes

.PHONY: uninstall-k8s-components
uninstall-k8s-components: ## Uninstall installed components on configured kind cluster
	kubectl delete -f $(NKG_MANIFEST)
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
