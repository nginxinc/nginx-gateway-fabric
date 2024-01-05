# variables that should not be overridden by the user
VERSION = edge
GIT_COMMIT = $(shell git rev-parse HEAD || echo "unknown")
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MANIFEST_DIR = $(shell pwd)/deploy/manifests
CHART_DIR = $(shell pwd)/deploy/helm-chart
NGINX_CONF_DIR = internal/mode/static/nginx/conf
NJS_DIR = internal/mode/static/nginx/modules/src
NGINX_DOCKER_BUILD_PLUS_ARGS = --secret id=nginx-repo.crt,src=nginx-repo.crt --secret id=nginx-repo.key,src=nginx-repo.key
BUILD_AGENT=local
TELEMETRY_REPORT_PERIOD = 24h
GW_API_VERSION = 1.0.0
INSTALL_WEBHOOK = false

# go build flags - should not be overridden by the user
GO_LINKER_FlAGS_VARS = -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE} -X main.telemetryReportPeriod=${TELEMETRY_REPORT_PERIOD}
GO_LINKER_FLAGS_OPTIMIZATIONS = -s -w
GO_LINKER_FLAGS = $(GO_LINKER_FLAGS_OPTIMIZATIONS) $(GO_LINKER_FlAGS_VARS)

# variables that can be overridden by the user
PREFIX ?= nginx-gateway-fabric## The name of the NGF image. For example, nginx-gateway-fabric
NGINX_PREFIX ?= $(PREFIX)/nginx## The name of the nginx image. For example: nginx-gateway-fabric/nginx
NGINX_PLUS_PREFIX ?= $(PREFIX)/nginxplus## The name of the nginx plus image. For example: nginx-gateway-fabric/nginxplus
TAG ?= $(VERSION:v%=%)## The tag of the image. For example, 0.3.0
TARGET ?= local## The target of the build. Possible values: local and container
KIND_KUBE_CONFIG=$${HOME}/.kube/kind/config## The location of the kind kubeconfig
OUT_DIR ?= $(shell pwd)/build/out## The folder where the binary will be stored
GOARCH ?= amd64## The architecture of the image and/or binary. For example: amd64 or arm64
GOOS ?= linux## The OS of the image and/or binary. For example: linux or darwin
override HELM_TEMPLATE_COMMON_ARGS += --set creator=template --set nameOverride=nginx-gateway## The common options for the Helm template command.
override HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE += --set service.create=false## The options to be passed to the full Helm templating command only.
override NGINX_DOCKER_BUILD_OPTIONS += --build-arg NJS_DIR=$(NJS_DIR) --build-arg NGINX_CONF_DIR=$(NGINX_CONF_DIR) --build-arg BUILD_AGENT=$(BUILD_AGENT)
.DEFAULT_GOAL := help

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m [VARIABLE=value...]\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@grep -E '^(override )?[a-zA-Z_-]+ \??\+?= .*?## .*$$' $< | sort | awk 'BEGIN {FS = " \\??\\+?= .*?## "; printf "\nVariables:\n\n"}; {gsub(/override /, "", $$1); printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build-images
build-images: build-ngf-image build-nginx-image ## Build the NGF and nginx docker images

.PHONY: build-images-with-plus
build-images-with-plus: build-ngf-image build-nginx-plus-image ## Build the NGF and NGINX Plus docker images

.PHONY: build-ngf-image
build-ngf-image: check-for-docker build ## Build the NGF docker image
	docker build --platform linux/$(GOARCH) --target $(strip $(TARGET)) -f build/Dockerfile -t $(strip $(PREFIX)):$(strip $(TAG)) .

.PHONY: build-nginx-image
build-nginx-image: check-for-docker ## Build the custom nginx image
	docker build --platform linux/$(GOARCH) $(strip $(NGINX_DOCKER_BUILD_OPTIONS)) -f build/Dockerfile.nginx -t $(strip $(NGINX_PREFIX)):$(strip $(TAG)) .

.PHONY: build-nginx-plus-image
build-nginx-plus-image: check-for-docker ## Build the custom nginx plus image
	docker build --platform linux/$(GOARCH) $(strip $(NGINX_DOCKER_BUILD_OPTIONS)) $(strip $(NGINX_DOCKER_BUILD_PLUS_ARGS))  -f build/Dockerfile.nginxplus -t $(strip $(NGINX_PLUS_PREFIX)):$(strip $(TAG)) .

.PHONY: check-for-docker
check-for-docker: ## Check if Docker is installed
	@docker -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with Docker\n"; exit $$code)

.PHONY: build
build: ## Build the binary
ifeq (${TARGET},local)
	@go version || (code=$$?; printf "\033[0;31mError\033[0m: unable to build locally\n"; exit $$code)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -a -ldflags "$(GO_LINKER_FLAGS)" $(ADDITIONAL_GO_BUILD_FLAGS) -o $(OUT_DIR)/gateway github.com/nginxinc/nginx-gateway-fabric/cmd/gateway
endif

.PHONY: build-goreleaser
build-goreleaser: ## Build the binary using GoReleaser
	@goreleaser -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with GoReleaser. Follow the docs to install it https://goreleaser.com/install\n"; exit $$code)
	GOOS=linux GOPATH=$(shell go env GOPATH) GOARCH=$(GOARCH) goreleaser build --clean --snapshot --single-target

.PHONY: generate
generate: ## Run go generate
	go generate ./...

.PHONY: generate-crds
generate-crds: ## Generate CRDs and Go types using kubebuilder
	go run sigs.k8s.io/controller-tools/cmd/controller-gen crd paths=./apis/... output:crd:dir=deploy/helm-chart/crds
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object paths=./apis/...

.PHONY: generate-manifests
generate-manifests: ## Generate manifests using Helm.
	cp $(CHART_DIR)/crds/* $(MANIFEST_DIR)/crds/
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) $(HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE) -n nginx-gateway | cat $(strip $(MANIFEST_DIR))/namespace.yaml - > $(strip $(MANIFEST_DIR))/nginx-gateway.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) --set metrics.enable=false -n nginx-gateway -s templates/deployment.yaml > conformance/provisioner/static-deployment.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) -n nginx-gateway -s templates/service.yaml > $(strip $(MANIFEST_DIR))/service/loadbalancer.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) --set service.annotations.'service\.beta\.kubernetes\.io\/aws-load-balancer-type'="nlb" -n nginx-gateway -s templates/service.yaml > $(strip $(MANIFEST_DIR))/service/loadbalancer-aws-nlb.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) --set service.type=NodePort --set service.externalTrafficPolicy="" -n nginx-gateway -s templates/service.yaml > $(strip $(MANIFEST_DIR))/service/nodeport.yaml

.PHONY: crds-release-file
crds-release-file: ## Generate combined crds file for releases
	scripts/combine-crds.sh

.PHONY: clean
clean: ## Clean the build
	-rm -r $(OUT_DIR)

.PHONY: clean-go-cache
clean-go-cache: ## Clean go cache
	@go clean -modcache

.PHONY: deps
deps: ## Add missing and remove unused modules, verify deps and download them to local cache
	@go mod tidy && go mod verify && go mod download

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	$(eval KIND_IMAGE=$(shell grep -m1 'FROM kindest/node' <conformance/tests/Dockerfile | awk -F'[ ]' '{print $$2}'))
	kind create cluster --image $(KIND_IMAGE)
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG)

.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: njs-fmt
njs-fmt: ## Run prettier against the njs httpmatches module
	docker run --rm -w /modules \
		-v $(PWD)/internal/nginx/modules/:/modules/ \
		node:18 \
		/bin/bash -c "npm install && npm run format"

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint against code
	docker run --pull always --rm -v $(shell pwd):/nginx-gateway-fabric -w /nginx-gateway-fabric -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -v $(shell go env GOPATH)/pkg:/go/pkg golangci/golangci-lint:latest golangci-lint --color always run

.PHONY: unit-test
unit-test: ## Run unit tests for the go code
	go test ./internal/... -race -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html

.PHONY: njs-unit-test
njs-unit-test: ## Run unit tests for the njs httpmatches module
	docker run --rm -w /modules \
		-v $(PWD)/internal/mode/static/nginx/modules:/modules/ \
		node:18 \
		/bin/bash -c "npm install && npm test && npm run clean"

.PHONY: lint-helm
lint-helm: ## Run the helm chart linter
	helm lint $(CHART_DIR)

.PHONY: load-images
load-images: ## Load NGF and NGINX images on configured kind cluster.
	kind load docker-image $(PREFIX):$(TAG) $(NGINX_PREFIX):$(TAG)

.PHONY: load-images-with-plus
load-images-with-plus: ## Load NGF and NGINX Plus images on configured kind cluster.
	kind load docker-image $(PREFIX):$(TAG) $(NGINX_PLUS_PREFIX):$(TAG)

.PHONY: install-ngf-local-build
install-ngf-local-build: build-images load-images helm-install-local ## Install NGF from local build on configured kind cluster.

.PHONY: install-ngf-local-build-with-plus
install-ngf-local-build-with-plus: build-images-with-plus load-images-with-plus helm-install-local-with-plus ## Install NGF with NGINX Plus from local build on configured kind cluster.

.PHONY: helm-install-local
helm-install-local: ## Helm install NGF on configured kind cluster with local images. To build, load, and install with helm run make install-ngf-local-build.
	./conformance/scripts/install-gateway.sh $(GW_API_VERSION) $(INSTALL_WEBHOOK)
	helm install dev ./deploy/helm-chart --create-namespace --wait --set service.type=NodePort --set nginxGateway.image.repository=$(PREFIX) --set nginxGateway.image.tag=$(TAG) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=$(NGINX_PREFIX) --set nginx.image.tag=$(TAG) --set nginx.image.pullPolicy=Never -n nginx-gateway

.PHONY: helm-install-local-with-plus
helm-install-local-with-plus: ## Helm install NGF with NGINX Plus on configured kind cluster with local images. To build, load, and install with helm run make install-ngf-local-build-with-plus.
	./conformance/scripts/install-gateway.sh $(GW_API_VERSION) $(INSTALL_WEBHOOK)
	helm install dev ./deploy/helm-chart --create-namespace --wait --set service.type=NodePort --set nginxGateway.image.repository=$(PREFIX) --set nginxGateway.image.tag=$(TAG) --set nginxGateway.image.pullPolicy=Never --set nginx.image.repository=$(NGINX_PLUS_PREFIX) --set nginx.image.tag=$(TAG) --set nginx.image.pullPolicy=Never --set nginx.plus=true -n nginx-gateway

# Debug Targets
.PHONY: debug-build
debug-build: GO_LINKER_FLAGS=$(GO_LINKER_FlAGS_VARS)
debug-build: ADDITIONAL_GO_BUILD_FLAGS=-gcflags "all=-N -l"
debug-build: build ## Build binary with debug info, symbols, and no optimizations

.PHONY: debug-build-dlv-image
debug-build-dlv-image: check-for-docker ## Build the dlv debugger image.
	docker build --platform linux/$(GOARCH) -f debug/Dockerfile -t dlv-debug:edge .

.PHONY: debug-build-images
debug-build-images: debug-build build-ngf-image build-nginx-image debug-build-dlv-image ## Build all images used in debugging.

.PHONY: debug-build-images-with-plus
debug-build-images-with-plus: debug-build build-ngf-image build-nginx-plus-image debug-build-dlv-image ## Build all images with NGINX plus used in debugging.

.PHONY: debug-load-images
debug-load-images: load-images ## Load all images used in debugging to kind cluster.
	kind load docker-image dlv-debug:edge

.PHONY: debug-load-images-with-plus
debug-load-images-with-plus: load-images-with-plus ## Load all images with NGINX Plus used in debugging to kind cluster.
	kind load docker-image dlv-debug:edge

.PHONY: debug-install-local-build
debug-install-local-build: debug-build-images debug-load-images helm-install-local ## Install NGF from local build using debug NGF binary on configured kind cluster.

.PHONY: debug-install-local-build-with-plus
debug-install-local-build-with-plus: debug-build-images-with-plus debug-load-images-with-plus helm-install-local-with-plus ## Install NGF with NGINX Plus from local build using debug NGF binary on configured kind cluster.

.PHONY: dev-all
dev-all: deps fmt njs-fmt vet lint unit-test njs-unit-test ## Run all the development checks
