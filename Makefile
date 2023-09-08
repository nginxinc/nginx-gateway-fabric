# variables that should not be overridden by the user
VERSION = edge
GIT_COMMIT = $(shell git rev-parse HEAD || echo "unknown")
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MANIFEST_DIR = $(shell pwd)/deploy/manifests
CHART_DIR = $(shell pwd)/deploy/helm-chart
NGINX_CONF_DIR = internal/mode/static/nginx/conf
NJS_DIR = internal/mode/static/nginx/modules/src

# go build flags - should not be overridden by the user
GO_LINKER_FlAGS_VARS = -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE}
GO_LINKER_FLAGS_OPTIMIZATIONS = -s -w
GO_LINKER_FLAGS = $(GO_LINKER_FLAGS_OPTIMIZATIONS) $(GO_LINKER_FlAGS_VARS)

# variables that can be overridden by the user
PREFIX ?= nginx-kubernetes-gateway## The name of the NKG image. For example, nginx-kubernetes-gateway
NGINX_PREFIX ?= $(PREFIX)/nginx## The name of the nginx image. For example: nginx-kubernetes-gateway/nginx
TAG ?= $(VERSION:v%=%)## The tag of the image. For example, 0.3.0
TARGET ?= local## The target of the build. Possible values: local and container
KIND_KUBE_CONFIG=$${HOME}/.kube/kind/config## The location of the kind kubeconfig
OUT_DIR ?= $(shell pwd)/build/out## The folder where the binary will be stored
ARCH ?= amd64## The architecture of the image and/or binary. For example: amd64 or arm64
override HELM_TEMPLATE_COMMON_ARGS += --set creator=template --set nameOverride=nginx-gateway## The common options for the Helm template command.
override HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE += --set service.create=false## The options to be passed to the full Helm templating command only.
override NGINX_DOCKER_BUILD_OPTIONS += --build-arg NJS_DIR=$(NJS_DIR) --build-arg NGINX_CONF_DIR=$(NGINX_CONF_DIR)
.DEFAULT_GOAL := help

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m [VARIABLE=value...]\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@grep -E '^(override )?[a-zA-Z_-]+ \??\+?= .*?## .*$$' $< | sort | awk 'BEGIN {FS = " \\??\\+?= .*?## "; printf "\nVariables:\n\n"}; {gsub(/override /, "", $$1); printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build-images
build-images: build-nkg-image build-nginx-image ## Build the NKG and nginx docker images

.PHONY: build-nkg-image
build-nkg-image: check-for-docker build ## Build the NKG docker image
	docker build --platform linux/$(ARCH) --target $(strip $(TARGET)) -f build/Dockerfile -t $(strip $(PREFIX)):$(strip $(TAG)) .

.PHONY: build-nginx-image
build-nginx-image: check-for-docker ## Build the custom nginx image
	docker build --platform linux/$(ARCH) $(strip $(NGINX_DOCKER_BUILD_OPTIONS)) -f build/Dockerfile.nginx -t $(strip $(NGINX_PREFIX)):$(strip $(TAG)) .

.PHONY: check-for-docker
check-for-docker: ## Check if Docker is installed
	@docker -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with Docker\n"; exit $$code)

.PHONY: build
build: ## Build the binary
ifeq (${TARGET},local)
	@go version || (code=$$?; printf "\033[0;31mError\033[0m: unable to build locally\n"; exit $$code)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -trimpath -a -ldflags "$(GO_LINKER_FLAGS)" $(ADDITIONAL_GO_BUILD_FLAGS) -o $(OUT_DIR)/gateway github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway
endif

.PHONY: build-goreleaser
build-goreleaser: ## Build the binary using GoReleaser
	@goreleaser -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with GoReleaser. Follow the docs to install it https://goreleaser.com/install\n"; exit $$code)
	GOOS=linux GOPATH=$(shell go env GOPATH) GOARCH=$(ARCH) goreleaser build --clean --snapshot --single-target

.PHONY: generate
generate: ## Run go generate
	go generate ./...

.PHONY: generate-crds
generate-crds: ## Generate CRDs and Go types using kubebuilder
	go run sigs.k8s.io/controller-tools/cmd/controller-gen crd paths=./apis/... output:crd:dir=deploy/helm-chart/crds
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object paths=./apis/...

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
	docker run --pull always --rm -v $(shell pwd):/nginx-kubernetes-gateway -w /nginx-kubernetes-gateway -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -v $(shell go env GOPATH)/pkg:/go/pkg golangci/golangci-lint:latest golangci-lint --color always run

.PHONY: unit-test
unit-test: ## Run unit tests for the go code
	go test ./... -race -coverprofile cover.out
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

.PHONY: debug-build
debug-build: GO_LINKER_FLAGS=$(GO_LINKER_FlAGS_VARS)
debug-build: ADDITIONAL_GO_BUILD_FLAGS=-gcflags "all=-N -l"
debug-build: build ## Build binary with debug info, symbols, and no optimizations

.PHONY: build-nkg-debug-image
build-nkg-debug-image: debug-build build-nkg-image ## Build NKG image with debug binary

.PHONY: generate-manifests
generate-manifests: ## Generate manifests using Helm.
	cp $(CHART_DIR)/crds/* $(MANIFEST_DIR)/crds/
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) $(HELM_TEMPLATE_EXTRA_ARGS_FOR_ALL_MANIFESTS_FILE) -n nginx-gateway | cat $(strip $(MANIFEST_DIR))/namespace.yaml - > $(strip $(MANIFEST_DIR))/nginx-gateway.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) --set metrics.enable=false -n nginx-gateway -s templates/deployment.yaml > conformance/provisioner/static-deployment.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) -n nginx-gateway -s templates/service.yaml > $(strip $(MANIFEST_DIR))/service/loadbalancer.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) --set service.annotations.'service\.beta\.kubernetes\.io\/aws-load-balancer-type'="nlb" -n nginx-gateway -s templates/service.yaml > $(strip $(MANIFEST_DIR))/service/loadbalancer-aws-nlb.yaml
	helm template nginx-gateway $(CHART_DIR) $(HELM_TEMPLATE_COMMON_ARGS) --set service.type=NodePort --set service.externalTrafficPolicy="" -n nginx-gateway -s templates/service.yaml > $(strip $(MANIFEST_DIR))/service/nodeport.yaml

.PHONY: dev-all
dev-all: deps fmt njs-fmt vet lint unit-test njs-unit-test ## Run all the development checks
