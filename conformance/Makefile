NKG_TAG = edge
NKG_PREFIX = nginx-kubernetes-gateway
GATEWAY_CLASS = nginx
SUPPORTED_FEATURES = Gateway,HTTPRoute
KIND_KUBE_CONFIG_FOLDER = $${HOME}/.kube/kind
TAG = latest
PREFIX = conformance-test-runner
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
prepare-nkg: ## Build and load NKG container on configured kind cluster
	cd .. && make PREFIX=$(NKG_PREFIX) TAG=$(NKG_TAG) container
	kind load docker-image $(NKG_PREFIX):$(NKG_TAG)

.PHONY: install-nkg
install-nkg: ## Install NKG with provisioner on configured kind cluster
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.0/standard-install.yaml
	kubectl wait --for=condition=available --timeout=60s deployment gateway-api-admission-server -n gateway-system 
	kubectl apply -f ../deploy/manifests/namespace.yaml
	kubectl create configmap njs-modules --from-file=../internal/nginx/modules/src/httpmatches.js -n nginx-gateway
	kubectl apply -f ../deploy/manifests/nginx-conf.yaml
	kubectl apply -f ../deploy/manifests/rbac.yaml
	kubectl apply -f ../deploy/manifests/gatewayclass.yaml
	kubectl apply -f ../deploy/manifests/service/nodeport.yaml
	kubectl apply -f provisioner/provisioner.yaml

.PHONY: update-test-kind-config 
update-test-kind-config: ## Update kind config
	sed -ir "s|server:.*|server: https://kind-control-plane:6443|" $(KIND_KUBE_CONFIG_FOLDER)/config

.PHONY: run-conformance-tests
run-conformance-tests: update-test-kind-config ## Run conformance tests
	docker run --network=kind --rm -v $(KIND_KUBE_CONFIG_FOLDER):/root/.kube $(PREFIX):$(TAG) \
		go test -v . -tags conformance -args --gateway-class=$(GATEWAY_CLASS) --supported-features=$(SUPPORTED_FEATURES)

.PHONY: uninstall-nkg
uninstall-nkg: ## Uninstall NKG on configured kind cluster
	kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.0/standard-install.yaml
	kubectl delete -f ../deploy/manifests/rbac.yaml
	kubectl delete -f ../deploy/manifests/namespace.yaml
	kubectl delete clusterrole nginx-gateway-provisioner
	kubectl delete clusterrolebinding nginx-gateway-provisioner
	
.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster
