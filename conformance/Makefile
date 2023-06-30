NKG_TAG = edge
NKG_PREFIX = nginx-kubernetes-gateway
GATEWAY_CLASS = nginx
SUPPORTED_FEATURES = HTTPRoute,HTTPRouteQueryParamMatching,HTTPRouteMethodMatching,HTTPRoutePortRedirect,HTTPRouteSchemeRedirect
KIND_KUBE_CONFIG=$${HOME}/.kube/kind/config
TAG = latest
PREFIX = conformance-test-runner
NKG_DEPLOYMENT_MANIFEST=../deploy/manifests/deployment.yaml
NGINX_IMAGE=$(shell yq '.spec.template.spec.containers[1].image as $$nginx_ver | $$nginx_ver' $(NKG_DEPLOYMENT_MANIFEST))
.DEFAULT_GOAL := help

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

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
	yq -i 'with(.spec.template.spec.containers[0]; .image = "$(NKG_PREFIX):$(NKG_TAG)" | .imagePullPolicy = "Never")' $(NKG_DEPLOYMENT_MANIFEST)

.PHONY: build-nkg-image
build-nkg-image: update-nkg-manifest ## Build NKG container and load it and NGINX container on configured kind cluster
	cd .. && make PREFIX=$(NKG_PREFIX) TAG=$(NKG_TAG) container

.PHONY: load-images
load-images: preload-nginx-container ## Load NKG and NGINX containers on configured kind cluster
	kind load docker-image $(NKG_PREFIX):$(NKG_TAG)

.PHONY: prepare-nkg-dependencies
prepare-nkg-dependencies: ## Install NKG dependencies on configured kind cluster
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml
	kubectl wait --for=condition=available --timeout=60s deployment gateway-api-admission-server -n gateway-system 
	kubectl apply -f ../deploy/manifests/namespace.yaml
	kubectl create configmap njs-modules --from-file=../internal/nginx/modules/src/httpmatches.js -n nginx-gateway
	kubectl apply -f ../deploy/manifests/nginx-conf.yaml
	kubectl apply -f ../deploy/manifests/rbac.yaml
	kubectl apply -f ../deploy/manifests/gatewayclass.yaml
	kubectl apply -f ../deploy/manifests/service/nodeport.yaml

.PHONY: deploy-updated-provisioner
deploy-updated-provisioner: ## Update provisioner manifest and deploy to the configured kind cluster
	yq '(select(di != 3))' provisioner/provisioner.yaml | kubectl apply -f -
	yq '(select(.spec.template.spec.containers[].image) | .spec.template.spec.containers[].image="$(NKG_PREFIX):$(NKG_TAG)" | .spec.template.spec.containers[].imagePullPolicy = "Never")' provisioner/provisioner.yaml | kubectl apply -f -

.PHONY: install-nkg-local-build
install-nkg-local-build: build-nkg-image load-images prepare-nkg-dependencies deploy-updated-provisioner ## Install NKG from local build with provisioner on configured kind cluster

.PHONY: install-nkg-local-build
install-nkg-local-no-build: load-images prepare-nkg-dependencies deploy-updated-provisioner ## Install NKG from local build with provisioner on configured kind cluster but do not build the NKG image

.PHONY: install-nkg-edge
install-nkg-edge: preload-nginx-container prepare-nkg-dependencies ## Install NKG with provisioner from edge on configured kind cluster
	kubectl apply -f provisioner/provisioner.yaml

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
uninstall-nkg: ## Uninstall NKG on configured kind cluster
	kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml
	kubectl delete -f ../deploy/manifests/rbac.yaml
	kubectl delete -f ../deploy/manifests/namespace.yaml
	kubectl delete clusterrole nginx-gateway-provisioner
	kubectl delete clusterrolebinding nginx-gateway-provisioner

.PHONY: undo-image-update
undo-image-update: ## Undo the NKG image name and tag in deployment manifest
	git checkout -- $(NKG_DEPLOYMENT_MANIFEST)

.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster
