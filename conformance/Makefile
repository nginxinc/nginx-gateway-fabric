NKG_TAG = edge
NKG_PREFIX = nginx-kubernetes-gateway
KIND_KUBE_CONFIG_FOLDER = $${HOME}/.kube/kind
TAG = latest
PREFIX = conformance-test-runner
NGINX_IMAGE=$(shell yq '.spec.template.spec.containers[1].image as $$nginx_ver | $$nginx_ver' ../deploy/manifests/deployment.yaml)
.DEFAULT_GOAL := help

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build-test-runner-image
build-test-runner-image: ## Build conformance test runner image
	docker build -t $(PREFIX):$(TAG) -f tests/Dockerfile ..

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	kind create cluster --image kindest/node:v1.27.1
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG_FOLDER)/config

.PHONY: prepare-nkg
prepare-nkg: ## Build and load NKG and NGINX containers on configured kind cluster
	cd .. && make PREFIX=$(NKG_PREFIX) TAG=$(NKG_TAG) container
	kind load docker-image $(NKG_PREFIX):$(NKG_TAG)
	docker pull $(NGINX_IMAGE)
	kind load docker-image $(NGINX_IMAGE)
	yq -i 'with(.spec.template.spec.containers[0]; .image = "$(NKG_PREFIX):$(NKG_TAG)" | .imagePullPolicy = "Never")' ../deploy/manifests/deployment.yaml

.PHONY: prepare-nkg-no-build
prepare-nkg-no-build: ## Load NKG and NGINX containers on configured kind cluster
	kind load docker-image $(NKG_PREFIX):$(NKG_TAG)
	docker pull $(NGINX_IMAGE)
	kind load docker-image $(NGINX_IMAGE)
	yq -i 'with(.spec.template.spec.containers[0]; .image = "$(NKG_PREFIX):$(NKG_TAG)" | .imagePullPolicy = "Never")' ../deploy/manifests/deployment.yaml

.PHONY: install-nkg
install-nkg: ## Install NKG with provisioner on configured kind cluster
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml
	kubectl wait --for=condition=available --timeout=60s deployment gateway-api-admission-server -n gateway-system 
	kubectl apply -f ../deploy/manifests/namespace.yaml
	kubectl create configmap njs-modules --from-file=../internal/nginx/modules/src/httpmatches.js -n nginx-gateway
	kubectl apply -f ../deploy/manifests/nginx-conf.yaml
	kubectl apply -f ../deploy/manifests/rbac.yaml
	kubectl apply -f ../deploy/manifests/gatewayclass.yaml
	kubectl apply -f ../deploy/manifests/service/nodeport.yaml
	yq '(select(di != 3))' provisioner/provisioner.yaml | kubectl apply -f -
	yq '(select(.spec.template.spec.containers[].image) | .spec.template.spec.containers[].image="$(NKG_PREFIX):$(NKG_TAG)" | .spec.template.spec.containers[].imagePullPolicy = "Never")' provisioner/provisioner.yaml | kubectl apply -f -

.PHONY: run-conformance-tests
run-conformance-tests: ## Run conformance tests
	kind load docker-image $(PREFIX):$(TAG)
	yq '(select(di != 3))' tests/conformance_spec.yaml | kubectl apply -f -
	yq '(select(.spec.containers[].image) | .spec.containers[].image="$(PREFIX):$(TAG)")' tests/conformance_spec.yaml | kubectl apply -f -
	kubectl wait --for=condition=ready pod conformance
	kubectl logs conformance --follow

.PHONY: cleanup-conformance-tests
cleanup-conformance-tests: ## Clean up conformance tests fixtures
	kubectl delete -f tests/conformance_spec.yaml

.PHONY: uninstall-nkg
uninstall-nkg: ## Uninstall NKG on configured kind cluster
	kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml
	kubectl delete -f ../deploy/manifests/rbac.yaml
	kubectl delete -f ../deploy/manifests/namespace.yaml
	kubectl delete clusterrole nginx-gateway-provisioner
	kubectl delete clusterrolebinding nginx-gateway-provisioner

.PHONY: undo-image-update
undo-image-update: ## Undo the NKG image name and tag in deployment manifest
	git checkout -- ../deploy/manifests/deployment.yaml

.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster
